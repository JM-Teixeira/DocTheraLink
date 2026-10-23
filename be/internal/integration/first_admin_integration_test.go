//go:build integration

package integration

import (
	"testing"

	"klyniva/internal/database"
	"klyniva/internal/models"
)

// TestCreateFirstAdmin: a fresh server gets admin.user with a one-time
// password that must be changed; with an admin already there nothing happens.
func TestCreateFirstAdmin(t *testing.T) {
	startDB(t)
	db := database.GetDB()
	if _, created, err := database.CreateFirstAdmin(); err != nil || created {
		t.Fatalf("an admin exists (demo data), nothing must be created: %v %v", created, err)
	}
	if err := db.Unscoped().Where("user_type IN ?", []models.UserRole{models.RoleStateAdmin, models.RoleDistrictAdmin}).
		Delete(&models.User{}).Error; err != nil {
		t.Fatal(err)
	}
	password, created, err := database.CreateFirstAdmin()
	if err != nil || !created || len(password) < 12 {
		t.Fatalf("first admin: %v %v %q", created, err, password)
	}
	var admin models.User
	if err := db.Where("username = ?", database.FirstAdminUsername).First(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if !admin.MustChangePassword || admin.UserType != models.RoleStateAdmin || !admin.IsActive {
		t.Errorf("first admin: %+v", admin)
	}
	if _, created, _ := database.CreateFirstAdmin(); created {
		t.Error("a second call must not create another admin")
	}
}
