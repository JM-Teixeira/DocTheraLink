package referral

import (
	"context"
	"errors"
	"testing"
	"time"

	"klyniva/internal/clinic"
	"klyniva/internal/models"
)

const patientID = "7f1c1f9e-6a3e-4b3c-9a43-0d2b5e1b7a11"

var (
	doctor      = Actor{ID: 1, Role: models.RoleDoctor}
	otherDoctor = Actor{ID: 2, Role: models.RoleDoctor}
	physio      = Actor{ID: 10, Role: models.RolePhysioTherapist}
	otherPhysio = Actor{ID: 11, Role: models.RolePhysioTherapist}
	admin       = Actor{ID: 99, Role: models.RoleStateAdmin}
	nurse       = Actor{ID: 20, Role: models.RoleNurse}
	farPhysio   = Actor{ID: 12, Role: models.RolePhysioTherapist} // works in another clinic
)

const (
	clinicA = "0b5a1c1e-0000-4000-8000-00000000000a"
	clinicB = "0b5a1c1e-0000-4000-8000-00000000000b"
)

var fixedNow = time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)

// newTestService returns a service over a seeded fake repository.
func newTestService(t *testing.T) (*Service, *fakeRepo) {
	t.Helper()
	repo := newFakeRepo()
	repo.patients[patientID] = true
	repo.addUser(physio.ID, string(models.RolePhysioTherapist), true)
	repo.addUser(otherPhysio.ID, string(models.RolePhysioTherapist), true)
	repo.addUser(nurse.ID, string(models.RoleNurse), true)
	repo.addUser(farPhysio.ID, string(models.RolePhysioTherapist), true)
	clinics := clinic.Static{
		doctor.ID: clinicA, otherDoctor.ID: clinicA, physio.ID: clinicA,
		otherPhysio.ID: clinicA, nurse.ID: clinicA, farPhysio.ID: clinicB,
	}
	return NewService(repo, clinics, func() time.Time { return fixedNow }), repo
}

func validCreate() CreateRequest {
	return CreateRequest{PatientID: patientID, Diagnosis: "ACL reconstruction", Reason: "Post-op rehab"}
}

// mustCreate opens a referral as doctor, optionally addressed to a therapist.
func mustCreate(t *testing.T, s *Service, therapist *uint) *Referral {
	t.Helper()
	req := validCreate()
	req.TherapistID = therapist
	r, err := s.Create(context.Background(), doctor, req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return r
}

func ptrUint(v uint) *uint { return &v }

func TestCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("doctor creates pending routine referral", func(t *testing.T) {
		s, _ := newTestService(t)
		r, err := s.Create(ctx, doctor, validCreate())
		if err != nil {
			t.Fatal(err)
		}
		if r.Status != StatusPending || r.Priority != PriorityRoutine || r.DoctorID != doctor.ID {
			t.Fatalf("unexpected referral: %+v", r)
		}
	})

	cases := []struct {
		name  string
		actor Actor
		edit  func(*CreateRequest)
		want  any // error sentinel or *ValidationError
	}{
		{"physio cannot refer", physio, nil, ErrForbidden},
		{"nurse cannot refer", nurse, nil, ErrForbidden},
		{"unknown patient", doctor, func(r *CreateRequest) { r.PatientID = "00000000-0000-0000-0000-000000000000" }, &ValidationError{}},
		{"blank diagnosis", doctor, func(r *CreateRequest) { r.Diagnosis = "  " }, &ValidationError{}},
		{"therapist must be a physio", doctor, func(r *CreateRequest) { r.TherapistID = ptrUint(nurse.ID) }, &ValidationError{}},
		{"therapist must exist", doctor, func(r *CreateRequest) { r.TherapistID = ptrUint(404) }, &ValidationError{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newTestService(t)
			req := validCreate()
			if tc.edit != nil {
				tc.edit(&req)
			}
			_, err := s.Create(ctx, tc.actor, req)
			assertErr(t, err, tc.want)
		})
	}

	t.Run("admin may refer", func(t *testing.T) {
		s, _ := newTestService(t)
		if _, err := s.Create(ctx, admin, validCreate()); err != nil {
			t.Fatal(err)
		}
	})
}

