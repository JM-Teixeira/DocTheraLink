//go:build e2e

package e2e

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

// TestPatientMetrics: filter by physio and diagnosis (clinical, so not for
// the front desk), monthly registrations, and the CSV export for
// coordinators only.
func TestPatientMetrics(t *testing.T) {
	doctor, _ := login(t, "doctor")
	sarah, _ := login(t, "physiotherapist")
	coord, _ := login(t, "coordinator")
	desk, _ := login(t, "receptionist")
	sarahID := userIDOf(sarah)

	diag := "Metatarsalgia " + time.Now().Format("150405.000")
	var p, ref map[string]any
	expect(t, "patient", desk.call("POST", "/patient/", map[string]any{"name": "Metricas " + time.Now().Format("150405.000")}, &p), 201)
	expect(t, "referral", doctor.call("POST", "/referrals/", map[string]any{
		"patient_id": p["id"], "diagnosis": diag, "reason": "Marcha", "therapist_id": sarahID,
	}, &ref), 201)

	var page struct {
		Count   int              `json:"count"`
		Results []map[string]any `json:"results"`
	}
	expect(t, "by diagnosis", sarah.call("GET", "/patient/?diagnosis="+urlq(diag), nil, &page), 200)
	if page.Count != 1 || page.Results[0]["id"] != p["id"] {
		t.Errorf("diagnosis filter: %d %v", page.Count, page.Results)
	}
	expect(t, "front desk cannot filter by diagnosis", desk.call("GET", "/patient/?diagnosis=x", nil, nil), 403)
	expect(t, "by physio", coord.call("GET", fmt.Sprintf("/patient/?physio=%v&limit=100&search=%s", sarahID, urlq(p["name"].(string))), nil, &page), 200)
	if page.Count != 1 {
		t.Errorf("physio filter: %d", page.Count)
	}

	var st struct {
		Monthly []struct {
			Month string `json:"month"`
			Count int    `json:"count"`
		} `json:"monthly"`
	}
	expect(t, "stats", coord.call("GET", "/patient/stats/", nil, &st), 200)
	if len(st.Monthly) != 12 || st.Monthly[11].Month != time.Now().Format("2006-01") || st.Monthly[11].Count < 1 {
		t.Errorf("monthly: %+v", st.Monthly)
	}

	expect(t, "physio cannot export", sarah.call("GET", "/patient/export/", nil, nil), 403)
	expect(t, "coordinator exports", coord.call("GET", "/patient/export/?search="+urlq(p["name"].(string)), nil, nil), 200)
}

func urlq(s string) string { return url.QueryEscape(s) }
