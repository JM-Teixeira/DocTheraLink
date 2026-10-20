package config

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	strong := strings.Repeat("s", 32)

	cases := []struct {
		name    string
		mode    string
		secret  string
		demo    bool
		wantErr bool
	}{
		{"debug mode allows defaults", "debug", defaultJWTSecret, true, false},
		{"release with default secret", "release", defaultJWTSecret, false, true},
		{"release with short secret", "release", "short", false, true},
		{"release with demo users", "release", strong, true, true},
		{"release properly configured", "release", strong, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{
				Server: ServerConfig{Mode: tc.mode},
				JWT:    JWTConfig{Secret: tc.secret},
				Seed:   SeedConfig{DemoUsers: tc.demo},
			}
			if err := c.Validate(); (err != nil) != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestLoadDemoUsersDefaultFollowsMode(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	t.Setenv("SEED_DEMO_USERS", "")
	if Load().Seed.DemoUsers {
		t.Error("demo users must default to off in release mode")
	}

	t.Setenv("GIN_MODE", "debug")
	if !Load().Seed.DemoUsers {
		t.Error("demo users should default to on in debug mode")
	}
}

func TestEnvParsing(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example,https://b.example")
	t.Setenv("JWT_EXPIRY", "15m")
	t.Setenv("BCRYPT_COST", "not-a-number")

	c := Load()
	if len(c.CORS.AllowedOrigins) != 2 {
		t.Errorf("origins: %v", c.CORS.AllowedOrigins)
	}
	if c.JWT.Expiry.Minutes() != 15 {
		t.Errorf("expiry: %v", c.JWT.Expiry)
	}
	if c.Security.BcryptCost != 12 {
		t.Errorf("invalid int should fall back to default, got %d", c.Security.BcryptCost)
	}
}
