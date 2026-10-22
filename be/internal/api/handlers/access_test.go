package handlers

import (
	"errors"
	"testing"

	"klyniva/internal/models"
)

var (
	clinicA, clinicB = "clinic-a", "clinic-b"
	admin            = &actor{ID: 1, Role: models.RoleStateAdmin}
	coordA           = &actor{ID: 2, Role: models.RoleCoordinator, Clinic: &clinicA}
	physioA          = &actor{ID: 3, Role: models.RolePhysioTherapist, Clinic: &clinicA}
	recepA           = &actor{ID: 4, Role: models.RoleReceptionist, Clinic: &clinicA}
	coordB           = &actor{ID: 5, Role: models.RoleCoordinator, Clinic: &clinicB}
	noClinic         = &actor{ID: 6, Role: models.RoleDoctor}
)

func userIn(id uint, role models.UserRole, clinic *string) *models.User {
	return &models.User{ID: id, UserType: role, HomeFacilityID: clinic}
}

func TestCanSeeUser(t *testing.T) {
	sarah := userIn(3, models.RolePhysioTherapist, &clinicA)
	cases := []struct {
		who  string
		a    *actor
		want bool
	}{
		{"admin", admin, true}, {"self", physioA, true}, {"same clinic", recepA, true},
		{"other clinic", coordB, false}, {"no clinic", noClinic, false},
	}
	for _, c := range cases {
		if got := canSeeUser(c.a, sarah); got != c.want {
			t.Errorf("%s: %v", c.who, got)
		}
	}
}

// The privilege-escalation cases: nobody raises their own role or flags,
// staff can't edit colleagues, coordinators can't hand out admin roles.
func TestUserUpdate(t *testing.T) {
	self := userIn(4, models.RoleReceptionist, &clinicA)
	cases := []struct {
		name string
		a    *actor
		body map[string]any
		want error
	}{
		{"own name", recepA, map[string]any{"first_name": "Sofia"}, nil},
		{"own role up", recepA, map[string]any{"user_type": "StateAdmin"}, errForbidden},
		{"own superuser", recepA, map[string]any{"is_superuser": true}, errForbidden},
		{"own active flag", recepA, map[string]any{"is_active": true}, errForbidden},
		{"own clinic", recepA, map[string]any{"home_facility": clinicB}, errForbidden},
		{"own password", recepA, map[string]any{"password": "x"}, errForbidden},
		{"colleague edits", physioA, map[string]any{"first_name": "X"}, errForbidden},
		{"other clinic coordinator", coordB, map[string]any{"first_name": "X"}, errForbidden},
		{"coordinator gives clinic role", coordA, map[string]any{"user_type": "Coordinator"}, nil},
		{"coordinator gives admin role", coordA, map[string]any{"user_type": "StateAdmin"}, errForbidden},
		{"admin gives admin role", admin, map[string]any{"user_type": "DistrictAdmin"}, nil},
		{"admin invents a role", admin, map[string]any{"user_type": "God"}, errForbidden},
		{"same role is no change", recepA, map[string]any{"user_type": "Receptionist", "last_name": "M"}, nil},
		{"bad email", recepA, map[string]any{"email": "nope"}, errBadField},
	}
	for _, c := range cases {
		out, err := userUpdate(c.a, self, c.body)
		if !errors.Is(err, c.want) {
			t.Errorf("%s: err %v, want %v", c.name, err, c.want)
		}
		if err == nil {
			if _, ok := out["is_superuser"]; ok {
				t.Errorf("%s: superuser leaked", c.name)
			}
		}
	}
	out, _ := userUpdate(recepA, self, map[string]any{"first_name": "S", "state": 1, "read_profile_picture_url": "x"})
	if out["first_name"] != "S" || out["state_id"] != 1 || len(out) != 2 {
		t.Errorf("columns: %v", out)
	}
	if out, _ := userUpdate(recepA, self, map[string]any{"user_type": "Receptionist"}); len(out) != 0 {
		t.Errorf("unchanged role must not be written: %v", out)
	}
}

func TestNewUserClinic(t *testing.T) {
	req := func(role models.UserRole, clinic *string) *newUserRequest {
		return &newUserRequest{UserType: role, HomeFacility: clinic}
	}
	if c, err := newUserClinic(coordA, req(models.RolePhysioTherapist, &clinicB)); err != nil || *c != clinicA {
		t.Errorf("coordinators add to their own clinic: %v %v", c, err)
	}
	if _, err := newUserClinic(coordA, req(models.RoleStateAdmin, nil)); !errors.Is(err, errForbidden) {
		t.Errorf("coordinator creates admin: %v", err)
	}
	for _, a := range []*actor{physioA, recepA, noClinic} {
		if _, err := newUserClinic(a, req(models.RoleDoctor, &clinicA)); !errors.Is(err, errForbidden) {
			t.Errorf("role %s creates users: %v", a.Role, err)
		}
	}
	if c, err := newUserClinic(admin, req(models.RoleDoctor, &clinicB)); err != nil || *c != clinicB {
		t.Errorf("admin chooses the clinic: %v %v", c, err)
	}
}

func TestClinicRules(t *testing.T) {
	if !canManageClinic(coordA, clinicA) || canManageClinic(coordA, clinicB) || canManageClinic(physioA, clinicA) {
		t.Error("coordinators manage only their own clinic; staff none")
	}
	if !canSeeClinic(physioA, clinicA) || canSeeClinic(physioA, clinicB) || canSeeClinic(noClinic, clinicA) || !canSeeClinic(admin, clinicB) {
		t.Error("clinic visibility")
	}
	out := clinicUpdate(map[string]any{"name": "N", "id": "x", "deleted_at": nil, "features": []any{1}})
	if len(out) != 1 || out["name"] != "N" {
		t.Errorf("clinic columns: %v", out)
	}
}

func TestPageLimit(t *testing.T) {
	for in, want := range map[int]int{0: 20, -5: 20, 50: 50, 100: 100, 100000: 20} {
		if got := pageLimit(in); got != want {
			t.Errorf("pageLimit(%d) = %d", in, got)
		}
	}
}
