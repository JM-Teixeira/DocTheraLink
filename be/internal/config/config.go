package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Logging  LoggingConfig
	Upload   UploadConfig
	Email    EmailConfig
	Redis    RedisConfig
	Security SecurityConfig
	Seed     SeedConfig
}

// SeedConfig controls which bootstrap data is inserted on startup.
type SeedConfig struct {
	// DemoUsers inserts the demo accounts (shared, publicly known password).
	// Defaults to on outside release mode and off in release mode.
	DemoUsers bool
}

// defaultJWTSecret is the development fallback; Validate rejects it in release mode.
const defaultJWTSecret = "your-secret-key"

// Validate checks settings that must never ship to production with
// development defaults. It only enforces them in release mode.
func (c *Config) Validate() error {
	if c.Server.Mode != "release" {
		return nil
	}
	if c.JWT.Secret == defaultJWTSecret || len(c.JWT.Secret) < 32 {
		return errors.New("JWT_SECRET must be set to a random value of at least 32 characters in release mode")
	}
	if c.Seed.DemoUsers {
		return errors.New("SEED_DEMO_USERS must be false in release mode")
	}
	return nil
}

type ServerConfig struct {
	Port    string
	Host    string
	Mode    string
	Debug   bool
	Swagger bool
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	Timezone string
}

type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type UploadConfig struct {
	MaxSize int64
	Path    string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type SecurityConfig struct {
	BcryptCost        int
	RateLimitRequests int
	RateLimitWindow   time.Duration
	GoogleMapsAPIKey  string
	GovDataAPIKey     string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	mode := getEnv("GIN_MODE", "debug")

	return &Config{
		Seed: SeedConfig{
			DemoUsers: getEnvBool("SEED_DEMO_USERS", mode != "release"),
		},
		Server: ServerConfig{
			Port:    getEnv("PORT", "9000"),
			Host:    getEnv("HOST", "0.0.0.0"),
			Mode:    mode,
			Debug:   getEnvBool("DEBUG", true),
			Swagger: getEnvBool("ENABLE_SWAGGER", true),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "klyniva"),
			Password: getEnv("DB_PASSWORD", "klyniva"),
			Name:     getEnv("DB_NAME", "klyniva"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			Timezone: getEnv("DB_TIMEZONE", "UTC"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", defaultJWTSecret),
			Expiry:        getEnvDuration("JWT_EXPIRY", 24*time.Hour),
			RefreshExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 168*time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:4000"}),
			AllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		Upload: UploadConfig{
			MaxSize: getEnvInt64("MAX_UPLOAD_SIZE", 10485760), // 10MB
			Path:    getEnv("UPLOAD_PATH", "./uploads"),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "noreply@klyniva.com"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Security: SecurityConfig{
			BcryptCost:        getEnvInt("BCRYPT_COST", 12),
			RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:   getEnvDuration("RATE_LIMIT_WINDOW", time.Hour),
			GoogleMapsAPIKey:  getEnv("GOOGLE_MAPS_API_KEY", ""),
			GovDataAPIKey:     getEnv("GOV_DATA_API_KEY", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
