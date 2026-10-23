package timeline

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"klyniva/internal/models"
	"klyniva/internal/modules/attachment"
	"klyniva/internal/modules/patient"
	"klyniva/internal/modules/referral"
	"klyniva/internal/modules/schedule"
)

// fakeSources returns fixed data; referrals is nil for roles without access.
type fakeSources struct {
	patient   *models.Patient
	referrals []referral.Referral
	notes     map[string][]referral.Note
	files     map[string][]attachment.Attachment
	sessions  []schedule.Appointment
}

func (f *fakeSources) Patient(_ context.Context, _ Actor, id string) (*models.Patient, error) {
	if f.patient == nil || f.patient.ID != id {
		return nil, patient.ErrNotFound
	}
	return f.patient, nil
}
func (f *fakeSources) Referrals(context.Context, Actor, string) ([]referral.Referral, error) {
	return f.referrals, nil
}
func (f *fakeSources) Notes(_ context.Context, _ Actor, id string) ([]referral.Note, error) {
	return f.notes[id], nil
}
func (f *fakeSources) Files(_ context.Context, _ Actor, id string) ([]attachment.Attachment, error) {
	return f.files[id], nil
}
func (f *fakeSources) Sessions(context.Context, Actor, string) ([]schedule.Appointment, error) {
	return f.sessions, nil
}

var (
	nurse = Actor{ID: 5, Role: models.RoleNurse}
	day   = func(d int) time.Time { return time.Date(2026, 10, d, 9, 0, 0, 0, time.UTC) }
	now   = day(20)
)

func TestTimelineMergesAndOrders(t *testing.T) {
	pain := 6
	accepted := day(3)
	src := &fakeSources{
		patient: &models.Patient{ID: "p1", CreatedAt: day(1)},
		referrals: []referral.Referral{{
			ID: "r1", Diagnosis: "Lombalgia", Status: referral.StatusInProgress, CreatedAt: day(2), AcceptedAt: &accepted,
			Doctor: &referral.UserSummary{FirstName: "Rita", LastName: "Lopes"},
		}},
		notes: map[string][]referral.Note{"r1": {
			{ID: "n1", Kind: referral.NoteProgress, Body: strings.Repeat("melhorou ", 60), Pain: &pain, CreatedAt: day(10)},
			{ID: "n2", Kind: referral.NoteSystem, Body: "accepted", CreatedAt: day(3)},
		}},
		files: map[string][]attachment.Attachment{"r1": {{ID: "f1", FileName: "rx.pdf", CreatedAt: day(4)}}},
		sessions: []schedule.Appointment{
			{ID: "s3", Date: day(27), Hour: 10, Status: schedule.StatusBooked, Therapist: &schedule.UserSummary{FirstName: "Rui"}},
			{ID: "s2", Date: day(22), Hour: 9, Status: schedule.StatusBooked},
			{ID: "s1", Date: day(12), Hour: 9, Status: schedule.StatusMissed, MissJustified: new(bool), MissReason: "doente"},
		},
	}
	tl, err := NewService(src, func() time.Time { return now }).For(context.Background(), nurse, "p1")
	if err != nil {
		t.Fatal(err)
	}

	// Upcoming: future booked sessions, soonest first
	if len(tl.Upcoming) != 2 || tl.Upcoming[0].ItemID != "s2" || tl.Upcoming[1].Who != "Rui" {
		t.Fatalf("upcoming: %+v", tl.Upcoming)
	}
	// History: newest first; the system note is left out (shown as a step)
	var got []string
	for _, e := range tl.History {
		got = append(got, string(e.Kind)+":"+e.Status)
	}
	want := "session:missed note:progress file: referral:accepted referral:created registered:"
	if strings.Join(got, " ") != want {
		t.Errorf("history order:\n got %s\nwant %s", strings.Join(got, " "), want)
	}
	note := tl.History[1]
	if *note.Pain != 6 || len([]rune(note.Text)) > excerptLen+1 || !strings.HasSuffix(note.Text, "…") {
		t.Errorf("note excerpt/pain: %q %v", note.Text, note.Pain)
	}
	if tl.History[0].Text != "doente" || *tl.History[0].Hour != 9 || tl.History[0].Date != "2026-10-12" {
		t.Errorf("missed session entry: %+v", tl.History[0])
	}
	if tl.History[4].Who != "Rita Lopes" || tl.History[4].Text != "Lombalgia" {
		t.Errorf("created step: %+v", tl.History[4])
	}
}

func TestTimelineReferralEndings(t *testing.T) {
	done, reason := day(9), "não precisa"
	for _, tc := range []struct {
		r    referral.Referral
		last string
		text string
	}{
		{referral.Referral{Status: referral.StatusCompleted, CompletedAt: &done, OutcomeSummary: &reason}, "completed", reason},
		{referral.Referral{Status: referral.StatusDeclined, DeclineReason: &reason, UpdatedAt: day(5)}, "declined", reason},
		{referral.Referral{Status: referral.StatusCancelled, CancelReason: &reason, UpdatedAt: day(5)}, "cancelled", reason},
		{referral.Referral{Status: referral.StatusPending}, "created", ""},
	} {
		steps := referralSteps(&tc.r)
		last := steps[len(steps)-1]
		if last.Status != tc.last || last.Text != tc.text {
			t.Errorf("%s: last step %+v", tc.r.Status, last)
		}
	}
}

func TestTimelineHiddenSourcesAndErrors(t *testing.T) {
	ctx := context.Background()
	// A receptionist: no referrals, only sessions; still a valid page
	src := &fakeSources{patient: &models.Patient{ID: "p1", CreatedAt: day(1)},
		sessions: []schedule.Appointment{{ID: "s1", Date: day(5), Hour: 8, Status: schedule.StatusAttended}}}
	tl, err := NewService(src, func() time.Time { return now }).For(ctx, nurse, "p1")
	if err != nil || len(tl.History) != 2 || len(tl.Upcoming) != 0 {
		t.Fatalf("err=%v %+v", err, tl)
	}
	// A patient the actor cannot see is not found
	if _, err := NewService(src, nil).For(ctx, nurse, "other"); !errors.Is(err, patient.ErrNotFound) {
		t.Errorf("want not found, got %v", err)
	}
	// Forbidden sources are hidden by the adapter, not failed
	if !hidden(referral.ErrForbidden) || !hidden(schedule.ErrNoClinic) || hidden(errors.New("db down")) {
		t.Error("hidden() classification")
	}
	items, err := visible([]int{1}, schedule.ErrNoClinic)
	if err != nil || items != nil {
		t.Errorf("visible: %v %v", items, err)
	}
}
