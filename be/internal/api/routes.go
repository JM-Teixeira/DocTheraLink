package api

import (
	"log"

	"klyniva/internal/api/handlers"
	"klyniva/internal/auth"
	"klyniva/internal/config"
	"klyniva/internal/database"
	"klyniva/internal/mailer"
	"klyniva/internal/models"
	"klyniva/internal/modules/attachment"
	"klyniva/internal/modules/notification"
	"klyniva/internal/modules/patient"
	"klyniva/internal/modules/referral"
	"klyniva/internal/storage"

	"github.com/gin-gonic/gin"
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
	timetableHandler := handlers.NewTimetableHandler(cfg)

	// Public routes (no authentication required)
	public := router.Group("")
	{
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
	users.GET("/getcurrentuser/", userHandler.GetCurrentUser)
	users.GET("/:username/", userHandler.GetUser)
	users.PUT("/:username/", userHandler.UpdateUser)
	users.PATCH("/:username/", userHandler.PartialUpdateUser)
	users.DELETE("/:username/", userHandler.DeleteUser)
	users.GET("/:username/check_availability/", userHandler.CheckUsername)
	users.GET("/:username/skill/", userHandler.ListUserSkills)
	users.POST("/:username/skill/", userHandler.AddUserSkill)
	users.DELETE("/:username/skill/:id/", userHandler.DeleteUserSkill)
	users.GET("/:username/get_facilities/", userHandler.ListUserFacilities)
	users.PUT("/:username/add_facility/", userHandler.AddUserFacility)
	users.DELETE("/:username/delete_facility/", userHandler.DeleteUserFacility)
	users.DELETE("/:username/clear_home_facility/", userHandler.ClearHomeFacility)
	users.PATCH("/:username/profile_picture/", userHandler.UpdateProfilePicture)
	users.DELETE("/:username/profile_picture/", userHandler.DeleteProfilePicture)
	users.GET("/:username/pnconfig/", userHandler.GetPNConfig)
	users.PATCH("/:username/pnconfig/", userHandler.UpdatePNConfig)

	protected.PUT("/password_change/", authHandler.UpdatePassword)
	protected.GET("/skill/", userHandler.ListAllSkills)

	// Facilities (a facility = a clinic; see CONTEXT.md "Clinics")
	facilities := protected.Group("/facility")
	facilities.GET("/", facilityHandler.ListFacilities)
	facilities.POST("/", facilityHandler.CreateFacility)
	facilities.GET("/:id/", facilityHandler.GetFacility)
	facilities.PUT("/:id/", facilityHandler.UpdateFacility)
	facilities.PATCH("/:id/", facilityHandler.PartialUpdateFacility)
	facilities.DELETE("/:id/", facilityHandler.DeleteFacility)
	facilities.GET("/:id/users/", facilityHandler.GetFacilityUsers)
	// Inventory: still stubs (501), to be rebuilt as a module (TODO M5)
	facilities.GET("/:id/inventory/", facilityHandler.GetInventoryLog)
	facilities.POST("/:id/inventory/", facilityHandler.CreateInventory)
	facilities.GET("/:id/min_quantity/", facilityHandler.GetMinQuantity)
	facilities.POST("/:id/min_quantity/", facilityHandler.SetMinQuantity)
	facilities.GET("/:id/min_quantity/:inventory_id/", facilityHandler.GetMinQuantityItem)
	facilities.PATCH("/:id/min_quantity/:inventory_id/", facilityHandler.UpdateMinQuantity)
	facilities.GET("/:id/inventorysummary/", facilityHandler.GetInventorySummary)
	facilities.PUT("/:id/inventory/:inventory_id/flag/", facilityHandler.FlagInventoryItem)
	facilities.DELETE("/:id/inventory/delete_last/", facilityHandler.DeleteLastInventoryLog)
	// Timetable sections
	facilities.GET("/:id/sections/", timetableHandler.ListSections)
	facilities.POST("/:id/sections/", timetableHandler.CreateSection)

	protected.GET("/getallfacilities/", facilityHandler.ListAllFacilities)
	protected.GET("/getallfacilities/:id/", facilityHandler.GetAnyFacility)
	protected.GET("/items/", facilityHandler.GetItems)

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

	// Patients (module)
	patient.RegisterRoutes(protected.Group("/patient"), db)

	// In-app + e-mail notifications (module); referrals publish events to it
	notifications := notification.NewService(notification.NewGormRepository(db),
		mailer.New(cfg.Email), cfg.App.PublicURL, nil)
	notification.RegisterRoutes(protected, notifications)

	// Doctor → physiotherapist referrals (module, core Klyniva workflow)
	referrals := referral.RegisterRoutes(protected, db, notifications)

	// Files attached to referrals (access follows the referral)
	uploads, err := storage.NewLocal(cfg.Upload.Path)
	if err != nil {
		log.Fatalf("attachments: %v", err)
	}
	attachment.RegisterRoutes(protected, attachment.NewService(
		attachment.NewGormRepository(db), uploads, referrals, cfg.Upload.MaxSize))

	// Staff time log (clock in/out)
	timelog := protected.Group("/timelog")
	timelog.POST("/clock-in", timeLogHandler.ClockIn)
	timelog.POST("/clock-out", timeLogHandler.ClockOut)
	timelog.GET("/my-logs", timeLogHandler.GetMyLogs)
	timelog.GET("/status", timeLogHandler.GetCurrentStatus)
	timelog.GET("/user/:id", timeLogHandler.GetUserLogs)

	// Timetable
	timetable := protected.Group("/timetable")
	timetable.GET("/:id", timetableHandler.GetTimetable)
	timetable.POST("/:id/slot", timetableHandler.CreateSlot)
	timetable.PUT("/slot/:id", timetableHandler.UpdateSlot)
	timetable.DELETE("/slot/:id", timetableHandler.DeleteSlot)

	sections := protected.Group("/sections")
	sections.GET("/:id", timetableHandler.GetSection)
	sections.PUT("/:id", timetableHandler.UpdateSection)
	sections.DELETE("/:id", timetableHandler.DeleteSection)

	// Admin-only CSV exports
	admin := protected.Group("")
	admin.Use(auth.RequireRole(models.RoleStateAdmin, models.RoleDistrictAdmin, models.RoleLocalBodyAdmin))
	admin.GET("/download/facility/", facilityHandler.DownloadFacility)
	admin.GET("/download/patient/", patientHandler.DownloadPatients)
}
