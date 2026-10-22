//go:build e2e

package e2e

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

type page struct {
	Count   int              `json:"count"`
	Results []map[string]any `json:"results"`
}

// TestPeopleAndEmployment: the people list filters by status (people who
// left stay on record but cannot sign in); skills are linked and unlinked;
// only a manager of the person's clinic changes their status.
func TestPeopleAndEmployment(t *testing.T) {
	coord, _ := login(t, "cm.coordenacao")
	physio, _ := login(t, "cm.fisio3")
	otherCoord, _ := login(t, "coordinator")

	var p page
	expect(t, "people who left", coord.call("GET", "/users/?employment_status=left", nil, &p), 200)
	found := false
	for _, u := range p.Results {
		found = found || u["username"] == "cm.fisio7"
		if u["employment_status"] != "left" {
			t.Errorf("filter leaked %v", u)
		}
	}
	if !found {
		t.Fatalf("cm.fisio7 should be listed as left: %v", p.Results)
	}
	expect(t, "search by full name", coord.call("GET", "/users/?search="+url.QueryEscape("Rui Nogueira"), nil, &p), 200)
	if p.Count != 1 {
		t.Errorf("full-name search: %d", p.Count)
	}
	expect(t, "other clinic's people are hidden", otherCoord.call("GET", "/users/?search=cm.fisio", nil, &p), 200)
	if p.Count != 0 {
		t.Errorf("coordinator of another clinic sees %d Climutua people", p.Count)
	}

	anon := &client{t: t, base: baseURL(t)}
	if code := anon.call("POST", "/auth/login/", map[string]string{"username": "cm.fisio7", "password": demoPassword}, nil); code == 200 {
		t.Error("someone who left must not sign in")
	}

	status := map[string]any{"status": "on_leave", "note": "e2e"}
	expect(t, "physio cannot change a colleague", physio.call("PATCH", "/users/cm.fisio2/employment/", status, nil), 403)
	expect(t, "other clinic cannot change", otherCoord.call("PATCH", "/users/cm.fisio2/employment/", status, nil), 403)
	expect(t, "nobody changes their own", coord.call("PATCH", "/users/cm.coordenacao/employment/", status, nil), 400)
	var u map[string]any
	expect(t, "coordinator sets on leave", coord.call("PATCH", "/users/cm.fisio2/employment/", status, &u), 200)
	if u["employment_status"] != "on_leave" {
		t.Fatalf("status: %v", u["employment_status"])
	}
	expect(t, "back to active", coord.call("PATCH", "/users/cm.fisio2/employment/", map[string]any{"status": "active"}, nil), 200)

	// Skills: catalogue, link (idempotent), unlink
	var catalogue page
	expect(t, "catalogue", physio.call("GET", "/skill/", nil, &catalogue), 200)
	if catalogue.Count < 12 {
		t.Fatalf("skills catalogue: %d", catalogue.Count)
	}
	var mine page
	expect(t, "own skills", physio.call("GET", "/users/cm.fisio3/skill/", nil, &mine), 200)
	have := map[string]bool{}
	for _, s := range mine.Results {
		have[s["skill_object"].(map[string]any)["id"].(string)] = true
	}
	var pick string
	for _, s := range catalogue.Results {
		if !have[s["id"].(string)] {
			pick = s["id"].(string)
			break
		}
	}
	if pick == "" {
		t.Fatal("no skill left to add")
	}
	var link map[string]any
	expect(t, "add own skill", physio.call("POST", "/users/cm.fisio3/skill/", map[string]any{"skill": pick}, &link), 201)
	expect(t, "cannot edit a colleague's skills", physio.call("POST", "/users/cm.fisio2/skill/", map[string]any{"skill": pick}, nil), 403)
	expect(t, "remove", coord.call("DELETE", "/users/cm.fisio3/skill/"+link["id"].(string)+"/", nil, nil), 204)
	expect(t, "own skills again", physio.call("GET", "/users/cm.fisio3/skill/", nil, &mine), 200)
	if mine.Count != len(have) {
		t.Errorf("skills after remove: %d, want %d", mine.Count, len(have))
	}
}

