package referral

import (
	"context"
	"slices"
	"sync"
)

// fakeRepo is an in-memory Repository used by service and handler tests.
type fakeRepo struct {
	mu        sync.Mutex
	referrals map[string]Referral
	notes     []Note
	patients  map[string]bool
	users     map[uint]struct {
		role   string
		active bool
	}
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		referrals: map[string]Referral{},
		patients:  map[string]bool{},
		users: map[uint]struct {
			role   string
			active bool
		}{},
	}
}

func (f *fakeRepo) addUser(id uint, role string, active bool) {
	f.users[id] = struct {
		role   string
		active bool
	}{role, active}
}

func (f *fakeRepo) Create(_ context.Context, r *Referral) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	_ = r.BeforeCreate(nil)
	f.referrals[r.ID] = *r
	return nil
}

func (f *fakeRepo) Get(_ context.Context, id string) (*Referral, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.referrals[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &r, nil
}

func (f *fakeRepo) Counts(ctx context.Context, flt ListFilter) (map[Status]int64, error) {
	flt.Status = ""
	items, _, _ := f.List(ctx, flt)
	out := map[Status]int64{}
	for _, r := range items {
		out[r.Status]++
	}
	return out, nil
}

// List mirrors the SQL filters except the patient-name search, which needs
// the patients table and is covered by the PostgreSQL integration tests.
func (f *fakeRepo) List(_ context.Context, flt ListFilter) ([]Referral, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Referral
	for _, r := range f.referrals {
		if flt.Status != "" && r.Status != flt.Status {
			continue
		}
		if flt.Priority != "" && r.Priority != flt.Priority {
			continue
		}
		if flt.AssignedTo != 0 && (r.TherapistID == nil || *r.TherapistID != flt.AssignedTo) {
			continue
		}
		if flt.DoctorID != 0 && r.DoctorID != flt.DoctorID {
			continue
		}
		if flt.OnlyClinic != nil && (r.FacilityID == nil || *r.FacilityID != *flt.OnlyClinic) {
			continue
		}
		if flt.TherapistID != 0 {
			assigned := r.TherapistID != nil && (*r.TherapistID == flt.TherapistID || slices.Contains(flt.AlsoTherapists, *r.TherapistID))
			open := flt.InboxForTherapist && r.TherapistID == nil && r.Status == StatusPending &&
				(flt.ClinicID == nil || r.FacilityID == nil || *r.FacilityID == *flt.ClinicID)
			if !assigned && !open {
				continue
			}
		}
		if flt.OpenQuestions && !f.hasOpenQuestion(r.ID, flt.Viewer) {
			continue
		}
		out = append(out, r)
	}
	return out, int64(len(out)), nil
}

// hasOpenQuestion: an unanswered question others asked the viewer (lock held).
func (f *fakeRepo) hasOpenQuestion(referralID string, viewer uint) bool {
	for _, q := range f.notes {
		if q.ReferralID != referralID || q.Kind != NoteQuestion || q.AuthorID == viewer {
			continue
		}
		if !slices.ContainsFunc(f.notes, func(r Note) bool { return r.ReplyToID != nil && *r.ReplyToID == q.ID }) {
			return true
		}
	}
	return false
}

func (f *fakeRepo) Transition(_ context.Context, r *Referral, from Status, note *Note) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cur, ok := f.referrals[r.ID]
	if !ok {
		return ErrNotFound
	}
	if cur.Status != from {
		return ErrConflict
	}
	f.referrals[r.ID] = *r
	if note != nil {
		f.notes = append(f.notes, *note)
	}
	return nil
}

func (f *fakeRepo) AddNote(_ context.Context, n *Note) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	_ = n.BeforeCreate(nil)
	f.notes = append(f.notes, *n)
	return nil
}

func (f *fakeRepo) ListNotes(_ context.Context, referralID string) ([]Note, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Note
	for _, n := range f.notes {
		if n.ReferralID == referralID {
			out = append(out, n)
		}
	}
	return out, nil
}

func (f *fakeRepo) PatientExists(_ context.Context, id string) (bool, error) {
	return f.patients[id], nil
}

func (f *fakeRepo) UserRole(_ context.Context, id uint) (string, bool, error) {
	u, ok := f.users[id]
	if !ok {
		return "", false, nil
	}
	return u.role, u.active, nil
}

func (f *fakeRepo) GetNote(_ context.Context, referralID, id string) (*Note, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, n := range f.notes {
		if n.ReferralID == referralID && n.ID == id {
			c := n
			return &c, nil
		}
	}
	return nil, ErrNotFound
}

func (f *fakeRepo) OpenQuestions(_ context.Context, referralIDs []string, viewer uint) (map[string]int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	replied := map[string]bool{}
	for _, n := range f.notes {
		if n.ReplyToID != nil {
			replied[*n.ReplyToID] = true
		}
	}
	out := map[string]int{}
	for _, n := range f.notes {
		if n.Kind == NoteQuestion && n.AuthorID != viewer && !replied[n.ID] && slices.Contains(referralIDs, n.ReferralID) {
			out[n.ReferralID]++
		}
	}
	return out, nil
}
