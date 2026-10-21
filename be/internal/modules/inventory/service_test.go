package inventory

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory Repository applying the same stock rules.
type fakeRepo struct {
	items     map[string]Item
	movements []Movement
}

func (f *fakeRepo) visible(clinicIDs []string, it Item) bool {
	return !it.Archived && (clinicIDs == nil || slices.Contains(clinicIDs, it.ClinicID))
}

func (f *fakeRepo) List(_ context.Context, clinicIDs []string, fl ListFilter) ([]Item, error) {
	var out []Item
	for _, it := range f.items {
		if f.visible(clinicIDs, it) && strings.Contains(strings.ToLower(it.Name), strings.ToLower(fl.Search)) &&
			(fl.Status == "" || it.Status() == fl.Status || (fl.Status == "attention" && it.Status() != "ok")) {
			out = append(out, it)
		}
	}
	return out, nil
}

func (f *fakeRepo) Get(_ context.Context, clinicIDs []string, id string) (*Item, error) {
	it, ok := f.items[id]
	if !ok || !f.visible(clinicIDs, it) {
		return nil, ErrNotFound
	}
	return &it, nil
}

func (f *fakeRepo) Create(_ context.Context, it *Item, opening *Movement) error {
	it.ID = uuid.NewString()
	f.items[it.ID] = *it
	if opening != nil {
		opening.ItemID = it.ID
		f.movements = append(f.movements, *opening)
	}
	return nil
}

func (f *fakeRepo) Update(_ context.Context, id string, fields map[string]any) error {
	it := f.items[id]
	for k, v := range fields {
		switch k {
		case "name":
			it.Name = v.(string)
		case "min_quantity":
			it.MinQuantity = v.(int)
		case "per_session":
			it.PerSession = v.(int)
		case "tracking":
			it.Tracking = v.(string)
		case "archived":
			it.Archived = v.(bool)
		}
	}
	f.items[id] = it
	return nil
}

func (f *fakeRepo) Apply(_ context.Context, id string, m *Movement) (*Item, error) {
	it := f.items[id]
	next, err := nextState(it.Quantity, it.InUse, m.Kind, m.Amount)
	if err != nil {
		return nil, err
	}
	it.Quantity, it.InUse = next.quantity, next.inUse
	f.items[id] = it
	m.ItemID, m.Delta, m.QuantityAfter, m.InUseAfter = id, next.delta, next.quantity, next.inUse
	f.movements = append(f.movements, *m)
	return &it, nil
}