// TestPatientFiltersAndStats: patients are found by NIF digits; the stats
// count what the list filters return.
func TestPatientFiltersAndStats(t *testing.T) {
	coord, _ := login(t, "cm.coordenacao")

	var all page
	expect(t, "list", coord.call("GET", "/patient/?limit=1", nil, &all), 200)
	var stats struct {
		Total      int64            `json:"total"`
		CareLevels map[string]int64 `json:"care_levels"`
		AgeBands   map[string]int64 `json:"age_bands"`
	}
	expect(t, "stats", coord.call("GET", "/patient/stats/", nil, &stats), 200)
	if stats.Total != int64(all.Count) {
		t.Fatalf("stats total %d, list %d", stats.Total, all.Count)
	}
	var levels int64
	for _, n := range stats.CareLevels {
		levels += n
	}
	if levels == 0 {
		t.Error("Climutua patients should have care levels")
	}

	var filtered page
	q := "care_level=4&care_level=5&age_min=65"
	expect(t, "filtered list", coord.call("GET", "/patient/?limit=200&"+q, nil, &filtered), 200)
	expect(t, "filtered stats", coord.call("GET", "/patient/stats/?"+q, nil, &stats), 200)
	if stats.Total != int64(filtered.Count) || filtered.Count >= all.Count {
		t.Fatalf("filtered: stats %d, list %d of %d", stats.Total, filtered.Count, all.Count)
	}
	for _, p := range filtered.Results {
		if lvl := p["care_level"].(float64); lvl < 4 {
			t.Errorf("care level %v leaked into the filter", lvl)
		}
	}

	// Search by NIF: the first digits of a seeded patient's NIF find them
	var withNIF page
	expect(t, "list", coord.call("GET", "/patient/?limit=50", nil, &withNIF), 200)
	for _, p := range withNIF.Results {
		nif, _ := p["nif"].(string)
		if len(nif) != 9 {
			continue
		}
		var hit page
		expect(t, "search NIF", coord.call("GET", "/patient/?search="+nif, nil, &hit), 200)
		if hit.Count != 1 || hit.Results[0]["id"] != p["id"] {
			t.Errorf("NIF %s: %d hits", nif, hit.Count)
		}
		return
	}
	t.Error("no Climutua patient with a NIF")
}

// TestReferralDoctorFilter: the referral list narrows to one referring doctor.
func TestReferralDoctorFilter(t *testing.T) {
	coord, _ := login(t, "cm.coordenacao")
	var refs page
	expect(t, "all", coord.call("GET", "/referrals/?limit=100", nil, &refs), 200)
	if refs.Count == 0 {
		t.Fatal("Climutua has referrals")
	}
	doctor := refs.Results[0]["doctor_id"].(float64)
	var byDoctor page
	expect(t, "by doctor", coord.call("GET", fmt.Sprintf("/referrals/?limit=100&doctor=%d", int(doctor)), nil, &byDoctor), 200)
	if byDoctor.Count == 0 || byDoctor.Count > refs.Count {
		t.Fatalf("by doctor: %d of %d", byDoctor.Count, refs.Count)
	}
	for _, r := range byDoctor.Results {
		if r["doctor_id"].(float64) != doctor {
			t.Errorf("other doctor's referral: %v", r["doctor_id"])
		}
	}
}

