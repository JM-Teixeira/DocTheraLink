package schedule

import (
	"context"
	"errors"
	"testing"
	"time"
)

// A miss is justified or not; three unjustified misses in 90 days flag the
// patient to the coordinator (justified ones never do).
func TestMissedSessions(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.svc.now = func() time.Time { return time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC) }
	f.shift(t, monday, sarah.ID, f.gym, 8, 14)
	often, ill := f.patient(1), f.patient(1)

	yes, no := true, false
	miss := func(patient string, hour int, justified *bool, reason string) *StatusResult {
		t.Helper()
		ap, err := f.book(coordinator, sarah.ID, hour, patient)
		if err != nil {
			t.Fatal(err)
		}
		res, err := f.svc.SetStatus(ctx, sarahActor, ap.ID, StatusRequest{Status: StatusMissed, Justified: justified, Reason: reason})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	r := miss(often, 8, nil, "")
	if r.MissJustified == nil || *r.MissJustified {
		t.Fatalf("a miss is unjustified unless said otherwise: %+v", r.MissJustified)
	}
	miss(often, 9, &no, "")
	// Ill patient: justified misses only
	for h := 10; h <= 13; h++ {
		if r := miss(ill, h, &yes, "Doente"); r.MissReason != "Doente" {
			t.Fatalf("reason: %q", r.MissReason)
		}
	}

	rep, err := f.svc.MissedFlags(ctx, coordinator)
	if err != nil || len(rep.Patients) != 0 || rep.Threshold != MissFlagThreshold {
		t.Fatalf("two unjustified misses are not flagged yet: %+v %v", rep, err)
	}
	third := miss(often, 11, nil, "")
	if _, err := f.svc.MissedFlags(ctx, sarahActor); !errors.Is(err, ErrForbidden) {
		t.Errorf("physios do not get the report: %v", err)
	}
	rep, _ = f.svc.MissedFlags(ctx, coordinator)
	if len(rep.Patients) != 1 || rep.Patients[0].PatientID != often || rep.Patients[0].Unjustified != 3 {
		t.Fatalf("flagged after three: %+v", rep.Patients)
	}

	// The third turns out justified: no longer flagged
	if _, err := f.svc.SetStatus(ctx, coordinator, third.ID, StatusRequest{Status: StatusMissed, Justified: &yes, Reason: "Avisou"}); err != nil {
		t.Fatal(err)
	}
	if rep, _ = f.svc.MissedFlags(ctx, coordinator); len(rep.Patients) != 0 {
		t.Errorf("justified later: %+v", rep.Patients)
	}
	counts, err := f.svc.PatientMisses(ctx, sarahActor, ill)
	if err != nil || counts.Justified != 4 || counts.Unjustified != 0 {
		t.Errorf("patient counts: %+v %v", counts, err)
	}
	// Marking it attended clears the justification
	res, _ := f.svc.SetStatus(ctx, coordinator, third.ID, StatusRequest{Status: StatusAttended})
	if res.MissJustified != nil || f.repo.appts[third.ID].MissJustified != nil {
		t.Errorf("attended keeps no miss justification")
	}
	if _, err := f.svc.PatientMisses(ctx, coordinator, "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Errorf("bad id: %v", err)
	}
}
