//go:build e2e

package e2e

import (
	"testing"
	"time"
)

// TestMoveAndReschedule: a session moves to another hour only where it fits;
// a plan's remaining sessions move to a new weekly pattern.
func TestMoveAndReschedule(t *testing.T) {
	coord, _ := login(t, "de.marta.ribeiro")
	desk, _ := login(t, "de.rita.lopes")
	tiago, _ := login(t, "de.tiago.martins")
	tiagoID := userIDOf(tiago)

	var divisions []map[string]any
	expect(t, "divisions", coord.call("GET", "/schedule/divisions/", nil, &divisions), 200)
	var room string
	for _, d := range divisions {
		if d["capacity"] == 1.0 {
			room = d["id"].(string)
		}
	}
	// A free Monday far ahead, and the two weeks after it
	m := time.Date(2070, 1, 6, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 7*int(time.Now().Unix()%300)) // a Monday
	for freeDay(t, coord, tiagoID, m) != m.Format("2006-01-02") {
		m = m.AddDate(0, 0, 7)
	}
	monday := m.Format("2006-01-02")
	var shifts []string
	var appts []string
	defer func() {
		for _, id := range appts {
			coord.call("DELETE", "/schedule/appointments/"+id+"/", nil, nil)
		}
		for _, id := range shifts {
			coord.call("DELETE", "/schedule/shifts/"+id+"/", nil, nil)
		}
	}()
	for i := 0; i < 14; i++ {
		var sh map[string]any
		date := m.AddDate(0, 0, i).Format("2006-01-02")
		if coord.call("POST", "/schedule/shifts/", map[string]any{"date": date, "start_hour": 9, "end_hour": 12,
			"therapist_id": tiagoID, "division_id": room}, &sh) == 201 {
			shifts = append(shifts, sh["id"].(string))
		}
	}
	var a, b map[string]any
	expect(t, "patient A", desk.call("POST", "/patient/", map[string]any{"name": "Mover A " + time.Now().Format("150405.000")}, &a), 201)
	expect(t, "patient B", desk.call("POST", "/patient/", map[string]any{"name": "Mover B " + time.Now().Format("150405.000")}, &b), 201)
	var apA, apB map[string]any
	expect(t, "book A at 9", coord.call("POST", "/schedule/appointments/", map[string]any{"date": monday, "hour": 9, "therapist_id": tiagoID, "patient_id": a["id"]}, &apA), 201)
	expect(t, "book B at 10", coord.call("POST", "/schedule/appointments/", map[string]any{"date": monday, "hour": 10, "therapist_id": tiagoID, "patient_id": b["id"]}, &apB), 201)
	appts = append(appts, apA["id"].(string), apB["id"].(string))

	move := func(c *client, id string, hour int) int {
		return c.call("POST", "/schedule/appointments/"+id+"/move/", map[string]any{"date": monday, "hour": hour, "therapist_id": tiagoID}, nil)
	}
	expect(t, "room for one: 10 is taken", move(coord, apA["id"].(string), 10), 409)
	expect(t, "not on shift at 15", move(coord, apA["id"].(string), 15), 409)
	expect(t, "physios do not move", move(tiago, apA["id"].(string), 11), 403)
	expect(t, "to 11", move(coord, apA["id"].(string), 11), 200)

	// A plan Mon/Wed/Fri at 9, then from Wednesday on Tue/Thu at 11
	var plan map[string]any
	expect(t, "plan", coord.call("POST", "/schedule/series/", map[string]any{"patient_id": b["id"], "therapist_id": tiagoID,
		"weekdays": []int{1, 3, 5}, "hour": 9, "start_date": monday, "sessions": 6}, &plan), 201)
	sid := plan["series_id"].(string)
	defer coord.call("DELETE", "/schedule/series/"+sid+"/?from="+monday, nil, nil)
	wed := m.AddDate(0, 0, 2).Format("2006-01-02")
	req := map[string]any{"from": wed, "weekdays": []int{2, 4}, "hour": 11, "therapist_id": tiagoID, "dry_run": true}
	var pre map[string]any
	expect(t, "preview", coord.call("POST", "/schedule/series/"+sid+"/reschedule/", req, &pre), 200)
	if pre["replaced"] != 5.0 {
		t.Fatalf("5 sessions from Wednesday: %v", pre["replaced"])
	}
	req["dry_run"] = false
	var done map[string]any
	expect(t, "reschedule", coord.call("POST", "/schedule/series/"+sid+"/reschedule/", req, &done), 200)
	if done["series_id"] != sid {
		t.Errorf("same plan: %v", done["series_id"])
	}
	var rest []map[string]any
	expect(t, "plan sessions", coord.call("GET", "/schedule/series/"+sid+"/", nil, &rest), 200)
	at11 := 0
	for _, s := range rest {
		if s["hour"] == 11.0 {
			at11++
		}
	}
	if at11 == 0 {
		t.Errorf("rescheduled sessions at 11: %v", rest)
	}
}

// TestWaitingList: the front desk queues a patient once; urgent ones come
// first; booking marks the entry scheduled and it leaves the list.
func TestWaitingList(t *testing.T) {
	coord, _ := login(t, "de.marta.ribeiro")
	desk, _ := login(t, "de.rita.lopes")
	physio, _ := login(t, "de.sarah.johnson")
	stamp := time.Now().Format("150405.000")
	var p1, p2 map[string]any
	expect(t, "patient 1", desk.call("POST", "/patient/", map[string]any{"name": "Espera " + stamp}, &p1), 201)
	expect(t, "patient 2", desk.call("POST", "/patient/", map[string]any{"name": "Espera urgente " + stamp}, &p2), 201)

	var e1, e2 map[string]any
	expect(t, "front desk queues", desk.call("POST", "/schedule/waiting/", map[string]any{"patient_id": p1["id"], "weekdays": []int{1, 3}, "period": "morning"}, &e1), 201)
	expect(t, "twice", desk.call("POST", "/schedule/waiting/", map[string]any{"patient_id": p1["id"]}, nil), 409)
	expect(t, "physios do not queue", physio.call("POST", "/schedule/waiting/", map[string]any{"patient_id": p2["id"]}, nil), 403)
	expect(t, "urgent", coord.call("POST", "/schedule/waiting/", map[string]any{"patient_id": p2["id"], "urgent": true}, &e2), 201)
	defer coord.call("PATCH", "/schedule/waiting/"+e2["id"].(string)+"/", map[string]any{"status": "removed"}, nil)

	var list []map[string]any
	expect(t, "list", physio.call("GET", "/schedule/waiting/", nil, &list), 200)
	pos := map[string]int{}
	for i, e := range list {
		pos[e["id"].(string)] = i
	}
	if pos[e2["id"].(string)] > pos[e1["id"].(string)] {
		t.Errorf("urgent comes first: %v", pos)
	}
	expect(t, "scheduled", coord.call("PATCH", "/schedule/waiting/"+e1["id"].(string)+"/", map[string]any{"status": "scheduled"}, nil), 200)
	expect(t, "list again", coord.call("GET", "/schedule/waiting/", nil, &list), 200)
	for _, e := range list {
		if e["id"] == e1["id"] {
			t.Error("scheduled entry still waiting")
		}
	}
}
