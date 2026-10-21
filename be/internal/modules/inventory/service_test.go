package inventory

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

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
		case "archived":
			it.Archived = v.(bool)
		}
	}
	f.items[id] = it
	return nil
}

func (f *fakeRepo) Apply(_ context.Context, id string, m *Movement) (*Item, error) {
	it := f.items[id]
	next, delta, err := nextQuantity(it.Quantity, m.Kind, m.Delta)
	if err != nil {
		return nil, err
	}
	it.Quantity = next
	f.items[id] = it
	m.ItemID, m.Delta, m.QuantityAfter = id, delta, next
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
		if it.ClinicID == clinicID && it.PerSession > 0 && !it.Archived {
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

var (
	clinicA = uuid.NewString()
	clinicB = uuid.NewString()

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

func TestStockMovements(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	gloves, err := s.Create(ctx, physio, ItemRequest{Name: " Gloves ", Category: "PPE", Unit: "pairs", Quantity: 10, MinQuantity: 4})
	if err != nil {
		t.Fatal(err)
	}
	if gloves.Name != "Gloves" || gloves.ClinicID != clinicA || gloves.Status != "ok" || len(repo.movements) != 1 {
		t.Fatalf("created: %+v (movements %d)", gloves, len(repo.movements))
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
		if err != nil {
			t.Fatalf("%s %d: %v", st.kind, st.qty, err)
		}
		if got.Quantity != st.want || got.Status != st.wantStatus {
			t.Fatalf("%s %d: quantity %d status %s, want %d %s", st.kind, st.qty, got.Quantity, got.Status, st.want, st.wantStatus)
		}
	}

	if _, err := s.Move(ctx, physio, gloves.ID, MovementRequest{Kind: KindUse, Quantity: 1}); !errors.Is(err, ErrInsufficient) {
		t.Errorf("using more than in stock: %v", err)
	}
	var verr *ValidationError
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

	sum, _ := s.Summary(ctx, physio)
	if sum != (Summary{Items: 1, Out: 1}) {
		t.Errorf("summary: %+v", sum)
	}
}

func TestInventoryPermissions(t *testing.T) {
	ctx := context.Background()
	s, _ := newSvc()
	it, _ := s.Create(ctx, physio, ItemRequest{Name: "Ultrasound gel", Quantity: 3})

	if _, err := s.Get(ctx, otherPhysio, it.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("another clinic must not see the item: %v", err)
	}
	if _, err := s.List(ctx, homeless, ListFilter{}); !errors.Is(err, ErrNoClinic) {
		t.Errorf("user without clinic: %v", err)
	}
	if _, err := s.Move(ctx, readOnly, it.ID, MovementRequest{Kind: KindUse, Quantity: 1}); !errors.Is(err, ErrForbidden) {
		t.Errorf("read-only use: %v", err)
	}
	if err := s.Archive(ctx, physio, it.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("physio archive: %v", err)
	}
	if _, err := s.Create(ctx, physio, ItemRequest{Name: "ULTRASOUND GEL"}); err == nil {
		t.Error("duplicate name in the same clinic should fail")
	}
	if _, err := s.Create(ctx, otherPhysio, ItemRequest{Name: "Ultrasound gel"}); err != nil {
		t.Errorf("same name in another clinic is fine: %v", err)
	}
	if _, err := s.Create(ctx, admin, ItemRequest{Name: "Tape"}); err == nil {
		t.Error("admins must name the clinic")
	}
	if _, err := s.Create(ctx, admin, ItemRequest{Name: "Tape", ClinicID: &clinicB}); err != nil {
		t.Errorf("admin with clinic: %v", err)
	}
	if _, err := s.Create(ctx, physio, ItemRequest{Name: "Tape", ClinicID: &clinicB}); !errors.Is(err, ErrForbidden) {
		t.Errorf("staff creating in another clinic: %v", err)
	}
	if err := s.Archive(ctx, coordinator, it.ID); err != nil {
		t.Fatalf("coordinator archive: %v", err)
	}
	if items, _ := s.List(ctx, physio, ListFilter{}); len(items) != 0 {
		t.Errorf("archived items disappear from the list: %d", len(items))
	}
	if items, _ := s.List(ctx, admin, ListFilter{}); len(items) != 2 {
		t.Errorf("admin sees every clinic: %d", len(items))
	}
}

func TestSessionUseAndAudit(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	gloves, _ := s.Create(ctx, physio, ItemRequest{Name: "Gloves", Quantity: 3, PerSession: 1})
	gel, _ := s.Create(ctx, physio, ItemRequest{Name: "Gel", Quantity: 0, PerSession: 1})
	if _, err := s.Create(ctx, physio, ItemRequest{Name: "Masks", Quantity: 10}); err != nil {
		t.Fatal(err)
	}
	appt := uuid.NewString()

	warnings, err := s.SessionAttended(ctx, clinicA, appt, physio.ID)
	if err != nil || len(warnings) != 1 || !strings.HasPrefix(warnings[0], "Gel") {
		t.Fatalf("attended: %v %v", warnings, err)
	}
	if repo.items[gloves.ID].Quantity != 2 || repo.items[gel.ID].Quantity != 0 {
		t.Fatalf("one pair of gloves used, gel short: %d / %d", repo.items[gloves.ID].Quantity, repo.items[gel.ID].Quantity)
	}
	if _, err := s.SessionAttended(ctx, clinicA, appt, physio.ID); err != nil || repo.items[gloves.ID].Quantity != 2 {
		t.Errorf("the same session uses stock once: %d", repo.items[gloves.ID].Quantity)
	}
	if err := s.SessionUndone(ctx, clinicA, appt, physio.ID); err != nil || repo.items[gloves.ID].Quantity != 3 {
		t.Errorf("undone gives the gloves back: %d %v", repo.items[gloves.ID].Quantity, err)
	}

	// Opening a pack is its own audited movement
	if it, err := s.Move(ctx, coordinator, gloves.ID, MovementRequest{Kind: KindOpen, Quantity: 1, Note: "new box"}); err != nil || it.Quantity != 2 {
		t.Fatalf("open: %+v %v", it, err)
	}

	if _, err := s.Audit(ctx, physio, AuditFilter{}); !errors.Is(err, ErrForbidden) {
		t.Errorf("physios do not see the audit log: %v", err)
	}
	opened, err := s.Audit(ctx, coordinator, AuditFilter{Kind: KindOpen})
	if err != nil || len(opened) != 1 || opened[0].UserID != coordinator.ID {
		t.Errorf("audit of openings: %+v %v", opened, err)
	}
	byPhysio, _ := s.Audit(ctx, coordinator, AuditFilter{UserID: physio.ID})
	if len(byPhysio) != 4 { // 2 opening stocks + session use + undo
		t.Errorf("movements by the physio: %d", len(byPhysio))
	}
}

func TestStatusFilter(t *testing.T) {
	ctx := context.Background()
	s, _ := newSvc()
	for name, q := range map[string]int{"Masks": 50, "Towels": 2, "Wipes": 0} {
		if _, err := s.Create(ctx, physio, ItemRequest{Name: name, Quantity: q, MinQuantity: 5}); err != nil {
			t.Fatal(err)
		}
	}
	attention, _ := s.List(ctx, physio, ListFilter{Status: "attention"})
	if len(attention) != 2 {
		t.Errorf("low + out = %d", len(attention))
	}
}
