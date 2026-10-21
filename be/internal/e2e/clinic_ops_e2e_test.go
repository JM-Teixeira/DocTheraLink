//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

// TestInventory: staff record use and restocks of the clinic's consumables;
// stock never goes negative; only coordinators/admins remove items.
func TestInventory(t *testing.T) {
	nurse, _ := login(t, "nurse")
	physio, _ := login(t, "physiotherapist")
	coord, _ := login(t, "coordinator")

	var seeded []map[string]any
	expect(t, "list", nurse.call("GET", "/inventory/items/", nil, &seeded), 200)
	if len(seeded) == 0 {
		t.Fatal("demo clinic should have stock")
	}

	var item map[string]any
	name := "E2E gauze " + time.Now().Format("150405.000")
	expect(t, "create", nurse.call("POST", "/inventory/items/", map[string]any{
		"name": name, "category": "Treatment", "unit": "packs", "quantity": 3, "min_quantity": 2,
	}, &item), 201)
	id := item["id"].(string)

	var after map[string]any
	expect(t, "use 2", physio.call("POST", "/inventory/items/"+id+"/movements/", map[string]any{"kind": "use", "quantity": 2}, &after), 200)
	if after["quantity"] != 1.0 || after["status"] != "low" {
		t.Fatalf("after use: %v", after)
	}
	expect(t, "use more than in stock", physio.call("POST", "/inventory/items/"+id+"/movements/", map[string]any{"kind": "use", "quantity": 5}, nil), 409)
	expect(t, "restock", nurse.call("POST", "/inventory/items/"+id+"/movements/", map[string]any{"kind": "restock", "quantity": 10, "note": "order #42"}, &after), 200)
	if after["quantity"] != 11.0 || after["status"] != "ok" {
		t.Fatalf("after restock: %v", after)
	}
	var hist []map[string]any
	expect(t, "history", coord.call("GET", "/inventory/items/"+id+"/movements/", nil, &hist), 200)
	if len(hist) != 3 || hist[0]["note"] != "order #42" {
		t.Fatalf("history: %v", hist)
	}
	expect(t, "physio cannot remove", physio.call("DELETE", "/inventory/items/"+id+"/", nil, nil), 403)
	expect(t, "coordinator removes", coord.call("DELETE", "/inventory/items/"+id+"/", nil, nil), 204)
}

// TestTimeBoard: the coordinator puts a physiotherapist in a division; the
// coordinator assigns patients up to the division's capacity per hour; the
// board shows the load; a shift with patients cannot be deleted.
func TestTimeBoard(t *testing.T) {
	coord, _ := login(t, "coordinator")
	rec, _ := login(t, "receptionist")
	doctor, _ := login(t, "doctor")
	tiago, _ := login(t, "physiotherapist2")
	tiagoID := userIDOf(tiago)

	// A date of its own per run, far from the seeded weeks
	date := time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(time.Now().Unix()%5000)).Format("2006-01-02")

	var divisions []map[string]any
	expect(t, "divisions", coord.call("GET", "/schedule/divisions/", nil, &divisions), 200)
	var room string
	for _, d := range divisions {
		if d["capacity"] == 1.0 {
			room = d["id"].(string)
		}
	}
	if room == "" {
		t.Fatal("expected a capacity-1 division in the demo clinic")
	}

	shiftBody := map[string]any{"date": date, "start_hour": 9, "end_hour": 11, "therapist_id": tiagoID, "division_id": room}
	expect(t, "receptionist cannot plan shifts", rec.call("POST", "/schedule/shifts/", shiftBody, nil), 403)
	var shift map[string]any
	expect(t, "coordinator plans shift", coord.call("POST", "/schedule/shifts/", shiftBody, &shift), 201)

	var p1, p2 map[string]any
	stamp := time.Now().Format("150405.000")
	expect(t, "patient 1", rec.call("POST", "/patient/", map[string]any{"name": "Board One " + stamp}, &p1), 201)
	expect(t, "patient 2", rec.call("POST", "/patient/", map[string]any{"name": "Board Two " + stamp}, &p2), 201)
	book := func(c *client, patientID string, hour int) int {
		return c.call("POST", "/schedule/appointments/", map[string]any{
			"date": date, "hour": hour, "therapist_id": tiagoID, "patient_id": patientID,
		}, nil)
	}
	expect(t, "front desk does not assign", book(rec, p1["id"].(string), 9), 403)
	expect(t, "coordinator assigns", book(coord, p1["id"].(string), 9), 201)
	expect(t, "hour full", book(coord, p2["id"].(string), 9), 409)
	expect(t, "outside the shift", book(coord, p2["id"].(string), 15), 409)
	expect(t, "doctors do not book", book(doctor, p2["id"].(string), 10), 403)

	var board struct {
		Therapists []struct {
			ID       float64 `json:"id"`
			Patients int     `json:"patients"`
			Hours    []struct {
				Hour         int              `json:"hour"`
				Appointments []map[string]any `json:"appointments"`
			} `json:"hours"`
		} `json:"therapists"`
	}
	expect(t, "physio views board", tiago.call("GET", "/schedule/board/?date="+date, nil, &board), 200)
	var apptID string
	for _, row := range board.Therapists {
		if row.ID == tiagoID {
			if row.Patients != 1 || len(row.Hours) != 2 || len(row.Hours[0].Appointments) != 1 {
				t.Fatalf("tiago's row: %+v", row)
			}
			apptID = row.Hours[0].Appointments[0]["id"].(string)
		}
	}
	if apptID == "" {
		t.Fatal("tiago missing from the board")
	}

	expect(t, "physio marks attended", tiago.call("PATCH", "/schedule/appointments/"+apptID+"/", map[string]any{"status": "attended"}, nil), 200)
	shiftID := shift["id"].(string)
	expect(t, "busy shift cannot be deleted", coord.call("DELETE", "/schedule/shifts/"+shiftID+"/", nil, nil), 409)
	expect(t, "unbook", coord.call("DELETE", "/schedule/appointments/"+apptID+"/", nil, nil), 204)
	expect(t, "delete empty shift", coord.call("DELETE", "/schedule/shifts/"+shiftID+"/", nil, nil), 204)
}

