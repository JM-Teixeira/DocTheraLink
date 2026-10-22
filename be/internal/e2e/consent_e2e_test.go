//go:build e2e

package e2e

import (
	"testing"
	"time"
)

// TestPatientConsents: the front desk records consents; a change adds a
// record and the newest is the current state; bad kinds are refused.
func TestPatientConsents(t *testing.T) {
	desk, _ := login(t, "receptionist")
	var p map[string]any
	expect(t, "patient", desk.call("POST", "/patient/", map[string]any{"name": "Consentimento " + time.Now().Format("150405.000")}, &p), 201)
	path := "/patient/" + p["id"].(string) + "/consents/"

	type view struct {
		Current map[string]struct {
			Granted bool   `json:"granted"`
			Method  string `json:"method"`
		} `json:"current"`
		History []map[string]any `json:"history"`
	}
	var v view
	expect(t, "empty", desk.call("GET", path, nil, &v), 200)
	if len(v.Current) != 0 || len(v.History) != 0 {
		t.Fatalf("new patient has consents: %+v", v)
	}
	expect(t, "give", desk.call("POST", path, map[string]any{"kind": "contact_sms", "granted": true, "method": "signed_paper"}, nil), 201)
	expect(t, "withdraw", desk.call("POST", path, map[string]any{"kind": "contact_sms", "granted": false, "method": "verbal", "note": "pediu"}, &v), 201)
	if c := v.Current["contact_sms"]; c.Granted || c.Method != "verbal" || len(v.History) != 2 {
		t.Errorf("after withdraw: %+v", v)
	}
	expect(t, "unknown kind", desk.call("POST", path, map[string]any{"kind": "marketing", "granted": true, "method": "verbal"}, nil), 400)
	expect(t, "granted is required", desk.call("POST", path, map[string]any{"kind": "photos", "method": "verbal"}, nil), 400)
	expect(t, "unknown patient", desk.call("GET", "/patient/00000000-0000-0000-0000-000000000000/consents/", nil, nil), 404)
}
