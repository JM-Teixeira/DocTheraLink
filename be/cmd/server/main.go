// Package main provides the entry point for the DocTheraLink backend server.
// This is a Go-based REST API server that provides healthcare management
// functionality including patient management, therapy scheduling, and
// facility administration.
//
// The server uses the Gin HTTP framework with the following key features:
// - JWT-based authentication and authorization
// - Role-based access control for healthcare workers
// - PostgreSQL database with GORM ORM
// - Swagger API documentation
// - CORS support for frontend integration
// - Graceful shutdown handling
// - Health check endpoints
//
// Environment variables are used for configuration, allowing the same
// binary to run in different environments (development, staging, production).
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"care-backend/internal/api"
	"care-backend/internal/auth"
	"care-backend/internal/config"
	"care-backend/internal/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title CARE Backend API
// @version 1.0
// @description Healthcare management system backend API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.care.ohc.network/support
// @contact.email support@ohc.network

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:9000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// main is the entry point for the DocTheraLink backend server.
// It initializes all components and starts the HTTP server with graceful shutdown.
func main() {
	// Load configuration from environment variables and .env file
	// This allows the same binary to work in different environments
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Set Gin framework mode (debug, release, test)
	// In production, this should be "release" for better performance
	gin.SetMode(cfg.Server.Mode)

	// Connect to PostgreSQL database using GORM
	// The connection includes automatic retry logic and connection pooling
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run database migrations to ensure schema is up to date
	// This creates/updates tables for all models automatically
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Seed initial data including states, districts, skills and (outside release mode) demo users
	if err := database.SeedData(cfg.Seed.DemoUsers); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	// Initialize JWT manager for token generation and validation
	// Handles both access tokens (short-lived) and refresh tokens (long-lived)
	jwtManager := auth.NewJWTManager(cfg)

	// Create Gin HTTP router with middleware stack
	router := gin.New()

	// Add essential middleware for logging and panic recovery
	router.Use(gin.Logger())   // Logs all HTTP requests
	router.Use(gin.Recovery()) // Recovers from panics and returns 500

	// Configure CORS (Cross-Origin Resource Sharing) for frontend integration
	// Allows the React frontend to make API calls from different origins
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins, // Frontend URLs
		AllowMethods:     cfg.CORS.AllowedMethods, // HTTP methods
		AllowHeaders:     cfg.CORS.AllowedHeaders, // Request headers
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,           // Allow cookies and auth headers
		MaxAge:           12 * time.Hour, // Cache preflight requests
	}
	router.Use(cors.New(corsConfig))

	// Health check endpoint for monitoring and load balancers
	// Returns server status, timestamp, and version information
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})

	// Readiness check: only report ready when the database answers, so
	// Kubernetes stops routing traffic to a pod that lost its connection.
	router.GET("/ready", func(c *gin.Context) {
		sqlDB, err := database.GetDB().DB()
		if err == nil {
			err = sqlDB.PingContext(c.Request.Context())
		}
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "database": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready", "database": "up"})
	})

	// Swagger API documentation endpoint (only in development)
	// Provides interactive API documentation at /swagger/
	if cfg.Server.Swagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Setup all API routes under /api/v1 prefix
	// Includes authentication, user management, facilities, patients, therapy, etc.
	apiGroup := router.Group("/api/v1")
	api.SetupRoutes(apiGroup, jwtManager, cfg)

	// Create HTTP server with configured address and router
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start HTTP server in a separate goroutine to allow graceful shutdown
	go func() {
		log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal (SIGINT or SIGTERM) for graceful shutdown
	// This allows the server to finish processing current requests
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete before forcing shutdown
	// This ensures data integrity and proper connection cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
