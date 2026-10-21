package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/config"
	"klyniva/internal/models"
	"klyniva/internal/modules/referral"

	"github.com/gin-gonic/gin"
)

func TestHTTPNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc, _, _ := newTestService()
	svc.ReferralEvent(context.Background(), referral.Event{Kind: referral.EventCreated, Referral: ref(nil), ActorID: doctorID})

	jwt := auth.NewJWTManager(&config.Config{JWT: config.JWTConfig{Secret: "s", Expiry: time.Hour, RefreshExpiry: time.Hour}})
	r := gin.New()
	mount(r.Group("/api/v1", auth.AuthMiddleware(jwt)).Group("/notifications"), NewHandler(svc))

	tok := func(id uint) string {
		p, _ := jwt.GenerateTokenPair(&models.User{ID: id, Username: "u", UserType: models.RolePhysioTherapist})
		return p.Access
	}
	do := func(method, path string, user uint) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		if user != 0 {
			req.Header.Set("Authorization", "Bearer "+tok(user))
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	if w := do("GET", "/api/v1/notifications/", 0); w.Code != http.StatusUnauthorized {
		t.Errorf("anonymous: %d", w.Code)
	}

	w := do("GET", "/api/v1/notifications/", sarahID)
	var page Page
	_ = json.Unmarshal(w.Body.Bytes(), &page)
	if w.Code != 200 || page.Unread != 1 || len(page.Results) != 1 || page.Results[0].Event != "referral_created" {
		t.Fatalf("list: %d %s", w.Code, w.Body)
	}
	id := page.Results[0].ID

	if w := do("POST", "/api/v1/notifications/"+id+"/read/", tiagoID); w.Code != http.StatusNotFound {
		t.Errorf("another user's notification: want 404, got %d", w.Code)
	}
	if w := do("POST", "/api/v1/notifications/"+id+"/read/", sarahID); w.Code != http.StatusNoContent {
		t.Errorf("mark read: %d", w.Code)
	}
	w = do("GET", "/api/v1/notifications/unread-count/", sarahID)
	if w.Code != 200 || w.Body.String() != `{"unread":0}` {
		t.Errorf("unread-count: %d %s", w.Code, w.Body)
	}
	if w := do("POST", "/api/v1/notifications/read-all/", tiagoID); w.Code != http.StatusNoContent {
		t.Errorf("read-all: %d", w.Code)
	}
}
