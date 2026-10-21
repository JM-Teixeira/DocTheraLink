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
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const demoPassword = "1234@1"

// TestMain waits until the target answers through its proxy/ingress without
// 5xx several times in a row before running the tests, so a just-finished
// rollout (old pod still draining behind the ingress) causes no spurious 502s.
func TestMain(m *testing.M) {
	if base := strings.TrimRight(os.Getenv("E2E_BASE_URL"), "/"); base != "" {
		const needed = 5 // consecutive healthy answers
		deadline := time.Now().Add(120 * time.Second)
		healthy := 0
		var lastErr error
		for healthy < needed {
			resp, err := http.Get(base + "/api/v1/auth/login/")
			if err == nil {
				resp.Body.Close()
			}
			if err == nil && resp.StatusCode < 500 {
				healthy++
			} else {
				healthy, lastErr = 0, err
			}
			if time.Now().After(deadline) {
				fmt.Fprintf(os.Stderr, "e2e: %s not stable after 120s (last err %v)\n", base, lastErr)
				os.Exit(1)
			}
			time.Sleep(time.Second)
		}
	}
	os.Exit(m.Run())
}

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

	// All demo staff work in the same clinic, so they share its patients
	// (cross-clinic isolation is covered by the integration tests).
	expect(t, "nurse sees same-clinic patient", nurse.call("GET", "/patient/"+pid+"/", nil, nil), 200)

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
	expect(t, "physio sees the referred patient", physio.call("GET", "/patient/"+pid+"/", nil, nil), 200)

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

// userIDOf returns the numeric id of a user via GET /users/getcurrentuser/.
func userIDOf(c *client) float64 {
	var me map[string]any
	c.call("GET", "/users/getcurrentuser/", nil, &me)
	id, _ := me["id"].(float64)
	return id
}

// TestHandOverAndReopen: a declined referral is re-opened to a colleague, who
// later hands it back; the status counts follow along.
func TestHandOverAndReopen(t *testing.T) {
	doctor, _ := login(t, "doctor")
	sarah, _ := login(t, "physiotherapist")
	tiago, _ := login(t, "physiotherapist2")
	sarahID, tiagoID := userIDOf(sarah), userIDOf(tiago)

	var p map[string]any
	expect(t, "patient", doctor.call("POST", "/patient/", map[string]any{"name": "Handover " + time.Now().Format("150405")}, &p), 201)
	var ref map[string]any
	expect(t, "referral to Sarah", doctor.call("POST", "/referrals/", map[string]any{
		"patient_id": p["id"], "therapist_id": sarahID, "diagnosis": "Tennis elbow", "reason": "Eccentric loading",
	}, &ref), 201)
	rid := ref["id"].(string)

	expect(t, "Sarah declines", sarah.call("POST", "/referrals/"+rid+"/decline/", map[string]string{"reason": "Full schedule"}, nil), 200)
	var reopened map[string]any
	expect(t, "doctor re-opens to Tiago", doctor.call("POST", "/referrals/"+rid+"/reopen/",
		map[string]any{"therapist_id": tiagoID, "reason": "Tiago has availability"}, &reopened), 200)
	if reopened["status"] != "pending" || reopened["therapist_id"] != tiagoID {
		t.Fatalf("re-opened: %v", reopened)
	}
	expect(t, "Tiago accepts", tiago.call("POST", "/referrals/"+rid+"/accept/", map[string]any{}, nil), 200)
	expect(t, "Sarah cannot act any more", sarah.call("POST", "/referrals/"+rid+"/start/", map[string]any{}, nil), 404)

	var handed map[string]any
	expect(t, "Tiago hands back to Sarah", tiago.call("POST", "/referrals/"+rid+"/reassign/",
		map[string]any{"therapist_id": sarahID, "reason": "Holiday cover ends"}, &handed), 200)
	if handed["status"] != "accepted" || handed["therapist_id"] != sarahID {
		t.Fatalf("handed over: %v", handed)
	}
	expect(t, "Sarah starts", sarah.call("POST", "/referrals/"+rid+"/start/", map[string]any{}, nil), 200)

	var counts map[string]float64
	expect(t, "counts", sarah.call("GET", "/referrals/counts/?search=Handover", nil, &counts), 200)
	if counts["in_progress"] < 1 {
		t.Errorf("counts should include the in-progress referral: %v", counts)
	}
}

// notificationFor finds the user's notification of a kind for a referral.
func notificationFor(c *client, referralID, event string) map[string]any {
	var page struct {
		Results []map[string]any `json:"results"`
	}
	c.call("GET", "/notifications/?limit=100", nil, &page)
	for _, n := range page.Results {
		if n["referral_id"] == referralID && n["event"] == event {
			return n
		}
	}
	return nil
}