// TestCarePlanAndAutoUse: care levels limit an hour to 8 credits (two
// level-5 patients never fit); a weekly plan books several sessions at
// once; marking a session attended uses the per-session gloves; the
// coordinator audits who used what; the year view counts the sessions.
func TestCarePlanAndAutoUse(t *testing.T) {
	coord, _ := login(t, "coordinator")
	tiago, _ := login(t, "physiotherapist2")
	tiagoID := userIDOf(tiago)
	start := time.Date(2045, 1, 2, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 7*int(time.Now().Unix()%2000)) // a Monday
	day := func(d int) string { return start.AddDate(0, 0, d).Format("2006-01-02") }

	var divisions []map[string]any
	expect(t, "divisions", coord.call("GET", "/schedule/divisions/", nil, &divisions), 200)
	var gym string
	for _, d := range divisions {
		if d["capacity"].(float64) >= 3 && d["credit_limit"] == 8.0 {
			gym = d["id"].(string)
		}
	}
	if gym == "" {
		t.Fatal("expected a multi-patient division with 8 credits")
	}
	for _, d := range []int{0, 2, 4} { // Mon, Wed, Fri
		expect(t, "shift", coord.call("POST", "/schedule/shifts/", map[string]any{
			"date": day(d), "start_hour": 9, "end_hour": 12, "therapist_id": tiagoID, "division_id": gym,
		}, nil), 201)
	}

	stamp := time.Now().Format("150405.000")
	newPatient := func(name string, level int) string {
		var p map[string]any
		expect(t, "patient "+name, coord.call("POST", "/patient/", map[string]any{"name": name + " " + stamp, "care_level": level}, &p), 201)
		if p["care_level"] != float64(level) {
			t.Fatalf("care level: %v", p["care_level"])
		}
		return p["id"].(string)
	}
	neuroA, neuroB, light := newPatient("Neuro A", 5), newPatient("Neuro B", 5), newPatient("Light", 1)
	book := func(patient string, hour int) int {
		return coord.call("POST", "/schedule/appointments/", map[string]any{"date": day(0), "hour": hour, "therapist_id": tiagoID, "patient_id": patient}, nil)
	}
	expect(t, "first level-5 patient", book(neuroA, 9), 201)
	expect(t, "second level-5 patient does not fit 8 credits", book(neuroB, 9), 409)
	expect(t, "a level-1 patient still fits", book(light, 9), 201)

	// Weekly plan: Mon/Wed/Fri at 10 for 3 sessions — preview, then book
	plan := map[string]any{"patient_id": neuroB, "therapist_id": tiagoID, "weekdays": []int{1, 3, 5}, "hour": 10,
		"start_date": day(0), "sessions": 3, "dry_run": true}
	var preview struct {
		Booked  []string `json:"booked"`
		Skipped []any    `json:"skipped"`
	}
	expect(t, "plan preview", coord.call("POST", "/schedule/series/", plan, &preview), 201)
	if len(preview.Booked) != 3 {
		t.Fatalf("preview: %+v", preview)
	}
	plan["dry_run"] = false
	var booked struct {
		SeriesID string   `json:"series_id"`
		Booked   []string `json:"booked"`
	}
	expect(t, "plan booked", coord.call("POST", "/schedule/series/", plan, &booked), 201)
	if len(booked.Booked) != 3 || booked.SeriesID == "" {
		t.Fatalf("plan: %+v", booked)
	}

	// A session-consumable item, then attending a session uses it
	var gloves map[string]any
	expect(t, "gloves", coord.call("POST", "/inventory/items/", map[string]any{
		"name": "E2E session gloves " + stamp, "unit": "pairs", "quantity": 10, "per_session": 1,
	}, &gloves), 201)
	var board struct {
		Therapists []struct {
			ID    float64 `json:"id"`
			Hours []struct {
				Hour         int              `json:"hour"`
				Appointments []map[string]any `json:"appointments"`
			} `json:"hours"`
		} `json:"therapists"`
	}
	expect(t, "board", coord.call("GET", "/schedule/board/?date="+day(2), nil, &board), 200)
	var apptID string
	for _, row := range board.Therapists {
		for _, h := range row.Hours {
			if row.ID == tiagoID && h.Hour == 10 && len(h.Appointments) == 1 {
				apptID = h.Appointments[0]["id"].(string)
			}
		}
	}
	if apptID == "" {
		t.Fatal("planned Wednesday session missing from the board")
	}
	expect(t, "tiago marks attended", tiago.call("PATCH", "/schedule/appointments/"+apptID+"/", map[string]any{"status": "attended"}, nil), 200)
	var items []map[string]any
	expect(t, "stock", coord.call("GET", "/inventory/items/?search=E2E+session+gloves+"+stamp, nil, &items), 200)
	if len(items) != 1 || items[0]["quantity"] != 9.0 {
		t.Fatalf("one pair used by the session: %v", items)
	}
	var audit []map[string]any
	expect(t, "physio cannot audit", tiago.call("GET", "/inventory/audit/", nil, nil), 403)
	expect(t, "audit", coord.call("GET", "/inventory/audit/?auto=true&item="+gloves["id"].(string), nil, &audit), 200)
	if len(audit) != 1 || audit[0]["user_id"] != tiagoID || audit[0]["appointment_id"] != apptID {
		t.Fatalf("audit of automatic use: %v", audit)
	}

	var year struct {
		Months []map[string]float64 `json:"months"`
	}
	expect(t, "year view", coord.call("GET", fmt.Sprintf("/schedule/year/?year=%d", start.Year()), nil, &year), 200)
	// The plan may cross into the next month: count the whole year
	var attended, booked2 float64
	for _, m := range year.Months {
		attended += m["attended"]
		booked2 += m["booked"]
	}
	if attended < 1 || booked2 < 4 {
		t.Fatalf("year view: %v attended, %v booked", attended, booked2)
	}
	expect(t, "cancel the rest of the plan", coord.call("DELETE", "/schedule/series/"+booked.SeriesID+"/?from="+day(0), nil, nil), 200)
}