// TestClinicIsolation: referrals belong to the doctor's clinic and never leak
// into another clinic's therapist inbox.
func TestClinicIsolation(t *testing.T) {
	ctx := context.Background()
	s, _ := newTestService(t)
	r := mustCreate(t, s, nil)
	if r.FacilityID == nil || *r.FacilityID != clinicA {
		t.Fatalf("referral should take the doctor's clinic, got %v", r.FacilityID)
	}

	if _, err := s.Get(ctx, farPhysio, r.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("therapist of another clinic must not see the open referral, got %v", err)
	}
	if _, total, _ := s.List(ctx, farPhysio, ListFilter{}); total != 0 {
		t.Errorf("other clinic inbox should be empty, got %d", total)
	}
	_, err := s.Accept(ctx, farPhysio, r.ID)
	assertErr(t, err, ErrNotFound)

	req := validCreate()
	req.TherapistID = ptrUint(farPhysio.ID)
	_, err = s.Create(ctx, doctor, req)
	assertErr(t, err, &ValidationError{})

	other := clinicB
	req = validCreate()
	req.FacilityID = &other
	_, err = s.Create(ctx, doctor, req)
	assertErr(t, err, &ValidationError{})
}

func TestReassign(t *testing.T) {
	ctx := context.Background()
	handOver := func(to uint) ReassignRequest { return ReassignRequest{TherapistID: to, Reason: "Holiday cover"} }

	t.Run("doctor addresses an open referral to a therapist", func(t *testing.T) {
		s, _ := newTestService(t)
		r := mustCreate(t, s, nil)
		r, err := s.Reassign(ctx, doctor, r.ID, handOver(physio.ID))
		if err != nil || r.TherapistID == nil || *r.TherapistID != physio.ID || r.Status != StatusPending {
			t.Fatalf("err=%v r=%+v", err, r)
		}
	})

	t.Run("assigned therapist hands over mid-treatment; status is kept", func(t *testing.T) {
		s, repo := newTestService(t)
		r := mustCreate(t, s, ptrUint(physio.ID))
		_, _ = s.Accept(ctx, physio, r.ID)
		_, _ = s.Start(ctx, physio, r.ID)
		r, err := s.Reassign(ctx, physio, r.ID, handOver(otherPhysio.ID))
		if err != nil || *r.TherapistID != otherPhysio.ID || r.Status != StatusInProgress {
			t.Fatalf("err=%v r=%+v", err, r)
		}
		// The new therapist can now act; the old one no longer sees it.
		if _, err := s.Complete(ctx, otherPhysio, r.ID, "done"); err != nil {
			t.Fatalf("new therapist complete: %v", err)
		}
		notes, _ := repo.ListNotes(ctx, r.ID)
		if notes[len(notes)-2].Body != "Referral handed over: Holiday cover" {
			t.Errorf("hand-over not recorded on thread: %+v", notes)
		}
	})

	cases := []struct {
		name  string
		actor Actor
		to    uint
		setup func(*Service, string)
		want  any
	}{
		{"nurse cannot hand over", nurse, otherPhysio.ID, nil, ErrNotFound},
		{"unrelated therapist cannot hand over", otherPhysio, physio.ID, nil, ErrNotFound},
		{"must be a physiotherapist", doctor, nurse.ID, nil, &ValidationError{}},
		{"must be in the same clinic", doctor, farPhysio.ID, nil, &ValidationError{}},
		{"not to the same therapist", doctor, physio.ID, nil, &ValidationError{}},
		{"not after completion", doctor, otherPhysio.ID, func(s *Service, id string) {
			_, _ = s.Accept(ctx, physio, id)
			_, _ = s.Start(ctx, physio, id)
			_, _ = s.Complete(ctx, physio, id, "ok")
		}, ErrInvalidTransition},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newTestService(t)
			r := mustCreate(t, s, ptrUint(physio.ID))
			if tc.setup != nil {
				tc.setup(s, r.ID)
			}
			_, err := s.Reassign(ctx, tc.actor, r.ID, handOver(tc.to))
			assertErr(t, err, tc.want)
		})
	}
}

