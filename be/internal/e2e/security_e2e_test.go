//go:build e2e

package e2e

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

// TestNoPrivilegeEscalation: nobody raises their own role or flags, edits a
// colleague's account, creates admins or deletes people; coordinators add
// people only to their own clinic and only with clinic roles.
func TestNoPrivilegeEscalation(t *testing.T) {
	rec, _ := login(t, "de.rita.lopes")
	coord, _ := login(t, "de.marta.ribeiro")
	admin, _ := login(t, "admin.user")

	for name, body := range map[string]map[string]any{
		"own role":      {"user_type": "StateAdmin"},
		"own superuser": {"is_superuser": true},
		"own clinic":    {"home_facility": "6b1f3c1e-5a8e-4c1d-9a0e-000000000002"},
		"own status":    {"employment_status": "active", "is_active": true},
		"own password":  {"password": "hunter22hunter22"},
		"own username":  {"username": "boss"},
	} {
		expect(t, name, rec.call("PATCH", "/users/de.rita.lopes/", body, nil), 403)
		expect(t, name+" (PUT)", rec.call("PUT", "/users/de.rita.lopes/", body, nil), 403)
	}
	var me map[string]any
	expect(t, "own profile field", rec.call("PATCH", "/users/de.rita.lopes/", map[string]any{"qualification": "Receção"}, &me), 200)
	if me["user_type"] != "Receptionist" || me["is_superuser"] == true {
		t.Fatalf("role changed: %v", me["user_type"])
	}
	expect(t, "colleague's account", rec.call("PATCH", "/users/de.marta.ribeiro/", map[string]any{"first_name": "X"}, nil), 403)
	expect(t, "coordinator makes an admin", coord.call("PATCH", "/users/de.rita.lopes/", map[string]any{"user_type": "StateAdmin"}, nil), 403)

	stamp := time.Now().Format("150405.000")
	newUser := func(role string) map[string]any {
		return map[string]any{
			"username": "e2e." + stamp + "." + role, "password": "Seguro#2026", "email": "e2e." + stamp + "." + role + "@example.test",
			"first_name": "E2E", "last_name": role, "user_type": role,
			"home_facility": "6b1f3c1e-5a8e-4c1d-9a0e-000000000002", // another clinic: must be ignored
		}
	}
	expect(t, "front desk adds people", rec.call("POST", "/users/", newUser("Doctor"), nil), 403)
	expect(t, "coordinator adds an admin", coord.call("POST", "/users/add_user/", newUser("StateAdmin"), nil), 403)
	short := newUser("Nurse")
	short["password"] = "123"
	expect(t, "weak password", coord.call("POST", "/users/", short, nil), 400)

	var created, coordMe map[string]any
	expect(t, "coordinator adds a physio", coord.call("POST", "/users/add_user/", newUser("PhysioTherapist"), &created), 201)
	coord.call("GET", "/users/getcurrentuser/", nil, &coordMe)
	if created["home_facility"] != coordMe["home_facility"] || created["is_superuser"] == true {
		t.Fatalf("new user must join the coordinator's clinic: %v vs %v", created["home_facility"], coordMe["home_facility"])
	}
	username := created["username"].(string)
	if !regexp.MustCompile(`^de\.e2e\.physiotherapist\d*$`).MatchString(username) {
		t.Errorf("username must follow <clinic code>.<first>.<last>: %q", username)
	}
	var sug map[string]string
	expect(t, "suggestion", coord.call("GET", "/users/username_suggestion/?first_name=Jo%C3%A3o&last_name=Silva+Teixeira", nil, &sug), 200)
	if sug["username"] != "de.joao.teixeira" {
		t.Errorf("suggestion: %v", sug)
	}
	expect(t, "front desk gets no suggestions", rec.call("GET", "/users/username_suggestion/?first_name=A&last_name=B", nil, nil), 403)
	anon := &client{t: t, base: baseURL(t)}
	var tokens map[string]any
	expect(t, "new user signs in", anon.call("POST", "/auth/login/", map[string]string{"username": username, "password": "Seguro#2026"}, &tokens), 200)
	// First sign-in: only "who am I" and the password change work
	newbie := &client{t: t, base: baseURL(t), token: tokens["access"].(string)}
	var who map[string]any
	expect(t, "who am I", newbie.call("GET", "/users/getcurrentuser/", nil, &who), 200)
	if who["must_change_password"] != true {
		t.Errorf("a new account must change its password: %v", who["must_change_password"])
	}
	expect(t, "blocked until the password changes", newbie.call("GET", "/patient/", nil, nil), 403)
	change := map[string]any{"username": username, "old_password": "Seguro#2026", "new_password": "Seguro#2026"}
	expect(t, "same password refused", newbie.call("PUT", "/password_change/", change, nil), 400)
	change["new_password"] = "Proprio#2026"
	expect(t, "own password", newbie.call("PUT", "/password_change/", change, nil), 200)
	expect(t, "now the app opens", newbie.call("GET", "/patient/", nil, nil), 200)
	expect(t, "coordinator deletes", coord.call("DELETE", "/users/"+username+"/", nil, nil), 403)
	expect(t, "admin deletes", admin.call("DELETE", "/users/"+username+"/", nil, nil), 200)
	expect(t, "admin cannot delete themself", admin.call("DELETE", "/users/admin.user/", nil, nil), 404)
}