// TestClinicPeople: the clinic's "People" page lists the users who work there.
func TestClinicPeople(t *testing.T) {
	coord, _ := login(t, "coordinator")
	var me map[string]any
	coord.call("GET", "/users/getcurrentuser/", nil, &me)
	clinic, _ := me["home_facility"].(string)
	if clinic == "" {
		if obj, ok := me["home_facility_object"].(map[string]any); ok {
			clinic, _ = obj["id"].(string)
		}
	}
	if clinic == "" {
		t.Fatalf("coordinator has no clinic: %v", me)
	}
	var page struct {
		Count   int              `json:"count"`
		Results []map[string]any `json:"results"`
	}
	expect(t, "people", coord.call("GET", "/facility/"+clinic+"/get_users/?limit=50", nil, &page), 200)
	if page.Count < 5 || len(page.Results) == 0 {
		t.Fatalf("people: %d", page.Count)
	}
	expect(t, "search", coord.call("GET", "/facility/"+clinic+"/get_users/?username=receptionist", nil, &page), 200)
	if page.Count != 1 {
		t.Errorf("search by username: %d", page.Count)
	}
}

// TestContainersAndReorder: containers are opened and finished (not used);
// single-use items are used (not opened); the reorder list says how much
// to order; per-division session rules can be set by the coordinator.
func TestContainersAndReorder(t *testing.T) {
	coord, _ := login(t, "coordinator")
	physio, _ := login(t, "physiotherapist")
	stamp := time.Now().Format("150405.000")
	var gel map[string]any
	expect(t, "container", coord.call("POST", "/inventory/items/", map[string]any{
		"name": "E2E gel " + stamp, "unit": "bottles", "tracking": "container", "quantity": 3, "min_quantity": 2, "target_quantity": 6, "open_life_days": 30,
	}, &gel), 201)
	id := gel["id"].(string)
	move := func(c *client, kind string, n int, out any) int {
		return c.call("POST", "/inventory/items/"+id+"/movements/", map[string]any{"kind": kind, "quantity": n}, out)
	}
	var after map[string]any
	expect(t, "containers are not used", move(physio, "use", 1, nil), 400)
	expect(t, "physio opens one", move(physio, "open", 1, &after), 200)
	if after["quantity"] != 2.0 || after["in_use"] != 1.0 || after["oldest_open_at"] == nil {
		t.Fatalf("after open: %v", after)
	}
	expect(t, "finish more than open", move(physio, "finish", 2, nil), 409)
	expect(t, "finish it", move(physio, "finish", 1, &after), 200)
	if after["quantity"] != 2.0 || after["in_use"] != 0.0 {
		t.Fatalf("after finish: %v", after)
	}

	var reorder []map[string]any
	expect(t, "reorder list", coord.call("GET", "/inventory/reorder/", nil, &reorder), 200)
	found := false
	for _, it := range reorder {
		if it["id"] == id {
			found = it["reorder_quantity"] == 4.0 // 6 target − 2 in stock
		}
	}
	if !found {
		t.Errorf("gel should be on the reorder list with 4 to order")
	}

	var divisions []map[string]any
	coord.call("GET", "/schedule/divisions/", nil, &divisions)
	var gloves map[string]any
	expect(t, "unit item", coord.call("POST", "/inventory/items/", map[string]any{"name": "E2E caps " + stamp, "quantity": 5}, &gloves), 201)
	rules := map[string]any{"rules": []map[string]any{{"division_id": divisions[0]["id"], "quantity": 1}}}
	expect(t, "physio cannot set rules", physio.call("PUT", "/inventory/items/"+gloves["id"].(string)+"/session-rules/", rules, nil), 403)
	var ruled map[string]any
	expect(t, "coordinator sets rules", coord.call("PUT", "/inventory/items/"+gloves["id"].(string)+"/session-rules/", rules, &ruled), 200)
	if r, _ := ruled["session_rules"].([]any); len(r) != 1 {
		t.Fatalf("rules: %v", ruled["session_rules"])
	}
	expect(t, "rules only for single-use", coord.call("PUT", "/inventory/items/"+id+"/session-rules/", rules, nil), 400)
}

