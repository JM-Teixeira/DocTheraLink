package database

import (
	"fmt"
	"log"
	"time"

	"care-backend/internal/config"
	"care-backend/internal/models"

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

// Migrate runs database migrations
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Enable UUID extension
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("Warning: Could not create uuid-ossp extension: %v", err)
	}

	// Phase 1: Core models without complex relationships
	log.Println("Migrating Phase 1: Core models...")
	err := DB.AutoMigrate(
		&models.State{},
		&models.District{},
		&models.LocalBody{},
		&models.Ward{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate Phase 1 models: %w", err)
	}

	// Phase 2: Facility and location models
	log.Println("Migrating Phase 2: Facility models...")
	err = DB.AutoMigrate(
		&models.Facility{},
		&models.Location{},
		&models.Bed{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate Phase 2 models: %w", err)
	}

	// Phase 3: User models
	log.Println("Migrating Phase 3: User models...")
	err = DB.AutoMigrate(
		&models.User{},
		&models.Skill{},
		&models.UserSkill{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate Phase 3 models: %w", err)
	}

	// Phase 4: Patient model (simplified)
	log.Println("Migrating Phase 4: Patient model...")
	err = DB.AutoMigrate(
		&models.Patient{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate Phase 4 models: %w", err)
	}

	// Phase 5: Consultation models
	log.Println("Migrating Phase 5: Consultation models...")
	err = DB.AutoMigrate(
		&models.Consultation{},
		&models.DailyRound{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate Phase 5 models: %w", err)
	}

	// Phase 6: Basic inventory models
	log.Println("Migrating Phase 6: Inventory models...")
	err = DB.AutoMigrate(
		&models.InventoryItem{},
		&models.Asset{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate Phase 6 models: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed")
	return nil
}

// Helper function at the top of the file (or somewhere accessible)
func ptr(s string) *string {
	return &s
}

// SeedData inserts initial data into the database
func SeedData() error {
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

	// Seed inventory items
	inventoryItems := []models.InventoryItem{
		{Name: "N95 Mask", Category: "PPE", Unit: "nos"},
		{Name: "Surgical Mask", Category: "PPE", Unit: "nos"},
		{Name: "Hand Sanitizer", Category: "Hygiene", Unit: "bottles"},
		{Name: "Gloves", Category: "PPE", Unit: "pairs"},
		{Name: "Face Shield", Category: "PPE", Unit: "nos"},
		{Name: "Oxygen Cylinder", Category: "Medical Equipment", Unit: "nos"},
		{Name: "Ventilator", Category: "Medical Equipment", Unit: "nos"},
		{Name: "Pulse Oximeter", Category: "Medical Equipment", Unit: "nos"},
		{Name: "Thermometer", Category: "Medical Equipment", Unit: "nos"},
		{Name: "Paracetamol", Category: "Medicine", Unit: "tablets"},
	}

	for _, item := range inventoryItems {
		if err := DB.FirstOrCreate(&item, models.InventoryItem{Name: item.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed inventory item %s: %w", item.Name, err)
		}
	}

	// TODO: Remove these test users in production - they are for development/testing only
	// Create bootstrap users with test credentials
	// Password for all test users: 1234@1
	testUsers := []models.User{
		{
			Username:  "stateadmin",
			Email:     "stateadmin@doctheralink.com",
			FirstName: "State",
			LastName:  "Administrator",
			UserType:  models.RoleStateAdmin,
			Password:  "$2a$12$fdWvXHhcyw5Fj5wLFrCeOea3ZupC9XaGGR9pJN06m76epTOcN7C82", // 1234@1
			IsActive:  true,
		},
		{
			Username:  "districtadmin",
			Email:     "districtadmin@doctheralink.com",
			FirstName: "District",
			LastName:  "Administrator",
			UserType:  models.RoleDistrictAdmin,
			Password:  "$2a$12$fdWvXHhcyw5Fj5wLFrCeOea3ZupC9XaGGR9pJN06m76epTOcN7C82", // 1234@1
			IsActive:  true,
		},
		{
			Username:  "facilitycoordinator",
			Email:     "coordinator@doctheralink.com",
			FirstName: "Facility",
			LastName:  "Coordinator",
			UserType:  models.RoleLocalBodyAdmin,
			Password:  "$2a$12$fdWvXHhcyw5Fj5wLFrCeOea3ZupC9XaGGR9pJN06m76epTOcN7C82", // 1234@1
			IsActive:  true,
		},
		{
			Username:  "doctor",
			Email:     "doctor@doctheralink.com",
			FirstName: "Dr. John",
			LastName:  "Smith",
			UserType:  models.RoleDoctor,
			Password:  "$2a$12$fdWvXHhcyw5Fj5wLFrCeOea3ZupC9XaGGR9pJN06m76epTOcN7C82", // 1234@1
			IsActive:  true,
		},
		{
			Username:  "physiotherapist",
			Email:     "physio@doctheralink.com",
			FirstName: "Sarah",
			LastName:  "Johnson",
			UserType:  models.RolePhysioTherapist,
			Password:  "$2a$12$fdWvXHhcyw5Fj5wLFrCeOea3ZupC9XaGGR9pJN06m76epTOcN7C82", // 1234@1
			IsActive:  true,
		},
		{
			Username:  "nurse",
			Email:     "nurse@doctheralink.com",
			FirstName: "Mary",
			LastName:  "Williams",
			UserType:  models.RoleNurse,
			Password:  "$2a$12$fdWvXHhcyw5Fj5wLFrCeOea3ZupC9XaGGR9pJN06m76epTOcN7C82", // 1234@1
			IsActive:  true,
		},
		{
			Username:  "staff",
			Email:     "staff@doctheralink.com",
			FirstName: "General",
			LastName:  "Staff",
			UserType:  models.RoleStaff,
			Password:  "$2a$12$fdWvXHhcyw5Fj5wLFrCeOea3ZupC9XaGGR9pJN06m76epTOcN7C82", // 1234@1
			IsActive:  true,
		},
	}

	for _, user := range testUsers {
		if err := DB.FirstOrCreate(&user, models.User{Username: user.Username}).Error; err != nil {
			return fmt.Errorf("failed to seed test user %s: %w", user.Username, err)
		}
	}

	log.Println("Database seeding completed")
	return nil
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