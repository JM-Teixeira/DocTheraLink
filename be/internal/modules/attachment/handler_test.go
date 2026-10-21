package attachment

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
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

func TestHTTPAttachments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc, _, _ := newSvc(t, 1<<10)
	jwt := auth.NewJWTManager(&config.Config{JWT: config.JWTConfig{Secret: "s", Expiry: time.Hour, RefreshExpiry: time.Hour}})
	r := gin.New()
	mount(r.Group("/api/v1", auth.AuthMiddleware(jwt)), NewHandler(svc))

	tok := func(a referral.Actor) string {
		p, _ := jwt.GenerateTokenPair(&models.User{ID: a.ID, Username: "u", UserType: a.Role})
		return "Bearer " + p.Access
	}
	upload := func(a referral.Actor, name string, content []byte) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, _ := mw.CreateFormFile("file", name)
		fw.Write(content)
		mw.Close()
		req := httptest.NewRequest("POST", "/api/v1/referrals/"+rid+"/attachments/", &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("Authorization", tok(a))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	get := func(a referral.Actor, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", tok(a))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	w := upload(doctor, "Relatório RM.pdf", pdf)
	if w.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", w.Code, w.Body)
	}
	var a Attachment
	_ = json.Unmarshal(w.Body.Bytes(), &a)

	w = get(physio, "/api/v1/attachments/"+a.ID+"/download/")
	if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), pdf) {
		t.Fatalf("download: %d", w.Code)
	}
	for h, want := range map[string]string{
		"Content-Type":           "application/pdf",
		"Content-Disposition":    "attachment; filename*=UTF-8''Relat%C3%B3rio%20RM.pdf",
		"X-Content-Type-Options": "nosniff",
		"Cache-Control":          "no-store",
	} {
		if got := w.Header().Get(h); got != want {
			t.Errorf("%s = %q, want %q", h, got, want)
		}
	}

	if w := get(nurse, "/api/v1/attachments/"+a.ID+"/download/"); w.Code != http.StatusNotFound {
		t.Errorf("nurse download: want 404, got %d", w.Code)
	}
	if w := upload(doctor, "virus.pdf", []byte("MZ\x90\x00 not a pdf at all")); w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("spoofed type: want 415, got %d", w.Code)
	}
	if w := upload(doctor, "big.pdf", append(pdf, bytes.Repeat([]byte("x"), 2<<10)...)); w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("too large: want 413, got %d", w.Code)
	}
	if w := get(physio, "/api/v1/referrals/"+rid+"/attachments/"); w.Code != 200 {
		t.Errorf("list: %d", w.Code)
	}
}