// TestClimutua: the PT-PT demo clinic is there with its own staff, patients
// and Portuguese data, isolated from the other clinic.
func TestClimutua(t *testing.T) {
	coord, _ := login(t, "cm.coordenacao")
	physio, _ := login(t, "cm.fisio3")
	other, _ := login(t, "coordinator")

	var patients struct {
		Count   int              `json:"count"`
		Results []map[string]any `json:"results"`
	}
	expect(t, "patients", coord.call("GET", "/patient/?limit=100", nil, &patients), 200)
	if patients.Count < 120 {
		t.Fatalf("Climutua patients: %d", patients.Count)
	}
	var refs struct {
		Results []map[string]any `json:"results"`
	}
	expect(t, "neuro physio inbox", physio.call("GET", "/referrals/?limit=50", nil, &refs), 200)
	pt := false
	for _, r := range refs.Results {
		pt = pt || r["diagnosis"] == "Sequelas de AVC isquémico" || r["diagnosis"] == "Doença de Parkinson" || r["diagnosis"] == "Esclerose múltipla (surto recente)"
	}
	if !pt {
		t.Error("the neuro physio should see Portuguese neuro referrals")
	}
	var divisions []map[string]any
	expect(t, "divisions", coord.call("GET", "/schedule/divisions/", nil, &divisions), 200)
	if len(divisions) != 6 || divisions[0]["name"] != "Ginásio" {
		t.Errorf("divisions: %v", divisions)
	}
	var items []map[string]any
	expect(t, "stock", coord.call("GET", "/inventory/items/?search=Luvas", nil, &items), 200)
	if len(items) == 0 {
		t.Error("Climutua stock in Portuguese")
	}
	// The other clinic's coordinator sees none of it
	id := patients.Results[0]["id"].(string)
	expect(t, "other clinic cannot open the patient", other.call("GET", "/patient/"+id+"/", nil, nil), 404)
}
