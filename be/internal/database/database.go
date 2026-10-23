package database

import (
	"fmt"
	"log"
	"time"

	"klyniva/internal/config"
	"klyniva/internal/models"
	"klyniva/internal/modules/attachment"
	"klyniva/internal/modules/feedback"
	"klyniva/internal/modules/inventory"
	"klyniva/internal/modules/notification"
	"klyniva/internal/modules/patient"
	"klyniva/internal/modules/referral"
	"klyniva/internal/modules/schedule"
	"klyniva/internal/modules/staff"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect initializes the database connection
func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
		cfg.Database.Timezone,
	)

	// Configure GORM logger
	var gormLogger logger.Interface
	if cfg.Server.Debug {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	log.Println("Database connection established")
	return nil
}

// ConnectWithRetry calls Connect until it succeeds or attempts run out.
// In containers the database often starts after the API; retrying avoids
// crash loops during startup.
func ConnectWithRetry(cfg *config.Config, attempts int, delay time.Duration) error {
	return retry(attempts, delay, func() error { return Connect(cfg) })
}

// retry runs fn up to attempts times, sleeping delay between failures, and
// returns the last error.
func retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 1; i <= attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i < attempts {
			log.Printf("database not ready (attempt %d/%d): %v", i, attempts, err)
			time.Sleep(delay)
		}
	}
	return err
}

// Migrate runs database migrations
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Enable UUID extension
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("Warning: Could not create uuid-ossp extension: %v", err)
	}

	// Tables are created in dependency order. Tables of removed CARE modules
	// (beds, consultations, assets, …) are no longer migrated; existing
	// databases keep them until versioned migrations drop them (TODO.md).
	phases := []struct {
		name   string
		models []any
	}{
		{"reference data", []any{&models.State{}, &models.District{}, &models.LocalBody{}, &models.Ward{}}},
		{"clinics", []any{&models.Facility{}}},
		{"users", []any{&models.User{}, &models.Skill{}, &models.UserSkill{}}},
		{"employment history", staff.Models()},
		{"patients", []any{&models.Patient{}}},
		{"patient consents", patient.Models()},
		{"staff time log", []any{&models.TimeLog{}}},
		{"inventory", inventory.Models()},
		{"time board", schedule.Models()},
		{"referrals", referral.Models()},
		{"notifications", notification.Models()},
		{"attachments", attachment.Models()},
		{"suggestions", feedback.Models()},
	}
	for i, p := range phases {
		log.Printf("Migrating phase %d: %s...", i+1, p.name)
		if err := DB.AutoMigrate(p.models...); err != nil {
			return fmt.Errorf("failed to migrate %s: %w", p.name, err)
		}
	}

	if err := DB.Exec(models.UniqueClinicCode).Error; err != nil {
		return fmt.Errorf("failed to create clinic code index: %w", err)
	}
	for _, stmt := range patient.UniqueIndexes {
		if err := DB.Exec(stmt).Error; err != nil {
			return fmt.Errorf("failed to create patient indexes: %w", err)
		}
	}

	log.Println("Database migration completed")
	return nil
}

// Helper function at the top of the file (or somewhere accessible)
func ptr(s string) *string {
	return &s
}

// SeedData inserts initial data into the database. Demo users share a
// publicly known password, so they are only inserted when includeDemoUsers is set.
func SeedData(includeDemoUsers bool) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Seed Portugal districts (Distritos)
	states := []models.State{
		{ID: 1, Name: "Aveiro"},
		{ID: 2, Name: "Beja"},
		{ID: 3, Name: "Braga"},
		{ID: 4, Name: "Bragança"},
		{ID: 5, Name: "Castelo Branco"},
		{ID: 6, Name: "Coimbra"},
		{ID: 7, Name: "Évora"},
		{ID: 8, Name: "Faro"},
		{ID: 9, Name: "Guarda"},
		{ID: 10, Name: "Leiria"},
		{ID: 11, Name: "Lisboa"},
		{ID: 12, Name: "Portalegre"},
		{ID: 13, Name: "Porto"},
		{ID: 14, Name: "Santarém"},
		{ID: 15, Name: "Setúbal"},
		{ID: 16, Name: "Viana do Castelo"},
		{ID: 17, Name: "Vila Real"},
		{ID: 18, Name: "Viseu"},
		{ID: 19, Name: "Região Autónoma dos Açores"},
		{ID: 20, Name: "Região Autónoma da Madeira"},
	}

	for _, state := range states {
		if err := DB.FirstOrCreate(&state, models.State{ID: state.ID}).Error; err != nil {
			return fmt.Errorf("failed to seed state %s: %w", state.Name, err)
		}
	}

	// Seed major municipalities (Concelhos) - sample for main districts
	districts := []models.District{
		// Lisboa District
		{ID: 1, Name: "Lisboa", StateID: 11},
		{ID: 2, Name: "Cascais", StateID: 11},
		{ID: 3, Name: "Sintra", StateID: 11},
		{ID: 4, Name: "Oeiras", StateID: 11},
		{ID: 5, Name: "Loures", StateID: 11},
		{ID: 6, Name: "Amadora", StateID: 11},
		// Porto District
		{ID: 7, Name: "Porto", StateID: 13},
		{ID: 8, Name: "Vila Nova de Gaia", StateID: 13},
		{ID: 9, Name: "Matosinhos", StateID: 13},
		{ID: 10, Name: "Gondomar", StateID: 13},
		{ID: 11, Name: "Maia", StateID: 13},
		// Other major cities
		{ID: 12, Name: "Coimbra", StateID: 6},
		{ID: 13, Name: "Braga", StateID: 3},
		{ID: 14, Name: "Faro", StateID: 8},
		{ID: 15, Name: "Aveiro", StateID: 1},
		{ID: 16, Name: "Setúbal", StateID: 15},
		{ID: 17, Name: "Leiria", StateID: 10},
		{ID: 18, Name: "Viseu", StateID: 18},
		{ID: 19, Name: "Funchal", StateID: 20},
		{ID: 20, Name: "Ponta Delgada", StateID: 19},
	}

	for _, district := range districts {
		if err := DB.FirstOrCreate(&district, models.District{ID: district.ID}).Error; err != nil {
			return fmt.Errorf("failed to seed district %s: %w", district.Name, err)
		}
	}

	// Seed basic skills
	skills := []models.Skill{
		{Name: "Critical Care", Description: ptr("Intensive care and critical patient management")},
		{Name: "Emergency Medicine", Description: ptr("Emergency medical procedures and protocols")},
		{Name: "Cardiology", Description: ptr("Heart and cardiovascular system expertise")},
		{Name: "Pulmonology", Description: ptr("Respiratory system and lung care")},
		{Name: "Nursing Care", Description: ptr("General nursing and patient care")},
		{Name: "Physiotherapy", Description: ptr("Physical therapy and rehabilitation")},
		{Name: "Laboratory", Description: ptr("Laboratory testing and diagnostics")},
		{Name: "Radiology", Description: ptr("Medical imaging and interpretation")},
	}

	for _, skill := range skills {
		if err := DB.FirstOrCreate(&skill, models.Skill{Name: skill.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed skill %s: %w", skill.Name, err)
		}
	}

	if err := staff.SeedSkills(DB); err != nil {
		return fmt.Errorf("skills catalogue: %w", err)
	}

	if !includeDemoUsers {
		log.Println("Database seeding completed (demo users skipped)")
		return nil
	}

	if err := seedDemoClinic(); err != nil {
		return err
	}
	// Second demo clinic with realistic PT-PT data
	if err := seedClimutua(time.Now()); err != nil {
		return fmt.Errorf("climutua demo: %w", err)
	}
	if err := seedStockAccess(); err != nil {
		return fmt.Errorf("stock permissions: %w", err)
	}
	log.Println("Database seeding completed")
	return nil
}