func TestReopen(t *testing.T) {
	ctx := context.Background()
	declined := func(t *testing.T, s *Service) *Referral {
		r := mustCreate(t, s, ptrUint(physio.ID))
		r, err := s.Decline(ctx, physio, r.ID, "No capacity")
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	t.Run("doctor sends a declined referral back to the clinic inbox", func(t *testing.T) {
		s, _ := newTestService(t)
		r := declined(t, s)
		r, err := s.Reopen(ctx, doctor, r.ID, ReopenRequest{Reason: "Try another therapist"})
		if err != nil || r.Status != StatusPending || r.TherapistID != nil || r.DeclineReason != nil {
			t.Fatalf("err=%v r=%+v", err, r)
		}
		// Any therapist of the clinic can now accept it.
		if _, err := s.Accept(ctx, otherPhysio, r.ID); err != nil {
			t.Fatalf("accept after reopen: %v", err)
		}
	})

	t.Run("doctor re-opens to a named therapist", func(t *testing.T) {
		s, _ := newTestService(t)
		r := declined(t, s)
		r, err := s.Reopen(ctx, doctor, r.ID, ReopenRequest{TherapistID: ptrUint(otherPhysio.ID), Reason: "x"})
		if err != nil || *r.TherapistID != otherPhysio.ID {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("rules", func(t *testing.T) {
		s, _ := newTestService(t)
		r := declined(t, s)
		_, err := s.Reopen(ctx, physio, r.ID, ReopenRequest{Reason: "x"})
		assertErr(t, err, ErrForbidden)
		_, err = s.Reopen(ctx, doctor, r.ID, ReopenRequest{Reason: " "})
		assertErr(t, err, &ValidationError{})
		_, err = s.Reopen(ctx, doctor, r.ID, ReopenRequest{TherapistID: ptrUint(farPhysio.ID), Reason: "x"})
		assertErr(t, err, &ValidationError{})

		open := mustCreate(t, s, nil)
		_, err = s.Reopen(ctx, doctor, open.ID, ReopenRequest{Reason: "x"})
		assertErr(t, err, ErrInvalidTransition)
	})
}

func TestCountsAndFilters(t *testing.T) {
	ctx := context.Background()
	s, _ := newTestService(t)
	a := mustCreate(t, s, ptrUint(physio.ID))
	_, _ = s.Accept(ctx, physio, a.ID)
	mustCreate(t, s, nil) // open inbox item

	urgent := validCreate()
	urgent.Priority = PriorityUrgent
	if _, err := s.Create(ctx, doctor, urgent); err != nil {
		t.Fatal(err)
	}

	counts, err := s.Counts(ctx, physio, ListFilter{})
	if err != nil || counts[StatusAccepted] != 1 || counts[StatusPending] != 2 {
		t.Fatalf("physio counts: %v %v", counts, err)
	}
	if _, err := s.Counts(ctx, nurse, ListFilter{}); !errors.Is(err, ErrForbidden) {
		t.Errorf("nurse counts: want forbidden, got %v", err)
	}
	if _, total, _ := s.List(ctx, doctor, ListFilter{Priority: PriorityUrgent}); total != 1 {
		t.Errorf("priority filter: want 1, got %d", total)
	}
	if _, total, _ := s.List(ctx, doctor, ListFilter{AssignedTo: physio.ID}); total != 1 {
		t.Errorf("therapist filter: want 1, got %d", total)
	}
}

// recordingNotifier remembers the events it receives.
type recordingNotifier struct{ events []Event }

func (r *recordingNotifier) ReferralEvent(_ context.Context, e Event) {
	r.events = append(r.events, e)
}

func (r *recordingNotifier) kinds() []EventKind {
	out := make([]EventKind, len(r.events))
	for i, e := range r.events {
		out[i] = e.Kind
	}
	return out
}

func TestEventsAreEmittedOnSuccessOnly(t *testing.T) {
	ctx := context.Background()
	s, _ := newTestService(t)
	rec := &recordingNotifier{}
	s.WithNotifier(rec)

	r := mustCreate(t, s, ptrUint(physio.ID))
	_, _ = s.Accept(ctx, otherPhysio, r.ID) // fails: not visible → no event
	_, _ = s.Accept(ctx, physio, r.ID)
	_, _ = s.AddNote(ctx, doctor, r.ID, NoteRequest{Body: "How is it going?"})
	_, _ = s.Reassign(ctx, physio, r.ID, ReassignRequest{TherapistID: otherPhysio.ID, Reason: "Leave"})
	_, _ = s.Start(ctx, otherPhysio, r.ID) // starting does not notify
	_, _ = s.Complete(ctx, otherPhysio, r.ID, "Recovered")

	want := []EventKind{EventCreated, EventAccepted, EventMessage, EventReassigned, EventCompleted}
	got := rec.kinds()
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
	hand := rec.events[3]
	if hand.PreviousTherapistID == nil || *hand.PreviousTherapistID != physio.ID || hand.ActorID != physio.ID ||
		hand.Excerpt != "Leave" || *hand.Referral.TherapistID != otherPhysio.ID {
		t.Errorf("hand-over event: %+v", hand)
	}
	if rec.events[2].Excerpt != "How is it going?" {
		t.Errorf("message excerpt: %q", rec.events[2].Excerpt)
	}
}

func TestHappyPathLifecycle(t *testing.T) {
	ctx := context.Background()
	s, repo := newTestService(t)
	r := mustCreate(t, s, nil)

	r, err := s.Accept(ctx, physio, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != StatusAccepted || r.TherapistID == nil || *r.TherapistID != physio.ID || !r.AcceptedAt.Equal(fixedNow) {
		t.Fatalf("after accept: %+v", r)
	}

	if r, err = s.Start(ctx, physio, r.ID); err != nil || r.Status != StatusInProgress {
		t.Fatalf("start: %v %+v", err, r)
	}
	if r, err = s.Complete(ctx, physio, r.ID, "Full ROM restored"); err != nil || r.Status != StatusCompleted {
		t.Fatalf("complete: %v %+v", err, r)
	}
	if r.OutcomeSummary == nil || *r.OutcomeSummary != "Full ROM restored" {
		t.Fatalf("summary not stored: %+v", r)
	}

	// Every transition leaves a system note on the thread.
	notes, _ := repo.ListNotes(ctx, r.ID)
	if len(notes) != 3 {
		t.Fatalf("want 3 system notes, got %d", len(notes))
	}
}

func TestAcceptRules(t *testing.T) {
	ctx := context.Background()

	t.Run("addressed referral can only be accepted by that therapist", func(t *testing.T) {
		s, _ := newTestService(t)
		r := mustCreate(t, s, ptrUint(physio.ID))
		// otherPhysio cannot even see it, so it looks like it does not exist.
		_, err := s.Accept(ctx, otherPhysio, r.ID)
		assertErr(t, err, ErrNotFound)
		if _, err := s.Accept(ctx, physio, r.ID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("doctor cannot accept", func(t *testing.T) {
		s, _ := newTestService(t)
		r := mustCreate(t, s, nil)
		_, err := s.Accept(ctx, doctor, r.ID)
		assertErr(t, err, ErrForbidden)
	})

	t.Run("second accept is rejected", func(t *testing.T) {
		s, _ := newTestService(t)
		r := mustCreate(t, s, nil)
		if _, err := s.Accept(ctx, physio, r.ID); err != nil {
			t.Fatal(err)
		}
		// Now assigned to physio, so otherPhysio no longer sees it.
		_, err := s.Accept(ctx, otherPhysio, r.ID)
		assertErr(t, err, ErrNotFound)
		_, err = s.Accept(ctx, physio, r.ID)
		assertErr(t, err, ErrInvalidTransition)
	})
}

func TestInvalidTransitions(t *testing.T) {
	ctx := context.Background()
	s, _ := newTestService(t)
	r := mustCreate(t, s, ptrUint(physio.ID))

	_, err := s.Start(ctx, physio, r.ID) // must accept first
	assertErr(t, err, ErrInvalidTransition)
	_, err = s.Complete(ctx, physio, r.ID, "done")
	assertErr(t, err, ErrInvalidTransition)
	_, err = s.Complete(ctx, physio, r.ID, "  ")
	assertErr(t, err, &ValidationError{})
}

func TestDecline(t *testing.T) {
	ctx := context.Background()

	t.Run("assigned therapist declines with reason", func(t *testing.T) {
		s, _ := newTestService(t)
		r := mustCreate(t, s, ptrUint(physio.ID))
		_, err := s.Decline(ctx, physio, r.ID, "")
		assertErr(t, err, &ValidationError{})
		r, err = s.Decline(ctx, physio, r.ID, "No capacity this month")
		if err != nil || r.Status != StatusDeclined {
			t.Fatalf("decline: %v %+v", err, r)
		}
	})

	t.Run("unassigned referral cannot be declined on behalf of everyone", func(t *testing.T) {
		s, _ := newTestService(t)
		r := mustCreate(t, s, nil)
		_, err := s.Decline(ctx, physio, r.ID, "busy")
		assertErr(t, err, ErrForbidden)
	})
}

func TestCancel(t *testing.T) {
	ctx := context.Background()

	t.Run("only the referring doctor or admin", func(t *testing.T) {
		s, _ := newTestService(t)
		r := mustCreate(t, s, nil)
		_, err := s.Cancel(ctx, otherDoctor, r.ID, "x")
		assertErr(t, err, ErrNotFound) // not involved → hidden
		if _, err := s.Cancel(ctx, admin, r.ID, "duplicate"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("cannot cancel a completed referral", func(t *testing.T) {
		s, _ := newTestService(t)
		r := mustCreate(t, s, ptrUint(physio.ID))
		_, _ = s.Accept(ctx, physio, r.ID)
		_, _ = s.Start(ctx, physio, r.ID)
		_, _ = s.Complete(ctx, physio, r.ID, "ok")
		_, err := s.Cancel(ctx, doctor, r.ID, "late")
		assertErr(t, err, ErrInvalidTransition)
	})
}

func TestVisibilityAndList(t *testing.T) {
	ctx := context.Background()
	s, _ := newTestService(t)
	open := mustCreate(t, s, nil)
	mine := mustCreate(t, s, ptrUint(physio.ID))
	theirs := mustCreate(t, s, ptrUint(otherPhysio.ID))

	if _, err := s.Get(ctx, physio, open.ID); err != nil {
		t.Errorf("therapist should see open referral: %v", err)
	}
	if _, err := s.Get(ctx, physio, theirs.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("therapist should not see another's referral, got %v", err)
	}

	items, total, err := s.List(ctx, physio, ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("physio inbox: want 2 (open + mine), got %d", total)
	}
	_ = items
	_ = mine

	if _, total, _ := s.List(ctx, doctor, ListFilter{}); total != 3 {
		t.Errorf("doctor sees own referrals: want 3, got %d", total)
	}
	if _, total, _ := s.List(ctx, otherDoctor, ListFilter{}); total != 0 {
		t.Errorf("other doctor: want 0, got %d", total)
	}
	if _, _, err := s.List(ctx, nurse, ListFilter{}); !errors.Is(err, ErrForbidden) {
		t.Errorf("nurse list: want forbidden, got %v", err)
	}
}

func TestNotes(t *testing.T) {
	ctx := context.Background()
	s, _ := newTestService(t)
	r := mustCreate(t, s, ptrUint(physio.ID))

	if _, err := s.AddNote(ctx, doctor, r.ID, NoteRequest{Kind: NoteQuestion, Body: "Any knee effusion?"}); err != nil {
		t.Fatal(err)
	}
	n, err := s.AddNote(ctx, physio, r.ID, NoteRequest{Body: "Mild, improving"})
	if err != nil {
		t.Fatal(err)
	}
	if n.Kind != NoteProgress {
		t.Errorf("default kind: want progress, got %s", n.Kind)
	}
	_, err = s.AddNote(ctx, physio, r.ID, NoteRequest{Body: "   "})
	assertErr(t, err, &ValidationError{})

	notes, err := s.Notes(ctx, doctor, r.ID)
	if err != nil || len(notes) != 2 {
		t.Fatalf("notes: %v len=%d", err, len(notes))
	}
}

func TestListPaginationBounds(t *testing.T) {
	s, _ := newTestService(t)
	repo := &capturingRepo{fakeRepo: newFakeRepo()}
	s.repo = repo
	_, _, _ = s.List(context.Background(), admin, ListFilter{Limit: 10_000, Offset: -5})
	if repo.last.Limit != maxLimit || repo.last.Offset != 0 {
		t.Fatalf("bounds not applied: %+v", repo.last)
	}
}

// capturingRepo records the last list filter.
type capturingRepo struct {
	*fakeRepo
	last ListFilter
}

func (c *capturingRepo) List(ctx context.Context, f ListFilter) ([]Referral, int64, error) {
	c.last = f
	return c.fakeRepo.List(ctx, f)
}

// assertErr checks err against a sentinel error or, for *ValidationError, its type.
func assertErr(t *testing.T, err error, want any) {
	t.Helper()
	switch w := want.(type) {
	case *ValidationError:
		var v *ValidationError
		if !errors.As(err, &v) {
			t.Fatalf("want ValidationError, got %v", err)
		}
	case error:
		if !errors.Is(err, w) {
			t.Fatalf("want %v, got %v", w, err)
		}
	}
}
