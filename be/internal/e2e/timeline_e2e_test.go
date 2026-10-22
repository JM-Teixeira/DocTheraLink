//go:build e2e

package e2e

import (
	"testing"
	"time"
)

// TestPatientExtrasAndTimeline: the new registration fields are validated and
// normalised; the timeline gathers the referral, its messages and sessions,
// and each role sees only what it may (no referrals for the front desk).
func TestPatientExtrasAndTimeline(t *testing.T) {
	doctor, _ := login(t, "de.john.smith")
	sarah, _ := login(t, "de.sarah.johnson")
	coord, _ := login(t, "de.marta.ribeiro")
	desk, _ := login(t, "de.rita.lopes")
	sarahID := userIDOf(sarah)

	expect(t, "bad postal code", desk.call("POST", "/patient/", map[string]any{"name": "Postal X", "postal_code": "12"}, nil), 400)
	expect(t, "bad allergy severity", desk.call("POST", "/patient/", map[string]any{"name": "Postal X", "allergy_list": []any{map[string]any{"substance": "Pó", "severity": "x"}}}, nil), 400)
	expect(t, "bad relation", desk.call("POST", "/patient/", map[string]any{"name": "Postal X", "emergency_contact_relation": "boss"}, nil), 400)
	var p map[string]any
	expect(t, "patient", desk.call("POST", "/patient/", map[string]any{
		"name": "Linha Tempo " + time.Now().Format("150405.000"), "postal_code": "4000123", "city": "Porto",
		"email": "linha@example.pt", "emergency_contact_name": "Rui", "emergency_contact_relation": "spouse",
		"insurer": "ADSE", "insurance_number": "123",
		"allergy_list":   []any{map[string]any{"substance": " Látex ", "severity": "severe"}, map[string]any{"substance": ""}},
		"condition_list": []any{map[string]any{"name": "Asma", "since": 2010}},
	}, &p), 201)
	if p["postal_code"] != "4000-123" || p["emergency_contact_relation"] != "spouse" || p["insurer"] != "ADSE" {
		t.Fatalf("extras: %v", p)
	}
	if al, _ := p["allergy_list"].([]any); len(al) != 1 || al[0].(map[string]any)["substance"] != "Látex" {
		t.Errorf("allergy list: %v", p["allergy_list"])
	}
	pid := p["id"].(string)
	var edited map[string]any
	expect(t, "edit city, clear e-mail", desk.call("PATCH", "/patient/"+pid+"/", map[string]any{"city": "Braga", "email": ""}, &edited), 200)
	if edited["city"] != "Braga" || edited["email"] != nil {
		t.Errorf("edit: city %v email %v", edited["city"], edited["email"])
	}

	var ref map[string]any
	expect(t, "referral", doctor.call("POST", "/referrals/", map[string]any{
		"patient_id": pid, "diagnosis": "Cervicalgia", "reason": "Mobilidade", "therapist_id": sarahID,
	}, &ref), 201)
	rid := ref["id"].(string)
	expect(t, "accept", sarah.call("POST", "/referrals/"+rid+"/accept/", map[string]any{}, nil), 200)
	expect(t, "progress", sarah.call("POST", "/referrals/"+rid+"/notes/", map[string]any{"kind": "progress", "body": "Melhor", "pain": 4}, nil), 201)

	// A session on a free day far ahead
	m := time.Date(2071, 1, 5, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 7*int(time.Now().Unix()%300))
	for freeDay(t, coord, sarahID, m) != m.Format("2006-01-02") {
		m = m.AddDate(0, 0, 7)
	}
	date := m.Format("2006-01-02")
	var divisions []map[string]any
	expect(t, "divisions", coord.call("GET", "/schedule/divisions/", nil, &divisions), 200)
	var sh, ap map[string]any
	expect(t, "shift", coord.call("POST", "/schedule/shifts/", map[string]any{"date": date, "start_hour": 9, "end_hour": 11, "therapist_id": sarahID, "division_id": divisions[0]["id"]}, &sh), 201)
	defer coord.call("DELETE", "/schedule/shifts/"+sh["id"].(string)+"/", nil, nil)
	expect(t, "book", coord.call("POST", "/schedule/appointments/", map[string]any{"date": date, "hour": 9, "therapist_id": sarahID, "patient_id": pid}, &ap), 201)
	defer coord.call("DELETE", "/schedule/appointments/"+ap["id"].(string)+"/", nil, nil)

	type entry struct {
		Kind   string `json:"kind"`
		Status string `json:"status"`
		Pain   *int   `json:"pain"`
		Date   string `json:"date"`
		Who    string `json:"who"`
	}
	var tl struct {
		Upcoming []entry `json:"upcoming"`
		History  []entry `json:"history"`
	}
	kinds := func(es []entry) map[string]int {
		out := map[string]int{}
		for _, e := range es {
			out[e.Kind+":"+e.Status]++
		}
		return out
	}
	expect(t, "physio timeline", sarah.call("GET", "/patient/"+pid+"/timeline/", nil, &tl), 200)
	k := kinds(tl.History)
	if k["referral:created"] != 1 || k["referral:accepted"] != 1 || k["note:progress"] != 1 || k["registered:"] != 1 {
		t.Errorf("history: %v", k)
	}
	if len(tl.Upcoming) != 1 || tl.Upcoming[0].Date != date || tl.Upcoming[0].Who == "" {
		t.Errorf("upcoming: %+v", tl.Upcoming)
	}
	if tl.History[0].Kind != "note" || *tl.History[0].Pain != 4 {
		t.Errorf("newest first: %+v", tl.History[0])
	}

	expect(t, "front desk timeline", desk.call("GET", "/patient/"+pid+"/timeline/", nil, &tl), 200)
	if k := kinds(tl.History); k["referral:created"] != 0 || k["note:progress"] != 0 || len(tl.Upcoming) != 1 {
		t.Errorf("front desk sees referrals or misses sessions: %v %d", k, len(tl.Upcoming))
	}
	expect(t, "unknown patient", desk.call("GET", "/patient/00000000-0000-0000-0000-000000000000/timeline/", nil, nil), 404)
}