// DemoClinicID is the fixed ID of the demo clinic (stable across reseeds).
const DemoClinicID = "6b1f3c1e-5a8e-4c1d-9a0e-000000000001"

// demoPasswordHash is bcrypt("1234@1"), shared by all demo accounts.
const demoPasswordHash = "$2a$12$fdWvXHhcyw5Fj5wLFrCeOea3ZupC9XaGGR9pJN06m76epTOcN7C82"

// seedDemoClinic creates one clinic and demo users working in it, for
// development and tests only (release mode refuses demo users).
// Admin accounts are organisation-wide and get no clinic.
func seedDemoClinic() error {
	demo := models.Facility{ID: DemoClinicID, Name: "Klyniva Demo Clinic", Code: demoClinicCode, Address: "Lisboa"}
	if err := DB.FirstOrCreate(&demo, models.Facility{ID: DemoClinicID}).Error; err != nil {
		return fmt.Errorf("failed to seed demo clinic: %w", err)
	}

	type demoUser struct {
		username, first, last string
		role                  models.UserRole
		inClinic              bool
	}
	users := []demoUser{
		// The organisation admin is an account, not a person (no name)
		{"admin.user", "Administrador", "", models.RoleStateAdmin, false},
		{"jorge.matos", "Jorge", "Matos", models.RoleDistrictAdmin, false},
		{"luis.pereira", "Luís", "Pereira", models.RoleLocalBodyAdmin, false},
		{"de.john.smith", "John", "Smith", models.RoleDoctor, true},
		{"de.sarah.johnson", "Sarah", "Johnson", models.RolePhysioTherapist, true},
		{"de.tiago.martins", "Tiago", "Martins", models.RolePhysioTherapist, true},
		{"de.mary.williams", "Mary", "Williams", models.RoleNurse, true},
		{"de.rita.lopes", "Rita", "Lopes", models.RoleReceptionist, true},
		{"de.marta.ribeiro", "Marta", "Ribeiro", models.RoleCoordinator, true},
		{"de.nuno.dias", "Nuno", "Dias", models.RoleStaff, true},
	}
	if err := applyNamingStandard(); err != nil {
		return err
	}
	for _, d := range users {
		u := models.User{
			Username: d.username, Email: d.username + "@klyniva.com",
			FirstName: d.first, LastName: d.last, UserType: d.role,
			Password: demoPasswordHash, IsActive: true,
		}
		if d.inClinic {
			clinicID := DemoClinicID
			u.HomeFacilityID = &clinicID
		}
		if err := DB.FirstOrCreate(&u, models.User{Username: d.username}).Error; err != nil {
			return fmt.Errorf("failed to seed demo user %s: %w", d.username, err)
		}
		// Bring demo users created by older versions up to date
		updates := map[string]any{"first_name": d.first, "last_name": d.last}
		if d.inClinic {
			updates["home_facility_id"] = DemoClinicID
		}
		if err := DB.Model(&models.User{}).Where("username = ?", d.username).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update demo user %s: %w", d.username, err)
		}
	}
	// Demo stock and a time board with divisions and shifts
	if err := inventory.SeedDemo(DB, DemoClinicID); err != nil {
		return err
	}
	return schedule.SeedDemo(DB, DemoClinicID, time.Now())
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
