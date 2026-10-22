package schedule

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Moving a session obeys the booking rules at the new slot.
func TestMoveSession(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.shift(t, monday, sarah.ID, f.gym, 9, 12)
	f.shift(t, monday, tiago.ID, f.room, 9, 12) // room: 1 patient
	p1, p2, heavy := f.patient(2), f.patient(2), f.patient(5)
	ap1, _ := f.book(coordinator, sarah.ID, 9, p1)
	f.book(coordinator, tiago.ID, 10, p2)
	ap3, _ := f.book(coordinator, sarah.ID, 11, heavy)

	move := func(a Actor, id string, hour int, who uint) error {
		_, err := f.svc.Move(ctx, a, id, MoveRequest{Date: monday, Hour: hour, TherapistID: who})
		return err
	}
	if err := move(coordinator, ap1.ID, 10, sarah.ID); err != nil {
		t.Fatalf("to another hour: %v", err)
	}
	if got := f.repo.appts[ap1.ID]; got.Hour != 10 || got.TherapistID != sarah.ID {
		t.Fatalf("moved: %+v", got)
	}
	if err := move(coordinator, ap1.ID, 10, tiago.ID); !errors.Is(err, ErrFull) && !errors.Is(err, ErrPatientBusy) {
		t.Errorf("Tiago's room is full at 10: %v", err)
	}
	if err := move(coordinator, ap1.ID, 11, tiago.ID); err != nil {
		t.Errorf("to Tiago at 11: %v", err)
	}
	var cerr *ConflictError
	if err := move(coordinator, ap1.ID, 15, sarah.ID); !errors.As(err, &cerr) {
		t.Errorf("Sarah does not work at 15: %v", err)
	}
	if err := move(sarahActor, ap3.ID, 9, sarah.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("physios do not move sessions: %v", err)
	}
	// Moving onto its own hour is fine (the session does not block itself)
	if err := move(coordinator, ap3.ID, 11, sarah.ID); err != nil {
		t.Errorf("same slot: %v", err)
	}
}

// A plan's remaining sessions move to a new weekly pattern; the preview is
// not blocked by the plan's own sessions and the count is kept.
func TestReschedulePlan(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.svc.now = func() time.Time { return time.Date(2026, 10, 25, 8, 0, 0, 0, time.UTC) }
	for i := 0; i < 21; i++ { // three weeks of shifts
		d := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i).Format(dateLayout)
		f.shift(t, d, sarah.ID, f.gym, 9, 18)
		f.shift(t, d, tiago.ID, f.gym, 9, 18)
	}
	p := f.patient(2)
	res, err := f.svc.Series(ctx, coordinator, SeriesRequest{PatientID: p, TherapistID: sarah.ID,
		Weekdays: []int{1, 3, 5}, Hour: 9, StartDate: monday, Sessions: 6})
	if err != nil || len(res.Booked) != 6 {
		t.Fatalf("plan: %+v %v", res, err)
	}
	// From Wednesday: Tue/Thu at 17:00 with Tiago
	req := PlanRescheduleRequest{From: "2026-10-28", Weekdays: []int{2, 4}, Hour: 17, TherapistID: tiago.ID, DryRun: true}
	pre, err := f.svc.ReschedulePlan(ctx, coordinator, res.SeriesID, req)
	if err != nil || pre.Replaced != 5 || len(pre.Booked) != 5 {
		t.Fatalf("preview: %+v %v", pre, err)
	}
	req.DryRun = false
	done, err := f.svc.ReschedulePlan(ctx, coordinator, res.SeriesID, req)
	if err != nil || done.SeriesID != res.SeriesID {
		t.Fatalf("reschedule keeps the plan: %+v %v", done, err)
	}
	monday9, tiago17 := 0, 0
	for _, ap := range f.repo.appts {
		switch {
		case ap.TherapistID == sarah.ID && ap.Hour == 9:
			monday9++
		case ap.TherapistID == tiago.ID && ap.Hour == 17 && (ap.Date.Weekday() == time.Tuesday || ap.Date.Weekday() == time.Thursday):
			tiago17++
		}
	}
	if monday9 != 1 || tiago17 != 5 {
		t.Errorf("Monday's session stays, 5 move: %d / %d", monday9, tiago17)
	}
	if _, err := f.svc.ReschedulePlan(ctx, sarahActor, res.SeriesID, req); !errors.Is(err, ErrForbidden) {
		t.Errorf("physios do not reschedule plans: %v", err)
	}
	if _, err := f.svc.ReschedulePlan(ctx, coordinator, "00000000-0000-4000-8000-000000000000", req); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown plan: %v", err)
	}
}

// The waiting list: urgent first, then oldest; once per patient while
// waiting; the front desk and coordinators manage it.
func TestWaitingList(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	a, b, c := f.patient(1), f.patient(3), f.patient(2)
	add := func(who Actor, p string, urgent bool) (*WaitingEntry, error) {
		return f.svc.AddWaiting(ctx, who, WaitingRequest{PatientID: p, Urgent: urgent, Weekdays: []int{1, 3}, Period: "morning"})
	}
	first, err := add(reception, a, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := add(coordinator, b, false); err != nil {
		t.Fatal(err)
	}
	if _, err := add(coordinator, c, true); err != nil {
		t.Fatal(err)
	}
	var cerr *ConflictError
	if _, err := add(reception, a, false); !errors.As(err, &cerr) {
		t.Errorf("twice: %v", err)
	}
	if _, err := add(sarahActor, f.patient(1), false); !errors.Is(err, ErrForbidden) {
		t.Errorf("physios do not queue: %v", err)
	}
	list, _ := f.svc.WaitingList(ctx, sarahActor, "")
	if len(list) != 3 || list[0].PatientID != c || list[1].PatientID != a {
		t.Fatalf("order urgent then oldest: %+v", list)
	}
	sid := "11111111-1111-4111-8111-111111111111"
	e, err := f.svc.UpdateWaiting(ctx, coordinator, first.ID, WaitingUpdate{Status: WaitingScheduled, SeriesID: &sid})
	if err != nil || e.Status != WaitingScheduled || e.SeriesID == nil {
		t.Fatalf("scheduled: %+v %v", e, err)
	}
	if list, _ = f.svc.WaitingList(ctx, coordinator, ""); len(list) != 2 {
		t.Errorf("scheduled leaves the list: %d", len(list))
	}
	if done, _ := f.svc.WaitingList(ctx, coordinator, WaitingScheduled); len(done) != 1 {
		t.Errorf("kept as history: %d", len(done))
	}
	// Once scheduled the patient can wait again later
	if _, err := add(reception, a, false); err != nil {
		t.Errorf("back on the list: %v", err)
	}
}
