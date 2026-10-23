package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"klyniva/internal/config"
	"klyniva/internal/database"
)

func main() {
	var (
		seed = flag.Bool("seed", false, "Run database seeding after migration")
		// A new server without demo data has no account: this makes admin.user
		createAdmin = flag.Bool("create-admin", false, "Create the first admin (admin.user) if there is none and print its one-time password")
	)
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.ConnectWithRetry(cfg, 30, 2*time.Second); err != nil {
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
		if err := database.SeedData(cfg.Seed.DemoUsers); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeding completed successfully")
	}

	if *createAdmin {
		password, created, err := database.CreateFirstAdmin()
		if err != nil {
			log.Fatalf("Failed to create the first admin: %v", err)
		}
		if !created {
			log.Println("An admin already exists; nothing created")
		} else {
			fmt.Printf("\nFirst admin: %s\nOne-time password: %s\n(change it at the first sign-in)\n",
				database.FirstAdminUsername, password)
		}
	}
}