// TestNotifications: a new open referral notifies the clinic's physios (not
// the doctor who created it); accepting notifies the doctor; read state works.
func TestNotifications(t *testing.T) {
	doctor, _ := login(t, "doctor")
	sarah, _ := login(t, "physiotherapist")
	tiago, _ := login(t, "physiotherapist2")

	var p, ref map[string]any
	expect(t, "patient", doctor.call("POST", "/patient/", map[string]any{"name": "Notify " + time.Now().Format("150405")}, &p), 201)
	expect(t, "open referral", doctor.call("POST", "/referrals/", map[string]any{
		"patient_id": p["id"], "diagnosis": "Neck pain", "reason": "Mobility",
	}, &ref), 201)
	rid := ref["id"].(string)

	for name, c := range map[string]*client{"sarah": sarah, "tiago": tiago} {
		n := notificationFor(c, rid, "referral_created")
		if n == nil || n["read_at"] != nil || n["patient_name"] != p["name"] {
			t.Fatalf("%s should have an unread referral_created notification, got %v", name, n)
		}
	}
	if notificationFor(doctor, rid, "referral_created") != nil {
		t.Error("the doctor must not be notified about their own referral")
	}

	expect(t, "accept", sarah.call("POST", "/referrals/"+rid+"/accept/", map[string]any{}, nil), 200)
	n := notificationFor(doctor, rid, "referral_accepted")
	if n == nil || n["actor_name"] != "Sarah Johnson" {
		t.Fatalf("doctor should be told Sarah accepted: %v", n)
	}
	expect(t, "mark read", doctor.call("POST", "/notifications/"+n["id"].(string)+"/read/", map[string]any{}, nil), 204)
	if n := notificationFor(doctor, rid, "referral_accepted"); n["read_at"] == nil {
		t.Error("read state not saved")
	}
	expect(t, "cannot read someone else's", tiago.call("POST", "/notifications/"+n["id"].(string)+"/read/", map[string]any{}, nil), 404)

	var count map[string]float64
	expect(t, "unread count", sarah.call("GET", "/notifications/unread-count/", nil, &count), 200)
	expect(t, "read all", sarah.call("POST", "/notifications/read-all/", map[string]any{}, nil), 204)
	sarah.call("GET", "/notifications/unread-count/", nil, &count)
	if count["unread"] != 0 {
		t.Errorf("unread after read-all: %v", count)
	}
}

