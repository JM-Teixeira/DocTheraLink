package referral

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/config"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
)

// testServer wires the real JWT middleware and routes over a fake repository.
type testServer struct {
	router *gin.Engine
	jwt    *auth.JWTManager
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc, _ := newTestService(t)
	h := NewHandler(svc)

	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-secret", Expiry: time.Hour, RefreshExpiry: 2 * time.Hour}}
	jwt := auth.NewJWTManager(cfg)

	r := gin.New()
	g := r.Group("/api/v1", auth.AuthMiddleware(jwt)).Group("/referrals")
	g.GET("/", h.List)
	g.POST("/", h.Create)
	g.GET("/:id/", h.Get)
	g.POST("/:id/accept/", h.Accept)
	g.POST("/:id/decline/", h.Decline)
	g.GET("/:id/notes/", h.ListNotes)
	g.POST("/:id/notes/", h.AddNote)
	return &testServer{router: r, jwt: jwt}
}

// tokens returns access and refresh tokens for the actor.
func (s *testServer) tokens(t *testing.T, a Actor) *models.LoginResponse {
	t.Helper()
	pair, err := s.jwt.GenerateTokenPair(&models.User{ID: a.ID, Username: "u", UserType: a.Role})
	if err != nil {
		t.Fatal(err)
	}
	return pair
}

// do performs a request with the given bearer token and JSON body.
func (s *testServer) do(t *testing.T, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

func TestHTTPCreateAndAccept(t *testing.T) {
	s := newTestServer(t)
	doctorTok := s.tokens(t, doctor).Access
	physioTok := s.tokens(t, physio).Access

	w := s.do(t, http.MethodPost, "/api/v1/referrals/", doctorTok, validCreate())
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	var created Referral
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	w = s.do(t, http.MethodPost, "/api/v1/referrals/"+created.ID+"/accept/", physioTok, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("accept: %d %s", w.Code, w.Body)
	}

	// Accepting twice is a conflict, not a server error.
	w = s.do(t, http.MethodPost, "/api/v1/referrals/"+created.ID+"/accept/", physioTok, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("second accept: want 409, got %d", w.Code)
	}

	w = s.do(t, http.MethodGet, "/api/v1/referrals/", physioTok, nil)
	var page models.PaginatedResponse
	_ = json.Unmarshal(w.Body.Bytes(), &page)
	if w.Code != http.StatusOK || page.Count != 1 {
		t.Fatalf("list: %d count=%d", w.Code, page.Count)
	}
}

func TestHTTPErrorMapping(t *testing.T) {
	s := newTestServer(t)
	doctorTok := s.tokens(t, doctor).Access
	physioTok := s.tokens(t, physio).Access

	cases := []struct {
		name   string
		method string
		path   string
		token  string
		body   any
		want   int
	}{
		{"no token", http.MethodGet, "/api/v1/referrals/", "", nil, http.StatusUnauthorized},
		{"refresh token is not an access token", http.MethodGet, "/api/v1/referrals/", s.tokens(t, doctor).Refresh, nil, http.StatusUnauthorized},
		{"binding error", http.MethodPost, "/api/v1/referrals/", doctorTok, map[string]string{"patient_id": "not-a-uuid"}, http.StatusBadRequest},
		{"physio cannot create", http.MethodPost, "/api/v1/referrals/", physioTok, validCreate(), http.StatusForbidden},
		{"unknown referral", http.MethodGet, "/api/v1/referrals/" + patientID + "/", doctorTok, nil, http.StatusNotFound},
		{"decline needs reason", http.MethodPost, "/api/v1/referrals/x/decline/", physioTok, map[string]string{}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := s.do(t, tc.method, tc.path, tc.token, tc.body)
			if w.Code != tc.want {
				t.Fatalf("want %d, got %d: %s", tc.want, w.Code, w.Body)
			}
		})
	}
}

func TestHTTPNotesThread(t *testing.T) {
	s := newTestServer(t)
	doctorTok := s.tokens(t, doctor).Access

	req := validCreate()
	req.TherapistID = ptrUint(physio.ID)
	w := s.do(t, http.MethodPost, "/api/v1/referrals/", doctorTok, req)
	var created Referral
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	w = s.do(t, http.MethodPost, "/api/v1/referrals/"+created.ID+"/notes/", doctorTok,
		NoteRequest{Kind: NoteQuestion, Body: "Weight bearing as tolerated"})
	if w.Code != http.StatusCreated {
		t.Fatalf("add note: %d %s", w.Code, w.Body)
	}

	w = s.do(t, http.MethodGet, "/api/v1/referrals/"+created.ID+"/notes/", s.tokens(t, physio).Access, nil)
	var notes []Note
	_ = json.Unmarshal(w.Body.Bytes(), &notes)
	if w.Code != http.StatusOK || len(notes) != 1 {
		t.Fatalf("list notes: %d len=%d", w.Code, len(notes))
	}
}
