package api

import (
	"care-backend/internal/api/handlers"
	"care-backend/internal/auth"
	"care-backend/internal/config"
	"care-backend/internal/database"
	"care-backend/internal/models"
	"care-backend/internal/modules/patient"
	"care-backend/internal/modules/referral"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.RouterGroup, jwtManager *auth.JWTManager, cfg *config.Config) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(jwtManager, cfg)
	userHandler := handlers.NewUserHandler(cfg)
	facilityHandler := handlers.NewFacilityHandler(cfg)
	patientHandler := handlers.NewPatientHandler(cfg)
	consultationHandler := handlers.NewConsultationHandler(cfg)
	assetHandler := handlers.NewAssetHandler(cfg)
	notificationHandler := handlers.NewNotificationHandler(cfg)
	timeLogHandler := handlers.NewTimeLogHandler(cfg)
	timetableHandler := handlers.NewTimetableHandler(cfg)
	// therapyHandler := handlers.NewTherapyHandler(cfg) // Temporarily disabled

	// Public routes (no authentication required)
	public := router.Group("")
	{
		// Authentication routes
		auth := public.Group("/auth")
		{
			auth.POST("/login/", authHandler.Login)
			auth.POST("/token/refresh/", authHandler.RefreshToken)
			auth.POST("/token/verify/", authHandler.VerifyToken)
		}

		// Password reset routes
		public.POST("/password_reset/", authHandler.ForgotPassword)
		public.POST("/password_reset/check/", authHandler.CheckResetToken)
		public.POST("/password_reset/confirm/", authHandler.ResetPassword)

		// Public asset QR code
		public.GET("/public/asset_qr/:qr_code_id/", assetHandler.GetAssetByQR)
	}

	// Protected routes (authentication required)
	protected := router.Group("")
	protected.Use(auth.AuthMiddleware(jwtManager))
	{
		// User routes
		users := protected.Group("/users")
		{
			users.GET("/", userHandler.ListUsers)
			users.POST("/", userHandler.CreateUser)
			users.GET("/getcurrentuser/", userHandler.GetCurrentUser)
			users.GET("/:username/", userHandler.GetUser)
			users.PUT("/:username/", userHandler.UpdateUser)
			users.PATCH("/:username/", userHandler.PartialUpdateUser)
			users.DELETE("/:username/", userHandler.DeleteUser)
			users.GET("/:username/check_availability/", userHandler.CheckUsername)

			// User skills
			users.GET("/:username/skill/", userHandler.ListUserSkills)
			users.POST("/:username/skill/", userHandler.AddUserSkill)
			users.DELETE("/:username/skill/:id/", userHandler.DeleteUserSkill)

			// User facilities
			users.GET("/:username/get_facilities/", userHandler.ListUserFacilities)
			users.PUT("/:username/add_facility/", userHandler.AddUserFacility)
			users.DELETE("/:username/delete_facility/", userHandler.DeleteUserFacility)
			users.DELETE("/:username/clear_home_facility/", userHandler.ClearHomeFacility)

			// Profile picture
			users.PATCH("/:username/profile_picture/", userHandler.UpdateProfilePicture)
			users.DELETE("/:username/profile_picture/", userHandler.DeleteProfilePicture)

			// Push notifications
			users.GET("/:username/pnconfig/", userHandler.GetPNConfig)
			users.PATCH("/:username/pnconfig/", userHandler.UpdatePNConfig)
		}

		// Password change
		protected.PUT("/password_change/", authHandler.UpdatePassword)

		// Skills
		protected.GET("/skill/", userHandler.ListAllSkills)

		// Facilities
		facilities := protected.Group("/facility")
		{
			facilities.GET("/", facilityHandler.ListFacilities)
			facilities.POST("/", facilityHandler.CreateFacility)
			facilities.GET("/:id/", facilityHandler.GetFacility)
			facilities.PUT("/:id/", facilityHandler.UpdateFacility)
			facilities.PATCH("/:id/", facilityHandler.PartialUpdateFacility)
			facilities.DELETE("/:id/", facilityHandler.DeleteFacility)

			// Facility users
			facilities.GET("/:id/users/", facilityHandler.GetFacilityUsers)

			// Facility spokes
			facilities.GET("/:id/hubs/", facilityHandler.GetFacilityHubs)
			facilities.GET("/:id/spokes/", facilityHandler.GetFacilitySpokes)
			facilities.POST("/:id/spokes/", facilityHandler.CreateFacilitySpoke)
			facilities.GET("/:id/spokes/:spoke_id/", facilityHandler.GetFacilitySpoke)
			facilities.PATCH("/:id/spokes/:spoke_id/", facilityHandler.UpdateFacilitySpoke)
			facilities.DELETE("/:id/spokes/:spoke_id/", facilityHandler.DeleteFacilitySpoke)

			// Asset locations
			facilities.GET("/:id/asset_location/", facilityHandler.ListAssetLocations)
			facilities.POST("/:id/asset_location/", facilityHandler.CreateAssetLocation)
			facilities.GET("/:id/asset_location/:location_id/", facilityHandler.GetAssetLocation)
			facilities.PUT("/:id/asset_location/:location_id/", facilityHandler.UpdateAssetLocation)
			facilities.PATCH("/:id/asset_location/:location_id/", facilityHandler.PartialUpdateAssetLocation)
			facilities.DELETE("/:id/asset_location/:location_id/", facilityHandler.DeleteAssetLocation)

			// Inventory
			facilities.GET("/:id/inventory/", facilityHandler.GetInventoryLog)
			facilities.POST("/:id/inventory/", facilityHandler.CreateInventory)
			facilities.GET("/:id/min_quantity/", facilityHandler.GetMinQuantity)
			facilities.POST("/:id/min_quantity/", facilityHandler.SetMinQuantity)
			facilities.GET("/:id/min_quantity/:inventory_id/", facilityHandler.GetMinQuantityItem)
			facilities.PATCH("/:id/min_quantity/:inventory_id/", facilityHandler.UpdateMinQuantity)
			facilities.GET("/:id/inventorysummary/", facilityHandler.GetInventorySummary)
			facilities.PUT("/:id/inventory/:inventory_id/flag/", facilityHandler.FlagInventoryItem)
			facilities.DELETE("/:id/inventory/delete_last/", facilityHandler.DeleteLastInventoryLog)

			// Discharged patients
			facilities.GET("/:id/discharged_patients/", facilityHandler.ListDischargedPatients)

			// Patient asset beds
			facilities.GET("/:id/patient_asset_beds/", facilityHandler.ListPatientAssetBeds)

			// Facility sections for timetable
			facilities.GET("/:id/sections/", timetableHandler.ListSections)
			facilities.POST("/:id/sections/", timetableHandler.CreateSection)
		}

		// All facilities (public access)
		protected.GET("/getallfacilities/", facilityHandler.ListAllFacilities)
		protected.GET("/getallfacilities/:id/", facilityHandler.GetAnyFacility)

		// Patients
		patients := protected.Group("/patient")
		{
			// List/create/search/get/update live in the patient module
			patient.RegisterRoutes(patients, database.GetDB())
			patients.POST("/:id/transfer/", patientHandler.TransferPatient)

			// Patient notes - TEMPORARILY DISABLED due to model conflicts
			// patients.GET("/:patientId/notes/", patientHandler.GetPatientNotes)
			// patients.POST("/:patientId/notes/", patientHandler.AddPatientNote)
			// patients.PUT("/:patientId/notes/:noteId/", patientHandler.UpdatePatientNote)
			// patients.GET("/:patientId/notes/:noteId/edits/", patientHandler.GetPatientNoteEditHistory)

			// Patient investigations
			patients.GET("/:id/investigation/", patientHandler.GetPatientInvestigation)
		}

		// Consultations
		consultations := protected.Group("/consultation")
		{
			consultations.GET("/", consultationHandler.ListConsultations)
			consultations.POST("/", consultationHandler.CreateConsultation)
			consultations.GET("/:id/", consultationHandler.GetConsultation)
			consultations.PUT("/:id/", consultationHandler.UpdateConsultation)
			consultations.PATCH("/:id/", consultationHandler.PartialUpdateConsultation)
			consultations.DELETE("/:id/", consultationHandler.DeleteConsultation)

			// Daily rounds
			consultations.GET("/:id/daily_rounds/", consultationHandler.GetDailyReports)
			consultations.POST("/:id/daily_rounds/", consultationHandler.CreateDailyRounds)
			consultations.GET("/:id/daily_rounds/:round_id/", consultationHandler.GetDailyReport)
			consultations.PUT("/:id/daily_rounds/:round_id/", consultationHandler.UpdateDailyReport)
			consultations.PATCH("/:id/daily_rounds/:round_id/", consultationHandler.UpdateDailyRound)
			consultations.POST("/:id/daily_rounds/analyse/", consultationHandler.AnalyseDailyRounds)

			// Events
			consultations.GET("/:id/events/", consultationHandler.GetEvents)

			// Discharge
			consultations.POST("/:id/discharge_patient/", consultationHandler.DischargePatient)
			consultations.POST("/:id/generate_discharge_summary/", consultationHandler.GenerateDischargeSummary)
			consultations.GET("/:id/preview_discharge_summary/", consultationHandler.PreviewDischargeSummary)
			consultations.POST("/:id/email_discharge_summary/", consultationHandler.EmailDischargeSummary)

			// Consents
			consultations.GET("/:id/consents/", consultationHandler.ListConsents)
			consultations.POST("/:id/consents/", consultationHandler.CreateConsent)
			consultations.GET("/:id/consents/:consent_id/", consultationHandler.GetConsent)
			consultations.PATCH("/:id/consents/:consent_id/", consultationHandler.PartialUpdateConsent)

			// Investigations
			consultations.GET("/:id/investigation/", consultationHandler.GetInvestigation)
			consultations.POST("/:id/investigation/", consultationHandler.CreateInvestigation)
			consultations.GET("/:id/investigation/get_sessions/", consultationHandler.GetInvestigationSessions)
			consultations.PUT("/:id/investigation/batchUpdate/", consultationHandler.EditInvestigation)
		}

		// Assets
		assets := protected.Group("/asset")
		{
			assets.GET("/", assetHandler.ListAssets)
			assets.POST("/", assetHandler.CreateAsset)
			assets.GET("/:external_id/", assetHandler.GetAsset)
			assets.PUT("/:external_id/", assetHandler.UpdateAsset)
			assets.PATCH("/:external_id/", assetHandler.PartialUpdateAsset)
			assets.DELETE("/:external_id/", assetHandler.DeleteAsset)
			assets.GET("/:external_id/availability/", assetHandler.ListAssetAvailability)
			assets.POST("/:external_id/operate_assets/", assetHandler.OperateAsset)

			// Asset service records
			assets.GET("/:external_id/service_records/", assetHandler.ListAssetService)
			assets.GET("/:external_id/service_records/:service_id/", assetHandler.GetAssetService)
			assets.PUT("/:external_id/service_records/:service_id/", assetHandler.UpdateAssetService)

			// User location
			assets.GET("/get_default_user_location/", assetHandler.GetAssetUserLocation)
			assets.POST("/set_default_user_location/", assetHandler.CreateAssetUserLocation)
		}

		// Asset beds
		assetBeds := protected.Group("/assetbed")
		{
			assetBeds.GET("/", assetHandler.ListAssetBeds)
			assetBeds.POST("/", assetHandler.CreateAssetBed)
			assetBeds.GET("/:external_id/", assetHandler.GetAssetBed)
			assetBeds.PUT("/:external_id/", assetHandler.UpdateAssetBed)
			assetBeds.PATCH("/:external_id/", assetHandler.PartialUpdateAssetBed)
			assetBeds.DELETE("/:external_id/", assetHandler.DeleteAssetBed)
		}

		// Asset transactions
		protected.GET("/asset_transaction/", assetHandler.ListAssetTransaction)
		protected.GET("/asset_transaction/:id/", assetHandler.GetAssetTransaction)

		// Beds
		beds := protected.Group("/bed")
		{
			beds.GET("/", facilityHandler.ListFacilityBeds)
			beds.POST("/", facilityHandler.CreateFacilityBed)
			beds.GET("/:external_id/", facilityHandler.GetFacilityBed)
			beds.PUT("/:external_id/", facilityHandler.UpdateFacilityBed)
			beds.DELETE("/:external_id/", facilityHandler.DeleteFacilityBed)
		}

		// Consultation beds
		consultationBeds := protected.Group("/consultationbed")
		{
			consultationBeds.GET("/", consultationHandler.ListConsultationBeds)
			consultationBeds.POST("/", consultationHandler.CreateConsultationBed)
			consultationBeds.GET("/:external_id/", consultationHandler.GetConsultationBed)
			consultationBeds.PUT("/:external_id/", consultationHandler.UpdateConsultationBed)
		}

		// Geographic data
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

		// Inventory items
		protected.GET("/items/", facilityHandler.GetItems)

		// Notifications
		notifications := protected.Group("/notification")
		{
			notifications.GET("/", notificationHandler.GetNotifications)
			notifications.GET("/:id/", notificationHandler.GetNotificationData)
			notifications.PATCH("/:id/", notificationHandler.MarkNotificationAsRead)
			notifications.GET("/public_key/", notificationHandler.GetPublicKey)
			notifications.POST("/notify/", notificationHandler.SendNotificationMessages)
		}

		// File uploads
		files := protected.Group("/files")
		{
			files.GET("/", facilityHandler.ViewUpload)
			files.POST("/", facilityHandler.CreateUpload)
			files.GET("/:id/", facilityHandler.RetrieveUpload)
			files.PATCH("/:id/", facilityHandler.EditUpload)
		}

		// Investigations
		protected.GET("/investigation/", consultationHandler.ListInvestigations)
		protected.GET("/investigation/group/", consultationHandler.ListInvestigationGroups)

		// Event types
		protected.GET("/event_types/", consultationHandler.ListEventTypes)

		// ICD11 and Medibase
		protected.GET("/icd/", consultationHandler.ListICD11Diagnosis)
		protected.GET("/icd/:id/", consultationHandler.GetICD11Diagnosis)
		protected.GET("/medibase/", consultationHandler.ListMedibaseMedicines)

		// Shifting
		shifting := protected.Group("/shift")
		{
			shifting.GET("/", facilityHandler.ListShiftRequests)
			shifting.POST("/", facilityHandler.CreateShift)
			shifting.GET("/:id/", facilityHandler.GetShiftDetails)
			shifting.PUT("/:id/", facilityHandler.UpdateShift)
			shifting.POST("/:id/transfer/", facilityHandler.CompleteTransfer)
			shifting.GET("/:id/comment/", facilityHandler.GetShiftComments)
			shifting.POST("/:id/comment/", facilityHandler.AddShiftComments)
		}

		// Time Log (Clock In/Out)
		timelog := protected.Group("/timelog")
		{
			timelog.POST("/clock-in", timeLogHandler.ClockIn)
			timelog.POST("/clock-out", timeLogHandler.ClockOut)
			timelog.GET("/my-logs", timeLogHandler.GetMyLogs)
			timelog.GET("/status", timeLogHandler.GetCurrentStatus)
			timelog.GET("/user/:id", timeLogHandler.GetUserLogs) // Admin only
		}

		// Timetable
		timetable := protected.Group("/timetable")
		{
			timetable.GET("/:id", timetableHandler.GetTimetable)
			timetable.POST("/:id/slot", timetableHandler.CreateSlot)
			timetable.PUT("/slot/:id", timetableHandler.UpdateSlot)
			timetable.DELETE("/slot/:id", timetableHandler.DeleteSlot)
		}

		// Sections
		sections := protected.Group("/sections")
		{
			sections.GET("/:id", timetableHandler.GetSection)
			sections.PUT("/:id", timetableHandler.UpdateSection)
			sections.DELETE("/:id", timetableHandler.DeleteSection)
		}

		// Doctor → physiotherapist referrals (core DocTheraLink workflow)
		referral.RegisterRoutes(protected, database.GetDB())

		// Physical Therapy routes - TEMPORARILY DISABLED due to model conflicts
		// All therapy-related functionality is disabled until database migration issues are resolved
		/*
			therapy := protected.Group("/therapy")
			{
				// All therapy routes temporarily disabled - see notes_disabled.go for details
			}
		*/

		// Admin routes (require admin privileges)
		admin := protected.Group("")
		admin.Use(auth.RequireRole(
			models.RoleStateAdmin,
			models.RoleDistrictAdmin,
			models.RoleLocalBodyAdmin,
		))
		{
			// CSV downloads
			admin.GET("/download/facility/", facilityHandler.DownloadFacility)
			admin.GET("/download/patient/", patientHandler.DownloadPatients)
			admin.GET("/download/shift/", facilityHandler.DownloadShiftRequests)
		}
	}
}
