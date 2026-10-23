package database

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"klyniva/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// FirstAdminUsername is the organisation admin created on a new server.
const FirstAdminUsername = "admin.user"

// CreateFirstAdmin makes the organisation admin of a fresh installation (no
// demo data) and returns its one-time password, which must be changed at
// the first sign-in. It does nothing when an admin already exists.
func CreateFirstAdmin() (password string, created bool, err error) {
	var n int64
	if err := DB.Model(&models.User{}).Where("user_type IN ?", []models.UserRole{
		models.RoleStateAdmin, models.RoleDistrictAdmin,
	}).Count(&n).Error; err != nil || n > 0 {
		return "", false, err
	}
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return "", false, err
	}
	password = base64.RawURLEncoding.EncodeToString(raw)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", false, err
	}
	admin := models.User{
		Username: FirstAdminUsername, Email: FirstAdminUsername + "@localhost", FirstName: "Administrador",
		UserType: models.RoleStateAdmin, Password: string(hash), IsActive: true,
		EmploymentStatus: "active", MustChangePassword: true,
	}
	if err := DB.Create(&admin).Error; err != nil {
		return "", false, fmt.Errorf("first admin: %w", err)
	}
	return password, true, nil
}
