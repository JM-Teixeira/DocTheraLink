package schedule

import (
	"context"
	"testing"
	"time"
)

// The example of the clinic: Sarah is away with 4 patients at 9:00; Tiago
// (a room for 1) is asked for two and takes one (the other is reported as
// full, not sent elsewhere behind the coordinator back), Ana (the cover)
// takes the unassigned one, one stays. The preview lists who works at that
// hour and how much room they have left.
func TestAbsenceSplitPerPatient(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.svc.now = func() time.Time { return time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC) }
	ana := UserSummary{ID: 12, FirstName: "Ana", LastName: "Costa"}
	f.repo.therapists = append(f.repo.therapists, ana)
	f.shift(t, monday, sarah.ID, f.gym, 9, 10)
	f.shift(t, monday, tiago.ID, f.room, 9, 10)
	f.shift(t, monday, ana.ID, f.gym, 9, 10)
	ids := map[string]string{} // appointment -> patient
	var appts []string
	for i := 0; i < 4; i++ {
		p := f.patient(2)
		ap, err := f.book(coordinator, sarah.ID, 9, p)
		if err != nil {
			t.Fatal(err)
		}
		ids[ap.ID] = p
		appts = append(appts, ap.ID)
	}
	cover := ana.ID
	req := AbsenceRequest{TherapistID: sarah.ID, CoverID: &cover, From: monday, To: monday, Reason: "vacation",
		MovePatients: true, DryRun: true, Assignments: []Assignment{
			{AppointmentID: appts[0], TherapistID: tiago.ID},
			{AppointmentID: appts[1], TherapistID: tiago.ID}, // Tiago has room for one only
			{AppointmentID: appts[3], TherapistID: 0},        // stays with Sarah
		}}
	res, err := f.svc.CreateAbsence(ctx, coordinator, req)
	if err != nil {
		t.Fatal(err)
	}
	to := map[uint]int{}
	byAppt := map[string]MovedSession{}
	for _, s := range res.Sessions {
		byAppt[s.AppointmentID] = s
		if s.Moved {
			to[s.To]++
		}
		if len(s.Candidates) != 2 {
			t.Errorf("candidates at 9:00: %+v", s.Candidates)
		}
	}
	if to[tiago.ID] != 1 || to[ana.ID] != 1 {
		t.Fatalf("Tiago takes 1, Ana 1: %v", to)
	}
	if a, b := byAppt[appts[0]], byAppt[appts[1]]; (a.Moved == b.Moved) || (!a.Moved && a.Reason != SkipFull) || (!b.Moved && b.Reason != SkipFull) {
		t.Errorf("one of the two sent to Tiago does not fit: %+v / %+v", a, b)
	}
	if s := byAppt[appts[3]]; s.Moved || s.Reason != "kept" {
		t.Errorf("kept with Sarah: %+v", s)
	}
	if s := byAppt[appts[2]]; !s.Moved || s.To != ana.ID {
		t.Errorf("unassigned goes to the cover: %+v", s)
	}

	// Saving moves them for real
	req.DryRun = false
	if _, err := f.svc.CreateAbsence(ctx, coordinator, req); err != nil {
		t.Fatal(err)
	}
	owners := map[uint]int{}
	for _, id := range appts {
		owners[f.repo.appts[id].TherapistID]++
	}
	if owners[tiago.ID] != 1 || owners[ana.ID] != 1 || owners[sarah.ID] != 2 {
		t.Errorf("after saving: %v", owners)
	}
	// Sessions never go to the absent physio, nor to someone outside the clinic
	bad := req
	bad.Assignments = []Assignment{{AppointmentID: appts[0], TherapistID: 999}}
	bad.From, bad.To = "2026-11-02", "2026-11-02"
	if _, err := f.svc.CreateAbsence(ctx, coordinator, bad); err == nil {
		t.Error("unknown physio must be refused")
	}
}