// TestClinicIsolation: people, clinics, skills and time logs of another
// clinic are invisible (404, so existence isn't revealed); lists are capped.
func TestClinicIsolation(t *testing.T) {
	physio, _ := login(t, "de.sarah.johnson")
	coord, _ := login(t, "de.marta.ribeiro")
	cm, _ := login(t, "cm.ana.ferreira")
	const climutua = "6b1f3c1e-5a8e-4c1d-9a0e-000000000002"

	expect(t, "own profile", physio.call("GET", "/users/de.sarah.johnson/", nil, nil), 200)
	expect(t, "colleague", physio.call("GET", "/users/de.marta.ribeiro/", nil, nil), 200)
	expect(t, "other clinic's person", physio.call("GET", "/users/cm.catarina.lopes/", nil, nil), 404)
	expect(t, "other clinic's skills", coord.call("GET", "/users/cm.catarina.lopes/skill/", nil, nil), 404)
	expect(t, "other clinic's clinic link", coord.call("GET", "/users/cm.catarina.lopes/get_facilities/", nil, nil), 404)
	expect(t, "other clinic's staff", coord.call("GET", "/facility/"+climutua+"/get_users/", nil, nil), 404)
	expect(t, "other clinic", coord.call("GET", "/facility/"+climutua+"/", nil, nil), 404)

	var clinics page
	expect(t, "clinic list", physio.call("GET", "/facility/", nil, &clinics), 200)
	if clinics.Count != 1 {
		t.Errorf("staff see only their own clinic, got %d", clinics.Count)
	}
	expect(t, "coordinator creates a clinic", coord.call("POST", "/facility/", map[string]any{"name": "X", "facility_type": "Clinic", "address": "Y"}, nil), 403)
	expect(t, "coordinator deletes another clinic", coord.call("DELETE", "/facility/"+climutua+"/", nil, nil), 403)
	expect(t, "coordinator edits another clinic", coord.call("PATCH", "/facility/"+climutua+"/", map[string]any{"name": "Hacked"}, nil), 403)

	var cmMe map[string]any
	cm.call("GET", "/users/getcurrentuser/", nil, &cmMe)
	expect(t, "other clinic's time log", coord.call("GET", fmt.Sprintf("/timelog/user/%v", cmMe["id"]), nil, nil), 404)
	expect(t, "own time log", cm.call("GET", fmt.Sprintf("/timelog/user/%v", cmMe["id"]), nil, nil), 200)

	var users page
	expect(t, "huge page", cm.call("GET", "/users/?limit=100000", nil, &users), 200)
	if len(users.Results) > 100 {
		t.Errorf("list not capped: %d", len(users.Results))
	}
}

// TestDatePreference: people choose their own date format; nobody else can.
func TestDatePreference(t *testing.T) {
	physio, _ := login(t, "de.sarah.johnson")
	coord, _ := login(t, "de.marta.ribeiro")
	var me map[string]any
	expect(t, "own preference", physio.call("PATCH", "/users/de.sarah.johnson/preferences/", map[string]any{"date_format": "ymd"}, &me), 200)
	if me["date_format"] != "ymd" {
		t.Fatalf("saved: %v", me["date_format"])
	}
	physio.call("GET", "/users/getcurrentuser/", nil, &me)
	if me["date_format"] != "ymd" {
		t.Errorf("current user carries it: %v", me["date_format"])
	}
	expect(t, "someone else's", coord.call("PATCH", "/users/de.sarah.johnson/preferences/", map[string]any{"date_format": "mdy"}, nil), 403)
	expect(t, "unknown format", physio.call("PATCH", "/users/de.sarah.johnson/preferences/", map[string]any{"date_format": "dd.mm"}, nil), 400)
	expect(t, "back to default", physio.call("PATCH", "/users/de.sarah.johnson/preferences/", map[string]any{"date_format": "dmy"}, nil), 200)
}

// TestClinicCodes: a new clinic needs a free 2-4 letter code; once set it
// cannot change (usernames use it).
func TestClinicCodes(t *testing.T) {
	admin, _ := login(t, "admin.user")
	coord, _ := login(t, "cm.ana.ferreira")
	clinic := func(code string) map[string]any {
		return map[string]any{"name": "Clínica E2E " + code, "facility_type": 1, "address": "Rua X", "code": code}
	}
	expect(t, "no code", admin.call("POST", "/facility/", clinic(""), nil), 400)
	expect(t, "bad code", admin.call("POST", "/facility/", clinic("C1"), nil), 400)
	expect(t, "taken code", admin.call("POST", "/facility/", clinic("cm"), nil), 400)

	var me map[string]any
	coord.call("GET", "/users/getcurrentuser/", nil, &me)
	path := "/facility/" + me["home_facility"].(string) + "/"
	expect(t, "code cannot change", admin.call("PATCH", path, map[string]any{"code": "xy"}, nil), 400)
	expect(t, "same code is fine", coord.call("PATCH", path, map[string]any{"code": "cm", "address": "Rua de Cedofeita 312, 4050-180 Porto"}, nil), 200)
}
