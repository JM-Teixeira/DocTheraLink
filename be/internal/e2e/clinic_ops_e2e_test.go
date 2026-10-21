//go:build e2e

package e2e

import (
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
// front desk assigns patients up to the division's capacity per hour; the
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
	expect(t, "front desk books", book(rec, p1["id"].(string), 9), 201)
	expect(t, "hour full", book(rec, p2["id"].(string), 9), 409)
	expect(t, "outside the shift", book(rec, p2["id"].(string), 15), 409)
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
	expect(t, "unbook", rec.call("DELETE", "/schedule/appointments/"+apptID+"/", nil, nil), 204)
	expect(t, "delete empty shift", coord.call("DELETE", "/schedule/shifts/"+shiftID+"/", nil, nil), 204)
}
