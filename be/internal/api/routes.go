package api

import (
	"errors"
	"log"
	"net/http"

	"klyniva/internal/api/handlers"
	"klyniva/internal/auth"
	"klyniva/internal/config"
	"klyniva/internal/database"
	"klyniva/internal/mailer"
	"klyniva/internal/models"
	"klyniva/internal/modules/attachment"
	"klyniva/internal/modules/catalog"
	"klyniva/internal/modules/feedback"
	"klyniva/internal/modules/inventory"
	"klyniva/internal/modules/notification"
	"klyniva/internal/modules/patient"
	"klyniva/internal/modules/referral"
	"klyniva/internal/modules/schedule"
	"klyniva/internal/modules/staff"
	"klyniva/internal/modules/timeline"
	"klyniva/internal/storage"
	"klyniva/internal/version"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all API routes under /api/v1.
// New features register themselves from their module (internal/modules/*).
func SetupRoutes(router *gin.RouterGroup, jwtManager *auth.JWTManager, cfg *config.Config) {
	// Legacy handlers (inherited; being replaced module by module)
	authHandler := handlers.NewAuthHandler(jwtManager, cfg)
	userHandler := handlers.NewUserHandler(cfg)
	facilityHandler := handlers.NewFacilityHandler(cfg)
	patientHandler := handlers.NewPatientHandler(cfg)
	timeLogHandler := handlers.NewTimeLogHandler(cfg)

	// Public routes (no authentication required)
	public := router.Group("")
	{
		// Release number (shown in the app next to the frontend version)
		public.GET("/version/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"version": version.Version})
		})
		authRoutes := public.Group("/auth")
		authRoutes.POST("/login/", authHandler.Login)
		authRoutes.POST("/token/refresh/", authHandler.RefreshToken)
		authRoutes.POST("/token/verify/", authHandler.VerifyToken)

		public.POST("/password_reset/", authHandler.ForgotPassword)
		public.POST("/password_reset/check/", authHandler.CheckResetToken)
		public.POST("/password_reset/confirm/", authHandler.ResetPassword)
	}

	// Protected routes (authentication required)
	protected := router.Group("")
	protected.Use(auth.AuthMiddleware(jwtManager))

	// Users
	users := protected.Group("/users")
	users.GET("/", userHandler.ListUsers)
	users.POST("/", userHandler.CreateUser)
	users.POST("/add_user/", userHandler.CreateUser) // path used by the frontend form
	users.GET("/getcurrentuser/", userHandler.GetCurrentUser)
	users.GET("/username_suggestion/", userHandler.UsernameSuggestion)
	users.GET("/:username/", userHandler.GetUser)
	users.PUT("/:username/", userHandler.UpdateUser)
	users.PATCH("/:username/", userHandler.PartialUpdateUser)
	users.DELETE("/:username/", userHandler.DeleteUser)
	users.GET("/:username/check_availability/", userHandler.CheckUsername)
	users.PATCH("/:username/profile_picture/", userHandler.UpdateProfilePicture)
	users.DELETE("/:username/profile_picture/", userHandler.DeleteProfilePicture)
	users.GET("/:username/pnconfig/", userHandler.GetPNConfig)
	users.PATCH("/:username/pnconfig/", userHandler.UpdatePNConfig)

	protected.PUT("/password_change/", authHandler.UpdatePassword)

	// Facilities (a facility = a clinic; see CONTEXT.md "Clinics")
	facilities := protected.Group("/facility")
	facilities.GET("/", facilityHandler.ListFacilities)
	facilities.POST("/", facilityHandler.CreateFacility)
	facilities.GET("/:id/", facilityHandler.GetFacility)
	facilities.PUT("/:id/", facilityHandler.UpdateFacility)
	facilities.PATCH("/:id/", facilityHandler.PartialUpdateFacility)
	facilities.DELETE("/:id/", facilityHandler.DeleteFacility)
	facilities.GET("/:id/users/", facilityHandler.GetFacilityUsers)
	facilities.GET("/:id/get_users/", facilityHandler.GetFacilityUsers) // path used by the frontend

	protected.GET("/getallfacilities/", facilityHandler.ListAllFacilities)
	protected.GET("/getallfacilities/:id/", facilityHandler.GetAnyFacility)

	// Geographic reference data (Portuguese districts/municipalities)
	protected.GET("/state/", facilityHandler.ListStates)
	protected.GET("/state/:id/", facilityHandler.GetState)
	protected.GET("/state/:id/districts/", facilityHandler.GetDistrictsByState)
	protected.GET("/district/", facilityHandler.GetDistrictByName)
	protected.GET("/district/:id/", facilityHandler.GetDistrict)
	protected.GET("/district/:id/get_all_local_body/", facilityHandler.GetAllLocalBodyByDistrict)
	protected.GET("/district/:id/local_bodies/", facilityHandler.GetLocalBodyByDistrict)
	protected.GET("/local_body/", facilityHandler.GetAllLocalBody)
	protected.GET("/local_body/:id/", facilityHandler.GetLocalBody)
	protected.GET("/ward/", facilityHandler.GetWards)
	protected.GET("/ward/:id/", facilityHandler.GetWard)

	db := database.GetDB()

	// Every request: the account is still active, and has its own password
	auth.AccountChecker = func(id uint) (auth.AccountState, error) {
		var u models.User
		err := db.Select("is_active", "must_change_password").Take(&u, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.AccountState{}, nil // deleted account
		}
		return auth.AccountState{Active: u.IsActive, MustChangePassword: u.MustChangePassword}, err
	}

	// Patients (module)
	patients := patient.RegisterRoutes(protected.Group("/patient"), db)

	// In-app + e-mail notifications (module); referrals publish events to it
	notifications := notification.NewService(notification.NewGormRepository(db),
		mailer.New(cfg.Email), cfg.App.PublicURL, nil)
	notification.RegisterRoutes(protected, notifications)

	// Doctor → physiotherapist referrals (module, core Klyniva workflow)
	referrals := referral.RegisterRoutes(protected, db, notifications)

	// Reference data for the referral form (ICD-10 subset, templates)
	catalog.RegisterRoutes(protected, catalog.MustLoad())

	// Files attached to referrals (access follows the referral)
	uploads, err := storage.NewLocal(cfg.Upload.Path)
	if err != nil {
		log.Fatalf("attachments: %v", err)
	}
	files := attachment.NewService(attachment.NewGormRepository(db), uploads, referrals, cfg.Upload.MaxSize)
	attachment.RegisterRoutes(protected, files)

	// Staff time log (clock in/out)
	timelog := protected.Group("/timelog")
	timelog.POST("/clock-in", timeLogHandler.ClockIn)
	timelog.POST("/clock-out", timeLogHandler.ClockOut)
	timelog.GET("/my-logs", timeLogHandler.GetMyLogs)
	timelog.GET("/status", timeLogHandler.GetCurrentStatus)
	timelog.GET("/user/:id", timeLogHandler.GetUserLogs)

	// Clinic consumables (stock, use, restock, low-stock alerts)
	stock := inventory.RegisterRoutes(protected, db)

	// Time board: divisions, physio shifts, patients per physio per hour
	board := schedule.RegisterRoutes(protected, db, stock) // attended sessions use consumables
	referrals.WithCovers(board)                            // cover physios see the referrals they cover

	// Patient timeline: referrals, messages, files and sessions in one list
	timeline.RegisterRoutes(protected, timeline.Services{
		Patients: patients, Referrals: referrals, Files: files, Board: board,
	})

	// Suggestions about the app (everyone sends, admins read)
	feedback.RegisterRoutes(protected, db)

	// Skills, clinic and employment status of staff
	staff.RegisterRoutes(protected, staff.NewService(staff.NewGormRepository(db)))

	// Admin-only CSV exports
	admin := protected.Group("")
	admin.Use(auth.RequireRole(models.RoleStateAdmin, models.RoleDistrictAdmin, models.RoleLocalBodyAdmin))
	admin.GET("/download/facility/", facilityHandler.DownloadFacility)
	admin.GET("/download/patient/", patientHandler.DownloadPatients)
}
