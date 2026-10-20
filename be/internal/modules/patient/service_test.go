package patient

import (
	"context"
	"errors"
	"testing"
	"time"

	"klyniva/internal/models"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory Repository. It records the scope it was called
// with; visibility itself is enforced in SQL and covered by integration tests.
type fakeRepo struct {
	patients  map[string]models.Patient
	homes     map[uint]*string
	lastScope Scope
	lastQuery ListQuery
	updates   map[string]any
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{patients: map[string]models.Patient{}, homes: map[uint]*string{}}
}

func (f *fakeRepo) List(_ context.Context, s Scope, q ListQuery) ([]models.Patient, int64, error) {
	f.lastScope, f.lastQuery = s, q
	var out []models.Patient
	for _, p := range f.patients {
		out = append(out, p)
	}
	return out, int64(len(out)), nil
}

func (f *fakeRepo) Get(_ context.Context, s Scope, id string) (*models.Patient, error) {
	f.lastScope = s
	p, ok := f.patients[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &p, nil
}

func (f *fakeRepo) Create(_ context.Context, p *models.Patient) error {
	_ = p.BeforeCreate(nil)
	f.patients[p.ID] = *p
	return nil
}

func (f *fakeRepo) Update(_ context.Context, id string, fields map[string]any) error {
	if _, ok := f.patients[id]; !ok {
		return ErrNotFound
	}
	f.updates = fields
	return nil
}

func (f *fakeRepo) HomeFacility(_ context.Context, userID uint) (*string, error) {
	return f.homes[userID], nil
}

var (
	nurse    = Actor{ID: 5, Role: models.RoleNurse}
	readOnly = Actor{ID: 6, Role: models.RoleStaffReadOnly}
	admin    = Actor{ID: 1, Role: models.RoleStateAdmin}
	now      = time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
)

func newSvc() (*Service, *fakeRepo) {
	repo := newFakeRepo()
	return NewService(repo, func() time.Time { return now }), repo
}

func strp(s string) *string { return &s }

func TestCreateSetsServerFields(t *testing.T) {
	s, repo := newSvc()
	home := uuid.NewString()
	repo.homes[nurse.ID] = &home

	p, err := s.Create(context.Background(), nurse, CreateRequest{
		Name: "  Ana Pereira ", DateOfBirth: strp("1980-05-02"), PhoneNumber: strp("+351 912 345 678"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Ana Pereira" || p.CreatedByID == nil || *p.CreatedByID != nurse.ID {
		t.Errorf("server fields: %+v", p)
	}
	if p.FacilityID == nil || *p.FacilityID != home {
		t.Error("facility should default to the user's home facility")
	}
	if p.YearOfBirth == nil || *p.YearOfBirth != 1980 {
		t.Error("year_of_birth should be derived from date_of_birth")
	}
}

func TestCreateValidation(t *testing.T) {
	cases := []struct {
		name  string
		actor Actor
		req   CreateRequest
		want  string // expected ValidationError field, or "" for ErrForbidden
	}{
		{"read-only role", readOnly, CreateRequest{Name: "Ana"}, ""},
		{"short name", nurse, CreateRequest{Name: "A"}, "name"},
		{"future birth date", nurse, CreateRequest{Name: "Ana", DateOfBirth: strp("2030-01-01")}, "date_of_birth"},
		{"bad date", nurse, CreateRequest{Name: "Ana", DateOfBirth: strp("02/05/1980")}, "date_of_birth"},
		{"bad phone", nurse, CreateRequest{Name: "Ana", PhoneNumber: strp("call me")}, "phone_number"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newSvc()
			_, err := s.Create(context.Background(), tc.actor, tc.req)
			if tc.want == "" {
				if !errors.Is(err, ErrForbidden) {
					t.Fatalf("want forbidden, got %v", err)
				}
				return
			}
			var v *ValidationError
			if !errors.As(err, &v) || v.Field != tc.want {
				t.Fatalf("want validation error on %s, got %v", tc.want, err)
			}
		})
	}
}

func TestScopeAndPagination(t *testing.T) {
	s, repo := newSvc()
	ctx := context.Background()

	_, _, _ = s.List(ctx, admin, ListQuery{Limit: 1000})
	if !repo.lastScope.All || repo.lastQuery.Limit != maxLimit {
		t.Errorf("admin scope/limit: %+v %+v", repo.lastScope, repo.lastQuery)
	}

	_, _, _ = s.List(ctx, nurse, ListQuery{Offset: -3})
	if repo.lastScope.All || repo.lastScope.UserID != nurse.ID || repo.lastQuery.Offset != 0 {
		t.Errorf("nurse scope: %+v %+v", repo.lastScope, repo.lastQuery)
	}

	if _, err := s.Get(ctx, nurse, "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Errorf("invalid id should be not found, got %v", err)
	}
}

func TestUpdateWhitelistAndCoercion(t *testing.T) {
	s, _ := newSvc()
	ctx := context.Background()
	p, _ := s.Create(ctx, nurse, CreateRequest{Name: "Ana"})

	s2, repo := s, s.repo.(*fakeRepo)
	_, err := s2.Update(ctx, nurse, p.ID, map[string]any{
		"name":          "Ana Maria",
		"gender":        float64(2),
		"is_active":     false,
		"id":            "attacker-chosen", // ignored
		"created_by_id": float64(999),      // ignored
		"allergies":     nil,               // clears the field
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.updates["name"] != "Ana Maria" || repo.updates["gender"] != int64(2) || repo.updates["is_active"] != false {
		t.Errorf("coerced fields: %+v", repo.updates)
	}
	if _, ok := repo.updates["id"]; ok {
		t.Error("id must not be updatable")
	}
	if _, ok := repo.updates["allergies"]; !ok {
		t.Error("null should clear allergies")
	}
	if repo.updates["last_edited_id"] != nurse.ID {
		t.Error("last_edited_id not set")
	}

	bad := []map[string]any{
		{"gender": "female"},
		{"gender": 1.5},
		{"facility": "nope"},
		{"name": nil},
		{"name": " x "},
		{"phone_number": "abc"},
	}
	for _, body := range bad {
		var v *ValidationError
		if _, err := s.Update(ctx, nurse, p.ID, body); !errors.As(err, &v) {
			t.Errorf("%v: want validation error, got %v", body, err)
		}
	}

	if _, err := s.Update(ctx, readOnly, p.ID, map[string]any{"name": "Bob"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("read-only update: want forbidden, got %v", err)
	}
}

func TestEscapeLike(t *testing.T) {
	if got := escapeLike(`50%_off\`); got != `50\%\_off\\` {
		t.Errorf("escapeLike: %q", got)
	}
}