// TestHolidayCover: a physio's absence moves their sessions to the cover
// physio at the same hour, with a handover note the cover can read; the
// absent physio cannot be booked meanwhile.
func TestHolidayCover(t *testing.T) {
	coord, _ := login(t, "coordinator")
	rec, _ := login(t, "receptionist")
	away, _ := login(t, "physiotherapist")
	cover, _ := login(t, "physiotherapist2")
	awayID, coverID := userIDOf(away), userIDOf(cover)

	// A day of its own per run, within the handover window (3 months)
	date := time.Now().AddDate(0, 0, 40+int(time.Now().Unix()%40)).Format("2006-01-02")

	var divisions []map[string]any
	expect(t, "divisions", coord.call("GET", "/schedule/divisions/", nil, &divisions), 200)
	var gym string
	for _, d := range divisions {
		if d["capacity"].(float64) >= 2 {
			gym = d["id"].(string)
		}
	}
	if gym == "" {
		t.Fatal("expected a division with room for two")
	}
	var shiftA, shiftB map[string]any
	expect(t, "away's shift", coord.call("POST", "/schedule/shifts/", map[string]any{
		"date": date, "start_hour": 18, "end_hour": 19, "therapist_id": awayID, "division_id": gym,
	}, &shiftA), 201)
	expect(t, "cover's shift", coord.call("POST", "/schedule/shifts/", map[string]any{
		"date": date, "start_hour": 18, "end_hour": 19, "therapist_id": coverID, "division_id": gym,
	}, &shiftB), 201)

	var p map[string]any
	expect(t, "patient", rec.call("POST", "/patient/", map[string]any{"name": "Cover " + time.Now().Format("150405.000")}, &p), 201)
	pid := p["id"].(string)
	var appt map[string]any
	expect(t, "book with away", coord.call("POST", "/schedule/appointments/", map[string]any{
		"date": date, "hour": 18, "therapist_id": awayID, "patient_id": pid,
	}, &appt), 201)

	body := map[string]any{
		"therapist_id": awayID, "cover_id": coverID, "from": date, "to": date, "reason": "vacation",
		"handover":      "Plano geral",
		"notes":         []map[string]any{{"patient_id": pid, "note": "Joelho: progredir carga"}},
		"move_patients": true, "dry_run": true,
	}
	expect(t, "physio cannot plan absences", away.call("POST", "/schedule/absences/", body, nil), 403)
	var preview struct {
		Sessions []map[string]any `json:"sessions"`
	}
	expect(t, "preview", coord.call("POST", "/schedule/absences/", body, &preview), 201)
	if len(preview.Sessions) != 1 || preview.Sessions[0]["moved"] != true {
		t.Fatalf("preview: %+v", preview)
	}
	body["dry_run"] = false
	var res struct {
		Absence map[string]any `json:"absence"`
	}
	expect(t, "create absence", coord.call("POST", "/schedule/absences/", body, &res), 201)
	absenceID := res.Absence["id"].(string)
	defer coord.call("DELETE", "/schedule/absences/"+absenceID+"/", nil, nil)

	var board struct {
		Therapists []struct {
			ID    float64 `json:"id"`
			Hours []struct {
				Appointments []map[string]any `json:"appointments"`
			} `json:"hours"`
		} `json:"therapists"`
		Absences []map[string]any `json:"absences"`
	}
	expect(t, "board", cover.call("GET", "/schedule/board/?date="+date, nil, &board), 200)
	moved := false
	for _, row := range board.Therapists {
		for _, h := range row.Hours {
			for _, a := range h.Appointments {
				if a["patient_id"] == pid {
					moved = row.ID == coverID
				}
			}
		}
	}
	if !moved || len(board.Absences) == 0 {
		t.Fatalf("session not with the cover (absences %d)", len(board.Absences))
	}
	var handover []map[string]any
	expect(t, "cover reads handover", cover.call("GET", "/schedule/handover/?patient="+pid, nil, &handover), 200)
	if len(handover) != 1 || handover[0]["note"] != "Joelho: progredir carga" {
		t.Fatalf("handover: %v", handover)
	}
	expect(t, "absent physio cannot be booked", coord.call("POST", "/schedule/appointments/", map[string]any{
		"date": date, "hour": 18, "therapist_id": awayID, "patient_id": pid,
	}, nil), 409)

	// Clean up: the session, then the cover's shift (the away shift was removed)
	for _, row := range board.Therapists {
		for _, h := range row.Hours {
			for _, a := range h.Appointments {
				if a["patient_id"] == pid {
					coord.call("DELETE", "/schedule/appointments/"+a["id"].(string)+"/", nil, nil)
				}
			}
		}
	}
	coord.call("DELETE", "/schedule/shifts/"+shiftB["id"].(string)+"/", nil, nil)
}

// TestLeftUserLosesSession: a token issued before someone is marked as having
// left stops working right away, and works again when they come back.
func TestLeftUserLosesSession(t *testing.T) {
	coord, _ := login(t, "cm.coordenacao")
	tiago, _ := login(t, "cm.fisio6")
	expect(t, "signed in", tiago.call("GET", "/users/getcurrentuser/", nil, nil), 200)

	expect(t, "mark as left", coord.call("PATCH", "/users/cm.fisio6/employment/", map[string]any{"status": "left", "note": "e2e"}, nil), 200)
	defer coord.call("PATCH", "/users/cm.fisio6/employment/", map[string]any{"status": "active"}, nil)
	expect(t, "old token refused", tiago.call("GET", "/users/getcurrentuser/", nil, nil), 401)

	expect(t, "back", coord.call("PATCH", "/users/cm.fisio6/employment/", map[string]any{"status": "active"}, nil), 200)
	expect(t, "token works again", tiago.call("GET", "/users/getcurrentuser/", nil, nil), 200)
}