// TestReceptionist covers the front-desk role: registers and edits patients,
// but has no access to referrals (clinical content).
func TestReceptionist(t *testing.T) {
	rec, _ := login(t, "receptionist")
	var p map[string]any
	expect(t, "receptionist registers patient", rec.call("POST", "/patient/", map[string]any{
		"name": "Front Desk " + time.Now().Format("150405"), "phone_number": "+351 920 000 000",
	}, &p), 201)
	expect(t, "receptionist edits patient", rec.call("PATCH", "/patient/"+p["id"].(string)+"/",
		map[string]any{"address": "Rua Augusta 1, Lisboa"}, nil), 200)
	expect(t, "receptionist cannot list referrals", rec.call("GET", "/referrals/", nil, nil), 403)
	expect(t, "receptionist cannot create referrals", rec.call("POST", "/referrals/", map[string]any{
		"patient_id": p["id"], "diagnosis": "x", "reason": "y",
	}, nil), 403)
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

// upload sends one file as multipart form field "file" and decodes the answer.
func (c *client) upload(path, name string, content []byte, out any) int {
	c.t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, _ := w.CreateFormFile("file", name)
	_, _ = part.Write(content)
	_ = w.Close()
	req, _ := http.NewRequest("POST", c.base+"/api/v1"+path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		c.t.Fatalf("upload %s: %v", path, err)
	}
	defer resp.Body.Close()
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

// download returns the status, body and headers of a raw GET.
func (c *client) download(path string) (int, []byte, http.Header) {
	c.t.Helper()
	req, _ := http.NewRequest("GET", c.base+"/api/v1"+path, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		c.t.Fatalf("download %s: %v", path, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b, resp.Header
}

// TestAttachments: the doctor attaches an MRI report, the therapist downloads
// it byte for byte; outsiders, disguised executables and deletes by the
// other participant are refused.
func TestAttachments(t *testing.T) {
	doctor, _ := login(t, "doctor")
	physio, _ := login(t, "physiotherapist")
	nurse, _ := login(t, "nurse")

	var p map[string]any
	expect(t, "patient", doctor.call("POST", "/patient/", map[string]any{"name": "Files " + time.Now().Format("150405")}, &p), 201)
	var ref map[string]any
	expect(t, "referral", doctor.call("POST", "/referrals/", map[string]any{
		"patient_id": p["id"], "therapist_id": userIDOf(physio), "diagnosis": "Rotator cuff tear", "reason": "Post-op rehab",
	}, &ref), 201)
	rid := ref["id"].(string)

	pdf := []byte("%PDF-1.7\n1 0 obj << >> endobj\ntrailer\n%%EOF\n")
	var a map[string]any
	expect(t, "doctor uploads PDF", doctor.upload("/referrals/"+rid+"/attachments/", "MRI shoulder.pdf", pdf, &a), 201)
	aid := a["id"].(string)
	if a["file_name"] != "MRI shoulder.pdf" || a["content_type"] != "application/pdf" {
		t.Fatalf("attachment: %v", a)
	}

	expect(t, "disguised executable", doctor.upload("/referrals/"+rid+"/attachments/", "scan.pdf",
		append([]byte("MZ\x90\x00"), make([]byte, 64)...), nil), 415)
	expect(t, "nurse cannot upload", nurse.upload("/referrals/"+rid+"/attachments/", "x.pdf", pdf, nil), 404)

	var list []map[string]any
	expect(t, "physio lists files", physio.call("GET", "/referrals/"+rid+"/attachments/", nil, &list), 200)
	if len(list) != 1 || list[0]["id"] != aid {
		t.Fatalf("list: %v", list)
	}

	code, body, h := physio.download("/attachments/" + aid + "/download/")
	expect(t, "physio downloads", code, 200)
	if !bytes.Equal(body, pdf) {
		t.Fatal("downloaded content differs")
	}
	if !strings.HasPrefix(h.Get("Content-Disposition"), "attachment") || h.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("download headers: %v", h)
	}
	code, _, _ = nurse.download("/attachments/" + aid + "/download/")
	expect(t, "nurse cannot download", code, 404)

	expect(t, "physio cannot delete doctor's file", physio.call("DELETE", "/attachments/"+aid+"/", nil, nil), 403)
	expect(t, "doctor deletes own file", doctor.call("DELETE", "/attachments/"+aid+"/", nil, nil), 204)
	code, _, _ = physio.download("/attachments/" + aid + "/download/")
	expect(t, "deleted file is gone", code, 404)
}

// TestCatalog: the referral form's ICD-10 search and templates, per language.
func TestCatalog(t *testing.T) {
	doctor, _ := login(t, "doctor")
	var codes []map[string]string
	expect(t, "icd search", doctor.call("GET", "/catalog/icd10/?search=lombalgia&lang=pt", nil, &codes), 200)
	if len(codes) == 0 || codes[0]["code"] != "M54.5" || codes[0]["description"] != "Lombalgia" {
		t.Fatalf("icd search: %v", codes)
	}
	var tpls []map[string]any
	expect(t, "templates", doctor.call("GET", "/catalog/referral-templates/?lang=en", nil, &tpls), 200)
	if len(tpls) == 0 || tpls[0]["name"] == "" || tpls[0]["icd_code"] == "" {
		t.Fatalf("templates: %v", tpls)
	}
	anon := &client{t: t, base: doctor.base}
	expect(t, "catalog needs login", anon.call("GET", "/catalog/icd10/?search=m54", nil, nil), 401)
}

// TestPatientIdentifiersAndDuplicates: NIF is validated and unique per
// clinic; a similar name with the same birth date needs confirmation.
func TestPatientIdentifiersAndDuplicates(t *testing.T) {
	rec, _ := login(t, "receptionist")
	// A fresh valid NIF per run: 8 digits from the clock + check digit
	base := fmt.Sprintf("2%07d", time.Now().UnixNano()%10_000_000)
	sum := 0
	for i := 0; i < 8; i++ {
		sum += int(base[i]-'0') * (9 - i)
	}
	check := 11 - sum%11
	if check >= 10 {
		check = 0
	}
	nif := fmt.Sprintf("%s%d", base, check)
	name := "Duplicate Check " + time.Now().Format("150405.000")

	expect(t, "invalid NIF", rec.call("POST", "/patient/", map[string]any{"name": name, "nif": "123456780"}, nil), 400)
	var p map[string]any
	expect(t, "register with NIF", rec.call("POST", "/patient/", map[string]any{
		"name": name, "date_of_birth": "1970-01-15", "nif": nif[:3] + " " + nif[3:6] + " " + nif[6:],
	}, &p), 201)
	if p["nif"] != nif {
		t.Fatalf("NIF stored normalised: %v", p["nif"])
	}

	var conflict struct {
		Error      string           `json:"error"`
		Blocking   bool             `json:"blocking"`
		Duplicates []map[string]any `json:"duplicates"`
	}
	expect(t, "same NIF", rec.call("POST", "/patient/", map[string]any{"name": "Someone Else", "nif": nif, "allow_duplicate": true}, &conflict), 409)
	if !conflict.Blocking || len(conflict.Duplicates) != 1 || conflict.Duplicates[0]["id"] != p["id"] {
		t.Fatalf("NIF conflict: %+v", conflict)
	}
	similar := map[string]any{"name": strings.ToLower(name), "date_of_birth": "1970-01-15"}
	expect(t, "similar name + birth date", rec.call("POST", "/patient/", similar, &conflict), 409)
	if conflict.Blocking || conflict.Duplicates[0]["reason"] != "name_dob" {
		t.Fatalf("soft duplicate: %+v", conflict)
	}
	similar["allow_duplicate"] = true
	expect(t, "confirmed different person", rec.call("POST", "/patient/", similar, nil), 201)
}
