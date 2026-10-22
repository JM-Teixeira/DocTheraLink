package staff

import (
	"context"
	"errors"
	"testing"

	"klyniva/internal/models"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory Repository.
type fakeRepo struct {
	users  map[uint]*models.User
	skills map[string]models.Skill
	links  []models.UserSkill
	clinic map[string]bool
}

func (f *fakeRepo) UserByUsername(_ context.Context, username string) (*models.User, error) {
	for _, u := range f.users {
		if u.Username == username {
			c := *u
			return &c, nil
		}
	}
	return nil, ErrNotFound
}
func (f *fakeRepo) UserByID(_ context.Context, id uint) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	c := *u
	return &c, nil
}
func (f *fakeRepo) UpdateUser(_ context.Context, id uint, fields map[string]any) error {
	u := f.users[id]
	for k, v := range fields {
		switch k {
		case "employment_status":
			u.EmploymentStatus = v.(string)
		case "is_active":
			u.IsActive = v.(bool)
		case "date_format":
			u.DateFormat = v.(string)
		case "home_facility_id":
			u.HomeFacilityID = v.(*string)
		}
	}
	return nil
}
func (f *fakeRepo) Skills(context.Context) ([]models.Skill, error) {
	var out []models.Skill
	for _, s := range f.skills {
		out = append(out, s)
	}
	return out, nil
}
func (f *fakeRepo) Skill(_ context.Context, id string) (*models.Skill, error) {
	s, ok := f.skills[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &s, nil
}
func (f *fakeRepo) UserSkills(_ context.Context, userID uint) ([]models.UserSkill, error) {
	var out []models.UserSkill
	for _, l := range f.links {
		if l.UserID == userID {
			out = append(out, l)
		}
	}
	return out, nil
}
func (f *fakeRepo) AddUserSkill(_ context.Context, userID uint, skillID string) (*models.UserSkill, error) {
	l := models.UserSkill{ID: uuid.NewString(), UserID: userID, SkillID: skillID, Skill: f.skills[skillID]}
	f.links = append(f.links, l)
	return &l, nil
}
func (f *fakeRepo) RemoveUserSkill(_ context.Context, userID uint, id string) error {
	for i, l := range f.links {
		if l.UserID == userID && (l.ID == id || l.SkillID == id) {
			f.links = append(f.links[:i], f.links[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
func (f *fakeRepo) Clinic(_ context.Context, id string) (*models.Facility, error) {
	if !f.clinic[id] {
		return nil, ErrNotFound
	}
	return &models.Facility{ID: id, Name: "Clinic"}, nil
}

var (
	clinicA, clinicB = uuid.NewString(), uuid.NewString()
	hydro            = uuid.NewString()
)

func newSvc() (*Service, *fakeRepo) {
	user := func(id uint, name string, role models.UserRole, clinic string) *models.User {
		c := clinic
		return &models.User{ID: id, Username: name, UserType: role, HomeFacilityID: &c, IsActive: true, EmploymentStatus: StatusActive}
	}
	repo := &fakeRepo{
		users: map[uint]*models.User{
			1: user(1, "sarah", models.RolePhysioTherapist, clinicA),
			2: user(2, "marta", models.RoleCoordinator, clinicA),
			3: user(3, "tiago", models.RolePhysioTherapist, clinicA),
			4: user(4, "far.coord", models.RoleCoordinator, clinicB),
			9: {ID: 9, Username: "admin", UserType: models.RoleStateAdmin, IsActive: true},
		},
		skills: map[string]models.Skill{hydro: {ID: hydro, Name: "Hidroterapia"}},
		clinic: map[string]bool{clinicA: true, clinicB: true},
	}
	return NewService(repo), repo
}

var (
	sarah     = Actor{ID: 1, Role: models.RolePhysioTherapist}
	marta     = Actor{ID: 2, Role: models.RoleCoordinator}
	tiago     = Actor{ID: 3, Role: models.RolePhysioTherapist}
	farCoord  = Actor{ID: 4, Role: models.RoleCoordinator}
	adminUser = Actor{ID: 9, Role: models.RoleStateAdmin}
)

func TestSkills(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	if _, err := s.AddSkill(ctx, sarah, "sarah", hydro); err != nil {
		t.Fatalf("own skill: %v", err)
	}
	if _, err := s.AddSkill(ctx, sarah, "sarah", hydro); err != nil || len(repo.links) != 1 {
		t.Errorf("adding twice keeps one link: %d %v", len(repo.links), err)
	}
	if _, err := s.AddSkill(ctx, marta, "tiago", hydro); err != nil {
		t.Errorf("coordinator of the clinic: %v", err)
	}
	for name, a := range map[string]Actor{"colleague": tiago, "other clinic": farCoord} {
		if _, err := s.AddSkill(ctx, a, "sarah", hydro); !errors.Is(err, ErrForbidden) {
			t.Errorf("%s: %v", name, err)
		}
	}
	var verr *ValidationError
	if _, err := s.AddSkill(ctx, sarah, "sarah", uuid.NewString()); !errors.As(err, &verr) {
		t.Errorf("unknown skill: %v", err)
	}
	if err := s.RemoveSkill(ctx, sarah, "sarah", hydro); err != nil {
		t.Errorf("remove by skill id: %v", err)
	}
	if got, _ := s.UserSkills(ctx, sarah, "sarah"); len(got) != 0 {
		t.Errorf("skills left: %d", len(got))
	}
}

func TestEmployment(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	u, err := s.SetEmployment(ctx, marta, "sarah", EmploymentRequest{Status: StatusLeft, Since: "2026-10-31", Note: "Moved away"})
	if err != nil || u.EmploymentStatus != StatusLeft || u.IsActive {
		t.Fatalf("left: %+v %v", u, err)
	}
	if u, err = s.SetEmployment(ctx, marta, "sarah", EmploymentRequest{Status: StatusActive}); err != nil || !u.IsActive {
		t.Errorf("back to active can sign in again: %v", err)
	}
	if _, err := s.SetEmployment(ctx, farCoord, "sarah", EmploymentRequest{Status: StatusOnLeave}); !errors.Is(err, ErrForbidden) {
		t.Errorf("other clinic's coordinator: %v", err)
	}
	if _, err := s.SetEmployment(ctx, tiago, "sarah", EmploymentRequest{Status: StatusOnLeave}); !errors.Is(err, ErrForbidden) {
		t.Errorf("colleague: %v", err)
	}
	var verr *ValidationError
	if _, err := s.SetEmployment(ctx, marta, "marta", EmploymentRequest{Status: StatusLeft}); !errors.As(err, &verr) {
		t.Errorf("own status: %v", err)
	}
	if _, err := s.SetEmployment(ctx, adminUser, "tiago", EmploymentRequest{Status: StatusOnLeave, Since: "31/10"}); !errors.As(err, &verr) {
		t.Errorf("bad date: %v", err)
	}
	if repo.users[3].EmploymentStatus != StatusActive {
		t.Error("a failed change must not be saved")
	}
}

func TestClinicLink(t *testing.T) {
	ctx := context.Background()
	s, _ := newSvc()
	if cl, _ := s.Clinics(ctx, sarah, "sarah"); len(cl) != 1 || cl[0].ID != clinicA {
		t.Fatalf("linked clinic: %+v", cl)
	}
	if _, err := s.SetClinic(ctx, marta, "sarah", clinicB); !errors.Is(err, ErrForbidden) {
		t.Errorf("only admins move people between clinics: %v", err)
	}
	if u, err := s.SetClinic(ctx, adminUser, "sarah", clinicB); err != nil || *u.HomeFacilityID != clinicB {
		t.Fatalf("admin moves: %v", err)
	}
	if _, err := s.SetClinic(ctx, adminUser, "sarah", uuid.NewString()); err == nil {
		t.Error("unknown clinic must fail")
	}
	if _, err := s.SetClinic(ctx, adminUser, "sarah", ""); err != nil {
		t.Fatal(err)
	}
	if cl, _ := s.Clinics(ctx, sarah, "sarah"); len(cl) != 0 {
		t.Errorf("cleared: %+v", cl)
	}
}

// Skills and clinic are visible to the person, their clinic and admins;
// to anyone else the user looks like it doesn't exist.
func TestVisibility(t *testing.T) {
	ctx := context.Background()
	s, _ := newSvc()
	for name, a := range map[string]Actor{"self": sarah, "colleague": tiago, "coordinator": marta, "admin": adminUser} {
		if _, err := s.UserSkills(ctx, a, "sarah"); err != nil {
			t.Errorf("%s reads skills: %v", name, err)
		}
		if _, err := s.Clinics(ctx, a, "sarah"); err != nil {
			t.Errorf("%s reads clinic: %v", name, err)
		}
	}
	if _, err := s.UserSkills(ctx, farCoord, "sarah"); !errors.Is(err, ErrNotFound) {
		t.Errorf("other clinic reads skills: %v", err)
	}
	if _, err := s.Clinics(ctx, farCoord, "sarah"); !errors.Is(err, ErrNotFound) {
		t.Errorf("other clinic reads clinic: %v", err)
	}
}

func TestPreferences(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	u, err := s.SetPreferences(ctx, sarah, "sarah", PreferencesRequest{DateFormat: "ymd"})
	if err != nil || u.DateFormat != "ymd" {
		t.Fatalf("own preference: %+v %v", u, err)
	}
	if _, err := s.SetPreferences(ctx, adminUser, "sarah", PreferencesRequest{DateFormat: "mdy"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("even admins don't set someone else's preferences: %v", err)
	}
	var verr *ValidationError
	if _, err := s.SetPreferences(ctx, sarah, "sarah", PreferencesRequest{DateFormat: "dd.mm"}); !errors.As(err, &verr) {
		t.Errorf("unknown format: %v", err)
	}
	if repo.users[1].DateFormat != "ymd" {
		t.Error("a rejected change must not be saved")
	}
}
