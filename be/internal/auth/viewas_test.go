package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
)

func viewAsRequest(method, role, clinicID string, claims *models.TokenClaims) (*httptest.ResponseRecorder, *gin.Context, bool) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/x", nil)
	if role != "" {
		c.Request.Header.Set(ViewAsRoleHeader, role)
		c.Request.Header.Set(ViewAsClinicHeader, clinicID)
	}
	return w, c, applyViewAs(c, claims)
}

const viewClinic = "6b1f3c1e-5a8e-4c1d-9a0e-000000000002"

func TestViewAs(t *testing.T) {
	admin := func() *models.TokenClaims { return &models.TokenClaims{UserID: 1, UserType: models.RoleStateAdmin} }

	// No headers: nothing changes
	claims := admin()
	if _, _, ok := viewAsRequest(http.MethodGet, "", "", claims); !ok || claims.UserType != models.RoleStateAdmin {
		t.Fatal("plain request changed")
	}

	// An admin reads as a physio of a clinic
	claims = admin()
	_, c, ok := viewAsRequest(http.MethodGet, "PhysioTherapist", viewClinic, claims)
	if !ok || claims.UserType != models.RolePhysioTherapist {
		t.Fatalf("view as: %v %v", ok, claims.UserType)
	}
	if va, on := ViewingAs(c); !on || va.RealRole != models.RoleStateAdmin || va.Clinic != viewClinic {
		t.Errorf("ViewingAs: %+v", va)
	}
	home, _ := clinic.NewGormResolver(nil).HomeClinic(c.Request.Context(), 1)
	if home == nil || *home != viewClinic {
		t.Errorf("clinic override: %v", home)
	}

	// An admin working on a clinic: full rights, that clinic
	gin.SetMode(gin.TestMode)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodPost, "/x", nil)
	c2.Request.Header.Set(WorkingClinicHeader, viewClinic)
	claims = admin()
	if !applyViewAs(c2, claims) || claims.UserType != models.RoleStateAdmin {
		t.Fatal("working clinic must not change the role or block writes")
	}
	if home, _ := clinic.NewGormResolver(nil).HomeClinic(c2.Request.Context(), 1); home == nil || *home != viewClinic {
		t.Errorf("working clinic: %v", home)
	}
	// Ignored for everyone else
	c3, _ := gin.CreateTestContext(httptest.NewRecorder())
	c3.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	c3.Request.Header.Set(WorkingClinicHeader, viewClinic)
	before := c3.Request.Context()
	applyViewAs(c3, &models.TokenClaims{UserID: 7, UserType: models.RolePhysioTherapist})
	if c3.Request.Context() != before {
		t.Error("a physio cannot pick another clinic")
	}

	// Read-only, admins only, known roles only
	cases := []struct {
		method, role, clinic string
		claims               *models.TokenClaims
		want                 int
	}{
		{http.MethodPost, "PhysioTherapist", viewClinic, admin(), http.StatusForbidden},
		{http.MethodGet, "PhysioTherapist", viewClinic, &models.TokenClaims{UserID: 2, UserType: models.RoleCoordinator}, http.StatusForbidden},
		{http.MethodGet, "StateAdmin", viewClinic, admin(), http.StatusBadRequest},
		{http.MethodGet, "Nurse", "nope", admin(), http.StatusBadRequest},
	}
	for _, tc := range cases {
		w, _, ok := viewAsRequest(tc.method, tc.role, tc.clinic, tc.claims)
		if ok || w.Code != tc.want {
			t.Errorf("%s %s: ok=%v code=%d want %d", tc.method, tc.role, ok, w.Code, tc.want)
		}
	}
}