func (f *fakeRepo) Movements(_ context.Context, id string, _ int) ([]Movement, error) {
	var out []Movement
	for _, m := range f.movements {
		if m.ItemID == id {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *fakeRepo) Audit(_ context.Context, clinicIDs []string, fl AuditFilter, _ int) ([]Movement, error) {
	var out []Movement
	for _, m := range f.movements {
		it := f.items[m.ItemID]
		if (clinicIDs == nil || slices.Contains(clinicIDs, it.ClinicID)) &&
			(fl.UserID == 0 || m.UserID == fl.UserID) && (fl.Kind == "" || m.Kind == fl.Kind) {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *fakeRepo) SessionItems(_ context.Context, clinicID string) ([]Item, error) {
	var out []Item
	for _, it := range f.items {
		hasRule := slices.ContainsFunc(it.Rules, func(r SessionRule) bool { return r.Quantity > 0 })
		if it.ClinicID == clinicID && it.Tracking == TrackUnit && !it.Archived && (it.PerSession > 0 || hasRule) {
			out = append(out, it)
		}
	}
	return out, nil
}

func (f *fakeRepo) NetForAppointment(_ context.Context, id string) (map[string]int, error) {
	out := map[string]int{}
	for _, m := range f.movements {
		if m.AppointmentID != nil && *m.AppointmentID == id {
			out[m.ItemID] += m.Delta
		}
	}
	return out, nil
}

func (f *fakeRepo) ReplaceRules(_ context.Context, id string, rules []SessionRule) error {
	it := f.items[id]
	it.Rules = rules
	f.items[id] = it
	return nil
}

var (
	clinicA = uuid.NewString()
	clinicB = uuid.NewString()
	gym     = uuid.NewString()
	pool    = uuid.NewString()

	physio      = Actor{ID: 1, Role: models.RolePhysioTherapist}
	coordinator = Actor{ID: 2, Role: models.RoleCoordinator}
	otherPhysio = Actor{ID: 3, Role: models.RolePhysioTherapist} // clinic B
	readOnly    = Actor{ID: 4, Role: models.RoleStaffReadOnly}
	homeless    = Actor{ID: 5, Role: models.RoleNurse} // no clinic
	admin       = Actor{ID: 9, Role: models.RoleStateAdmin}
)

func newSvc() (*Service, *fakeRepo) {
	repo := &fakeRepo{items: map[string]Item{}}
	clinics := clinic.Static{physio.ID: clinicA, coordinator.ID: clinicA, readOnly.ID: clinicA, otherPhysio.ID: clinicB}
	return NewService(repo, clinics), repo
}

func TestUnitItems(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	gloves, err := s.Create(ctx, physio, ItemRequest{Name: " Gloves ", Category: "PPE", Unit: "pairs", Quantity: 10, MinQuantity: 4})
	if err != nil {
		t.Fatal(err)
	}
	if gloves.Name != "Gloves" || gloves.Tracking != TrackUnit || gloves.Status != "ok" || len(repo.movements) != 1 {
		t.Fatalf("created: %+v", gloves)
	}
	steps := []struct {
		kind       string
		qty, want  int
		wantStatus string
	}{
		{KindUse, 6, 4, "low"},
		{KindRestock, 20, 24, "ok"},
		{KindAdjust, 0, 0, "out"},
	}
	for _, st := range steps {
		got, err := s.Move(ctx, physio, gloves.ID, MovementRequest{Kind: st.kind, Quantity: st.qty})
		if err != nil || got.Quantity != st.want || got.Status != st.wantStatus {
			t.Fatalf("%s %d: %+v %v", st.kind, st.qty, got, err)
		}
	}
	if _, err := s.Move(ctx, physio, gloves.ID, MovementRequest{Kind: KindUse, Quantity: 1}); !errors.Is(err, ErrInsufficient) {
		t.Errorf("using more than in stock: %v", err)
	}
	var verr *ValidationError
	for _, kind := range []string{KindOpen, KindFinish} {
		if _, err := s.Move(ctx, physio, gloves.ID, MovementRequest{Kind: kind, Quantity: 1}); !errors.As(err, &verr) {
			t.Errorf("single-use items are not %s: %v", kind, err)
		}
	}
	if _, err := s.Move(ctx, physio, gloves.ID, MovementRequest{Kind: KindRestock, Quantity: 0}); !errors.As(err, &verr) {
		t.Errorf("restock of 0: %v", err)
	}
	hist, _ := s.History(ctx, coordinator, gloves.ID)
	deltas := []int{}
	for _, m := range hist {
		deltas = append(deltas, m.Delta)
	}
	if !slices.Equal(deltas, []int{10, -6, 20, -24}) {
		t.Errorf("history deltas = %v", deltas)
	}
}

func TestContainers(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	now := time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	gel, err := s.Create(ctx, physio, ItemRequest{Name: "Ultrasound gel", Unit: "bottles", Tracking: TrackContainer,
		Quantity: 3, MinQuantity: 1, OpenLifeDays: 30})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, physio, ItemRequest{Name: "Oil", Tracking: TrackContainer, PerSession: 1}); err == nil {
		t.Error("containers are not used per session")
	}

	opened, err := s.Move(ctx, physio, gel.ID, MovementRequest{Kind: KindOpen, Quantity: 1, Note: "room 1"})
	if err != nil || opened.Quantity != 2 || opened.InUse != 1 {
		t.Fatalf("open: %+v %v", opened, err)
	}
	var verr *ValidationError
	if _, err := s.Move(ctx, physio, gel.ID, MovementRequest{Kind: KindUse, Quantity: 1}); !errors.As(err, &verr) {
		t.Errorf("containers are opened, not used: %v", err)
	}
	finished, err := s.Move(ctx, coordinator, gel.ID, MovementRequest{Kind: KindFinish, Quantity: 1})
	if err != nil || finished.Quantity != 2 || finished.InUse != 0 {
		t.Fatalf("finish keeps sealed stock, empties the open one: %+v %v", finished, err)
	}
	if _, err := s.Move(ctx, physio, gel.ID, MovementRequest{Kind: KindFinish, Quantity: 1}); !errors.Is(err, ErrNothingOpen) {
		t.Errorf("finishing with nothing open: %v", err)
	}

	// Open life: a container open for 31 of its 30 days is flagged
	it := repo.items[gel.ID]
	it.InUse = 1
	opened31 := now.AddDate(0, 0, -31)
	it.OldestOpenAt = &opened31
	repo.items[gel.ID] = it
	list, _ := s.List(ctx, physio, ListFilter{})
	if !list[0].OpenExpired {
		t.Error("expired open container not flagged")
	}
	if sum, _ := s.Summary(ctx, physio); sum.OpenExpired != 1 {
		t.Errorf("summary: %+v", sum)
	}
	if _, err := s.Update(ctx, physio, gel.ID, ItemRequest{Name: "Ultrasound gel", Tracking: TrackUnit}); !errors.As(err, &verr) {
		t.Errorf("changing kind with open containers: %v", err)
	}
}

func TestSessionUseByDivisionAndAudit(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	gloves, _ := s.Create(ctx, physio, ItemRequest{Name: "Gloves", Quantity: 10, PerSession: 1})
	caps, _ := s.Create(ctx, physio, ItemRequest{Name: "Swim caps", Quantity: 0}) // only in the pool
	if _, err := s.SetSessionRules(ctx, physio, caps.ID, SessionRulesRequest{}); !errors.Is(err, ErrForbidden) {
		t.Errorf("physios do not set rules: %v", err)
	}
	req := SessionRulesRequest{Rules: []RuleInput{{pool, 1}}}
	if _, err := s.SetSessionRules(ctx, coordinator, caps.ID, req); err != nil {
		t.Fatal(err)
	}
	// Gloves: 2 in the pool, 0 in the gym
	req.Rules[0].Quantity = 2
	req.Rules = append(req.Rules, req.Rules[0])
	req.Rules[1].DivisionID, req.Rules[1].Quantity = gym, 0
	if _, err := s.SetSessionRules(ctx, coordinator, gloves.ID, req); err != nil {
		t.Fatal(err)
	}

	gymSession, poolSession := uuid.NewString(), uuid.NewString()
	if w, err := s.SessionAttended(ctx, clinicA, gymSession, gym, physio.ID); err != nil || len(w) != 0 {
		t.Fatalf("gym session: %v %v", w, err)
	}
	if repo.items[gloves.ID].Quantity != 10 {
		t.Errorf("no gloves in the gym: %d", repo.items[gloves.ID].Quantity)
	}
	w, err := s.SessionAttended(ctx, clinicA, poolSession, pool, physio.ID)
	if err != nil || len(w) != 1 || !strings.HasPrefix(w[0], "Swim caps") {
		t.Fatalf("pool session: %v %v", w, err)
	}
	if repo.items[gloves.ID].Quantity != 8 {
		t.Errorf("2 pairs in the pool: %d", repo.items[gloves.ID].Quantity)
	}
	if _, err := s.SessionAttended(ctx, clinicA, poolSession, pool, physio.ID); err != nil || repo.items[gloves.ID].Quantity != 8 {
		t.Errorf("once per session: %d", repo.items[gloves.ID].Quantity)
	}
	// Another division without a rule uses the default (1)
	if _, err := s.SessionAttended(ctx, clinicA, uuid.NewString(), uuid.NewString(), physio.ID); err != nil || repo.items[gloves.ID].Quantity != 7 {
		t.Errorf("default per session: %d", repo.items[gloves.ID].Quantity)
	}
	if err := s.SessionUndone(ctx, clinicA, poolSession, physio.ID); err != nil || repo.items[gloves.ID].Quantity != 9 {
		t.Errorf("undo gives back: %d %v", repo.items[gloves.ID].Quantity, err)
	}

	if _, err := s.Audit(ctx, physio, AuditFilter{}); !errors.Is(err, ErrForbidden) {
		t.Errorf("physios do not see the audit log: %v", err)
	}
	used, err := s.Audit(ctx, coordinator, AuditFilter{Kind: KindUse})
	if err != nil || len(used) != 2 {
		t.Errorf("audit of use: %d %v", len(used), err)
	}
}

func TestReorderAndPermissions(t *testing.T) {
	ctx := context.Background()
	s, _ := newSvc()
	for name, q := range map[string]int{"Masks": 50, "Towels": 2, "Wipes": 0} {
		if _, err := s.Create(ctx, physio, ItemRequest{Name: name, Quantity: q, MinQuantity: 5}); err != nil {
			t.Fatal(err)
		}
	}
	tape, _ := s.Create(ctx, physio, ItemRequest{Name: "Tape", Quantity: 1, MinQuantity: 4, TargetQuantity: 20})
	order, _ := s.Reorder(ctx, coordinator)
	want := map[string]int{"Towels": 8, "Wipes": 10, "Tape": 19}
	if len(order) != 3 {
		t.Fatalf("reorder list: %+v", order)
	}
	for _, it := range order {
		if it.Reorder != want[it.Name] {
			t.Errorf("%s: order %d, want %d", it.Name, it.Reorder, want[it.Name])
		}
	}

	if _, err := s.Get(ctx, otherPhysio, tape.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("another clinic must not see the item: %v", err)
	}
	if _, err := s.List(ctx, homeless, ListFilter{}); !errors.Is(err, ErrNoClinic) {
		t.Errorf("user without clinic: %v", err)
	}
	if _, err := s.Move(ctx, readOnly, tape.ID, MovementRequest{Kind: KindUse, Quantity: 1}); !errors.Is(err, ErrForbidden) {
		t.Errorf("read-only use: %v", err)
	}
	if err := s.Archive(ctx, physio, tape.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("physio archive: %v", err)
	}
	if _, err := s.Create(ctx, physio, ItemRequest{Name: "TAPE"}); err == nil {
		t.Error("duplicate name in the same clinic should fail")
	}
	if _, err := s.Create(ctx, admin, ItemRequest{Name: "Gauze"}); err == nil {
		t.Error("admins must name the clinic")
	}
	if _, err := s.Create(ctx, physio, ItemRequest{Name: "Gauze", ClinicID: &clinicB}); !errors.Is(err, ErrForbidden) {
		t.Errorf("staff creating in another clinic: %v", err)
	}
	if err := s.Archive(ctx, coordinator, tape.ID); err != nil {
		t.Fatalf("coordinator archive: %v", err)
	}
}
