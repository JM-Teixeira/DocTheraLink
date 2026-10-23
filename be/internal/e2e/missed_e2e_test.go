//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

// TestMissedSessionsFlag: three unjustified misses in the last 90 days flag
// the patient to the coordinator; justifying one removes the flag.
func TestMissedSessionsFlag(t *testing.T) {
	coord, _ := login(t, "de.marta.ribeiro")
	desk, _ := login(t, "de.rita.lopes")
	tiago, _ := login(t, "de.tiago.martins")
	tiagoID := userIDOf(tiago)

	var divisions []map[string]any
	expect(t, "divisions", coord.call("GET", "/schedule/divisions/", nil, &divisions), 200)
	div := divisions[0]["id"].(string)
	var p map[string]any
	expect(t, "patient", desk.call("POST", "/patient/", map[string]any{"name": "Faltas " + time.Now().Format("150405.000")}, &p), 201)
	pid := p["id"].(string)

	// Three past days of its own (within the 90-day window), late hour
	base := 20 + int(time.Now().Unix()%30)
	var shifts, appts []string
	defer func() {
		for _, id := range appts {
			coord.call("DELETE", "/schedule/appointments/"+id+"/", nil, nil)
		}
		for _, id := range shifts {
			coord.call("DELETE", "/schedule/shifts/"+id+"/", nil, nil)
		}
	}()
	for i := 0; i < 3; i++ {
		date := time.Now().AddDate(0, 0, -(base + i)).Format("2006-01-02")
		var sh, ap map[string]any
		if code := coord.call("POST", "/schedule/shifts/", map[string]any{
			"date": date, "start_hour": 19, "end_hour": 20, "therapist_id": tiagoID, "division_id": div,
		}, &sh); code != 201 {
			t.Fatalf("shift %s: %d", date, code)
		}
		shifts = append(shifts, sh["id"].(string))
		expect(t, "book", coord.call("POST", "/schedule/appointments/", map[string]any{
			"date": date, "hour": 19, "therapist_id": tiagoID, "patient_id": pid,
		}, &ap), 201)
		appts = append(appts, ap["id"].(string))
	}

	flagged := func() map[string]any {
		var rep struct {
			Patients []map[string]any `json:"patients"`
		}
		expect(t, "report", coord.call("GET", "/schedule/missed/", nil, &rep), 200)
		for _, m := range rep.Patients {
			if m["patient_id"] == pid {
				return m
			}
		}
		return nil
	}
	var res map[string]any
	expect(t, "physio marks missed", tiago.call("PATCH", "/schedule/appointments/"+appts[0]+"/", map[string]any{"status": "missed"}, &res), 200)
	if res["miss_justified"] != false {
		t.Fatalf("a miss is unjustified by default: %v", res["miss_justified"])
	}
	expect(t, "missed 2", tiago.call("PATCH", "/schedule/appointments/"+appts[1]+"/", map[string]any{"status": "missed", "justified": false}, nil), 200)
	if flagged() != nil {
		t.Fatal("two misses are not flagged yet")
	}
	expect(t, "missed 3", tiago.call("PATCH", "/schedule/appointments/"+appts[2]+"/", map[string]any{"status": "missed"}, nil), 200)
	if m := flagged(); m == nil || m["unjustified"] != 3.0 {
		t.Fatalf("flagged after three: %v", m)
	}
	expect(t, "physios do not see the report", tiago.call("GET", "/schedule/missed/", nil, nil), 403)

	expect(t, "justify one", coord.call("PATCH", "/schedule/appointments/"+appts[2]+"/", map[string]any{
		"status": "missed", "justified": true, "reason": "Doente, avisou",
	}, &res), 200)
	if res["miss_justified"] != true || res["miss_reason"] != "Doente, avisou" {
		t.Fatalf("justified: %v", res)
	}
	if flagged() != nil {
		t.Error("a justified miss must not count")
	}
	var counts map[string]any
	expect(t, "patient counts", tiago.call("GET", "/schedule/missed/"+pid+"/", nil, &counts), 200)
	if counts["unjustified"] != 2.0 || counts["justified"] != 1.0 {
		t.Errorf("counts: %v", counts)
	}
	other, _ := login(t, "cm.ana.ferreira")
	expect(t, "other clinic", other.call("GET", fmt.Sprintf("/schedule/missed/%s/", pid), nil, nil), 404)
}

// TestVersion: the release number is public and looks like 0.minor.patch.
func TestVersion(t *testing.T) {
	anon := &client{t: t, base: baseURL(t)}
	var v map[string]string
	expect(t, "version", anon.call("GET", "/version/", nil, &v), 200)
	var major, minor, patch int
	if n, _ := fmt.Sscanf(v["version"], "%d.%d.%d", &major, &minor, &patch); n != 3 || major != 0 {
		t.Errorf("version %q", v["version"])
	}
}
