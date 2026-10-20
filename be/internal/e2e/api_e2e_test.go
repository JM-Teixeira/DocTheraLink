//go:build e2e

// Package e2e tests a *running* Klyniva deployment over HTTP, exactly as the
// browser does (through the frontend nginx proxy or the ingress).
//
//	E2E_BASE_URL=http://klyniva.localhost:8088 go test -tags e2e ./internal/e2e/
//
// The target must have demo users seeded (SEED_DEMO_USERS=true). Each run
// creates its own patients/referrals, so it can be repeated.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const demoPassword = "1234@1"

// client is a tiny JSON HTTP client bound to one user's access token.
type client struct {
	t     *testing.T
	base  string
	token string
}

func baseURL(t *testing.T) string {
	t.Helper()
	u := strings.TrimRight(os.Getenv("E2E_BASE_URL"), "/")
	if u == "" {
		t.Skip("E2E_BASE_URL not set")
	}
	return u
}

// call sends a request and decodes the JSON response into out (if non-nil).
func (c *client) call(method, path string, body, out any) int {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, c.base+"/api/v1"+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

// login returns a client authenticated as a demo user.
func login(t *testing.T, username string) (*client, map[string]string) {
	t.Helper()
	anon := &client{t: t, base: baseURL(t)}
	var tokens map[string]string
	if code := anon.call("POST", "/auth/login/", map[string]string{"username": username, "password": demoPassword}, &tokens); code != 200 {
		t.Fatalf("login %s: status %d (are demo users seeded?)", username, code)
	}
	return &client{t: t, base: anon.base, token: tokens["access"]}, tokens
}

func expect(t *testing.T, what string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: status %d, want %d", what, got, want)
	}
}

// TestAuthRequired checks that the API is reachable and rejects anonymous calls.
func TestAuthRequired(t *testing.T) {
	c := &client{t: t, base: baseURL(t)}
	expect(t, "unauthenticated list", c.call("GET", "/referrals/", nil, nil), 401)
	expect(t, "bad password", c.call("POST", "/auth/login/",
		map[string]string{"username": "doctor", "password": "wrong"}, nil), 401)
}

// TestReferralLifecycle walks the core product flow end to end:
// doctor registers a patient and refers them; a physiotherapist accepts,
// treats and closes the referral; both talk on the shared thread.
func TestReferralLifecycle(t *testing.T) {
	doctor, doctorTokens := login(t, "doctor")
	physio, _ := login(t, "physiotherapist")
	nurse, _ := login(t, "nurse")
	stamp := time.Now().Format("150405")

	var patient map[string]any
	expect(t, "create patient", doctor.call("POST", "/patient/", map[string]any{
		"name": "E2E Patient " + stamp, "date_of_birth": "1980-02-03", "gender": 2,
		"phone_number": "+351 910 000 000",
	}, &patient), 201)
	pid := patient["id"].(string)

	expect(t, "nurse cannot see unrelated patient", nurse.call("GET", "/patient/"+pid+"/", nil, nil), 404)
	expect(t, "physio cannot see patient before referral", physio.call("GET", "/patient/"+pid+"/", nil, nil), 404)

	var ref map[string]any
	expect(t, "create referral", doctor.call("POST", "/referrals/", map[string]any{
		"patient_id": pid, "diagnosis": "Lumbar disc herniation", "icd_code": "M51.2",
		"reason": "Core stabilisation", "priority": "urgent", "recommended_sessions": 8,
	}, &ref), 201)
	rid := ref["id"].(string)
	if ref["status"] != "pending" {
		t.Fatalf("new referral status %v", ref["status"])
	}

	expect(t, "physio cannot create referrals", physio.call("POST", "/referrals/", map[string]any{
		"patient_id": pid, "diagnosis": "x", "reason": "y",
	}, nil), 403)

	var inbox struct {
		Results []map[string]any `json:"results"`
	}
	expect(t, "physio inbox", physio.call("GET", "/referrals/?status=pending&limit=100", nil, &inbox), 200)
	found := false
	for _, r := range inbox.Results {
		found = found || r["id"] == rid
	}
	if !found {
		t.Fatal("new referral not in physiotherapist inbox")
	}

	for _, step := range []struct{ action, want string }{{"accept", "accepted"}, {"start", "in_progress"}} {
		var out map[string]any
		expect(t, step.action, physio.call("POST", "/referrals/"+rid+"/"+step.action+"/", map[string]any{}, &out), 200)
		if out["status"] != step.want {
			t.Fatalf("%s: status %v, want %s", step.action, out["status"], step.want)
		}
	}
	expect(t, "accept twice conflicts", physio.call("POST", "/referrals/"+rid+"/accept/", map[string]any{}, nil), 409)
	expect(t, "physio sees patient after accepting", physio.call("GET", "/patient/"+pid+"/", nil, nil), 200)

	expect(t, "doctor asks on thread", doctor.call("POST", "/referrals/"+rid+"/notes/",
		map[string]string{"kind": "question", "body": "Any radicular symptoms?"}, nil), 201)
	expect(t, "physio replies", physio.call("POST", "/referrals/"+rid+"/notes/",
		map[string]string{"kind": "reply", "body": "None, improving"}, nil), 201)
	expect(t, "nurse cannot post", nurse.call("POST", "/referrals/"+rid+"/notes/",
		map[string]string{"body": "hi"}, nil), 404)

	var done map[string]any
	expect(t, "complete", physio.call("POST", "/referrals/"+rid+"/complete/",
		map[string]string{"reason": "Pain 2/10, back to work"}, &done), 200)
	if done["status"] != "completed" || done["outcome_summary"] != "Pain 2/10, back to work" {
		t.Fatalf("completed referral: %v", done)
	}
	expect(t, "cannot cancel completed", doctor.call("POST", "/referrals/"+rid+"/cancel/",
		map[string]string{"reason": "late"}, nil), 409)

	var notes []map[string]any
	expect(t, "thread", doctor.call("GET", "/referrals/"+rid+"/notes/", nil, &notes), 200)
	kinds := make([]string, len(notes))
	for i, n := range notes {
		kinds[i] = fmt.Sprint(n["kind"])
	}
	if got := strings.Join(kinds, ","); got != "system,system,question,reply,system" {
		t.Fatalf("thread kinds = %s", got)
	}

	// A refresh token must never work as an access token.
	refreshClient := &client{t: t, base: doctor.base, token: doctorTokens["refresh"]}
	expect(t, "refresh token as access", refreshClient.call("GET", "/referrals/", nil, nil), 401)
}

// TestClockInOut covers the staff time log (it used to always return 401).
func TestClockInOut(t *testing.T) {
	staff, _ := login(t, "staff")
	var status map[string]any
	staff.call("GET", "/timelog/status", nil, &status)
	if status["clocked_in"] == true {
		staff.call("POST", "/timelog/clock-out", map[string]string{}, nil)
	}
	expect(t, "clock in", staff.call("POST", "/timelog/clock-in", map[string]string{"notes": "e2e"}, nil), 201)
	expect(t, "clock out", staff.call("POST", "/timelog/clock-out", map[string]string{}, nil), 200)
}
