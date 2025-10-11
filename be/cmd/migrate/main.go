package main

import (
	"flag"
	"log"

	"care-backend/internal/config"
	"care-backend/internal/database"
)

func main() {
	var (
		seed = flag.Bool("seed", false, "Run database seeding after migration")
	)
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")

	// Run seeding if requested
	if *seed {
		if err := database.SeedData(); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeding completed successfully")
	}
}