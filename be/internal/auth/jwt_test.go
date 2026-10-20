package auth

import (
	"errors"
	"testing"
	"time"

	"klyniva/internal/config"
	"klyniva/internal/models"
)

func newManager(expiry time.Duration) *JWTManager {
	return NewJWTManager(&config.Config{JWT: config.JWTConfig{
		Secret: "unit-test-secret", Expiry: expiry, RefreshExpiry: time.Hour,
	}})
}

var testUser = &models.User{ID: 42, Username: "physio", UserType: models.RolePhysioTherapist}

func TestTokenPairRoundTrip(t *testing.T) {
	m := newManager(time.Minute)
	pair, err := m.GenerateTokenPair(testUser)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := m.ValidateToken(pair.Access)
	if err != nil {
		t.Fatalf("access token rejected: %v", err)
	}
	if claims.UserID != 42 || claims.Username != "physio" || claims.UserType != models.RolePhysioTherapist {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	if _, err := m.ValidateRefreshToken(pair.Refresh); err != nil {
		t.Fatalf("refresh token rejected: %v", err)
	}
}

func TestTokenTypesAreNotInterchangeable(t *testing.T) {
	m := newManager(time.Minute)
	pair, _ := m.GenerateTokenPair(testUser)

	if _, err := m.ValidateToken(pair.Refresh); !errors.Is(err, ErrWrongTokenType) {
		t.Errorf("refresh used as access: want ErrWrongTokenType, got %v", err)
	}
	if _, err := m.ValidateRefreshToken(pair.Access); !errors.Is(err, ErrWrongTokenType) {
		t.Errorf("access used as refresh: want ErrWrongTokenType, got %v", err)
	}
}

func TestRejectsExpiredAndForeignTokens(t *testing.T) {
	expired, _ := newManager(-time.Minute).GenerateTokenPair(testUser)
	if _, err := newManager(time.Minute).ValidateToken(expired.Access); err == nil {
		t.Error("expired token accepted")
	}

	other := NewJWTManager(&config.Config{JWT: config.JWTConfig{Secret: "different", Expiry: time.Minute}})
	foreign, _ := other.GenerateTokenPair(testUser)
	if _, err := newManager(time.Minute).ValidateToken(foreign.Access); err == nil {
		t.Error("token signed with another secret accepted")
	}

	if _, err := newManager(time.Minute).ValidateToken("not.a.jwt"); err == nil {
		t.Error("garbage accepted")
	}
}

func TestRoleHelpers(t *testing.T) {
	if !IsAdmin(models.RoleDistrictAdmin) || IsAdmin(models.RoleDoctor) {
		t.Error("IsAdmin misclassifies roles")
	}
	if CanModifyResource(models.RoleStaffReadOnly) || !CanModifyResource(models.RoleNurse) {
		t.Error("CanModifyResource misclassifies roles")
	}
}
