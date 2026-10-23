//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"klyniva/internal/clinic"
	"klyniva/internal/database"
	"klyniva/internal/models"
	"klyniva/internal/modules/patient"
	"klyniva/internal/modules/referral"
)

// TestMigrationKeepsColumnTypes guards against AutoMigrate rewriting tables
// owned by other modules through relation structs (it once turned
// patients.id into text via referral.PatientSummary).
func TestMigrationKeepsColumnTypes(t *testing.T) {
	startDB(t)
	want := map[[2]string]string{
		{"patients", "id"}:          "uuid",
		{"referrals", "patient_id"}: "uuid",
		{"users", "id"}:             "bigint",
		{"facilities", "id"}:        "uuid",
	}
	for key, typ := range want {
		var got string
		err := database.GetDB().Raw(
			"SELECT data_type FROM information_schema.columns WHERE table_name = ? AND column_name = ?",
			key[0], key[1]).Scan(&got).Error
		if err != nil || got != typ {
			t.Errorf("%s.%s: want %s, got %q (err %v)", key[0], key[1], typ, got, err)
		}
	}
}

// newUser inserts an active user, optionally working in a clinic.
func newUser(t *testing.T, username string, role models.UserRole, clinicID *string) referral.Actor {
	t.Helper()
	u := models.User{Username: username, Email: username + "@example.com", FirstName: "Test",
		LastName: username, Password: "x", UserType: role, IsActive: true, HomeFacilityID: clinicID}
	if err := database.GetDB().Create(&u).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return referral.Actor{ID: u.ID, Role: role}
}

// TestPatientVisibility checks the SQL scoping: staff see the patients of
// their own clinic, never another clinic's, and a referral grants access to a
// therapist who has no clinic assigned.
func TestPatientVisibility(t *testing.T) {
	startDB(t)
	ctx := context.Background()
	db := database.GetDB()
	svc := patient.NewService(patient.NewGormRepository(db), clinic.NewGormResolver(db), nil)

	doctor := userID(t, "de.john.smith")
	nurse := userID(t, "de.mary.williams")
	asPatientActor := func(a referral.Actor) patient.Actor { return patient.Actor{ID: a.ID, Role: a.Role} }

	otherClinic := models.Facility{Name: "Other Clinic"}
	if err := db.Create(&otherClinic).Error; err != nil {
		t.Fatal(err)
	}
	farNurse := newUser(t, "far-nurse", models.RoleNurse, &otherClinic.ID)
	freelancePhysio := newUser(t, "freelance-physio", models.RolePhysioTherapist, nil)

	p, err := svc.Create(ctx, asPatientActor(doctor), patient.CreateRequest{Name: "Rui Almeida"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.FacilityID == nil || *p.FacilityID != database.DemoClinicID {
		t.Fatalf("patient should belong to the doctor's clinic, got %v", p.FacilityID)
	}

	// Search matches case-insensitively; wildcards in input are literal.
	items, total, err := svc.List(ctx, asPatientActor(doctor), patient.ListQuery{Search: "rui"})
	if err != nil || total != 1 || items[0].ID != p.ID {
		t.Fatalf("doctor search: total=%d err=%v", total, err)
	}
	if _, total, _ := svc.List(ctx, asPatientActor(doctor), patient.ListQuery{Search: "%"}); total != 0 {
		t.Errorf("literal %% should not match everything, got %d", total)
	}

	// Same clinic: visible. Another clinic: invisible.
	if _, err := svc.Get(ctx, asPatientActor(nurse), p.ID); err != nil {
		t.Errorf("nurse of the same clinic should see patient, got %v", err)
	}
	if _, err := svc.Get(ctx, asPatientActor(farNurse), p.ID); !errors.Is(err, patient.ErrNotFound) {
		t.Errorf("nurse of another clinic must not see patient, got %v", err)
	}
	if _, total, _ := svc.List(ctx, asPatientActor(farNurse), patient.ListQuery{}); total != 0 {
		t.Errorf("other clinic list should be empty, got %d", total)
	}

	// A therapist without a clinic sees the patient only once a referral links them.
	if _, err := svc.Get(ctx, asPatientActor(freelancePhysio), p.ID); !errors.Is(err, patient.ErrNotFound) {
		t.Errorf("unlinked physio should not see patient, got %v", err)
	}
	refSvc := referral.NewService(referral.NewGormRepository(db), clinic.NewGormResolver(db), nil)
	if _, err := refSvc.Create(ctx, doctor, referral.CreateRequest{
		PatientID: p.ID, TherapistID: &freelancePhysio.ID, Diagnosis: "Ankle sprain", Reason: "Proprioception",
	}); err != nil {
		t.Fatalf("referral: %v", err)
	}
	if _, err := svc.Get(ctx, asPatientActor(freelancePhysio), p.ID); err != nil {
		t.Errorf("physio should see referred patient, got %v", err)
	}

	// Admins see everything.
	if _, err := svc.Get(ctx, patient.Actor{ID: 999, Role: models.RoleStateAdmin}, p.ID); err != nil {
		t.Errorf("admin get: %v", err)
	}

	// Partial update round-trips through PostgreSQL.
	upd, err := svc.Update(ctx, asPatientActor(doctor), p.ID, map[string]any{
		"date_of_birth": "1975-03-14", "gender": float64(1), "phone_number": "+351 910 000 000",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.DateOfBirth == nil || upd.DateOfBirth.Year() != 1975 || upd.Gender == nil || *upd.Gender != 1 {
		t.Errorf("update not persisted: %+v", upd)
	}
}

// TestPatientDuplicatesOnPostgres checks the SQL pre-filter (birth date,
// phone by last 9 digits, identifiers), clinic scoping of duplicates and the
// partial unique indexes that back the NIF/SNS rule.
func TestPatientDuplicatesOnPostgres(t *testing.T) {
	startDB(t)
	ctx := context.Background()
	db := database.GetDB()
	svc := patient.NewService(patient.NewGormRepository(db), clinic.NewGormResolver(db), nil)
	nurse := userID(t, "de.mary.williams")
	as := patient.Actor{ID: nurse.ID, Role: nurse.Role}
	s := func(v string) *string { return &v }

	first, err := svc.Create(ctx, as, patient.CreateRequest{
		Name: "Inês Conceição Ramos", DateOfBirth: s("1984-02-29"), PhoneNumber: s("+351 913 111 222"), NIF: s("123456789"),
	})
	if err != nil {
		t.Fatal(err)
	}

	for name, req := range map[string]patient.CreateRequest{
		"birth date": {Name: "ines ramos", DateOfBirth: s("1984-02-29")},
		"phone":      {Name: "INES RAMOS", PhoneNumber: s("913111222")},
		"NIF":        {Name: "Someone Else", NIF: s("123456789")},
	} {
		var dup *patient.DuplicateError
		if _, err := svc.Create(ctx, as, req); !errors.As(err, &dup) || len(dup.Matches) != 1 || dup.Matches[0].ID != first.ID {
			t.Errorf("%s: want one duplicate, got %v", name, err)
		}
	}

	// Another clinic neither sees the patient as a duplicate nor clashes on NIF
	other := models.Facility{Name: "Clinic B"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	far := newUser(t, "dup-far-nurse", models.RoleNurse, &other.ID)
	if _, err := svc.Create(ctx, patient.Actor{ID: far.ID, Role: far.Role},
		patient.CreateRequest{Name: "Inês Ramos", DateOfBirth: s("1984-02-29"), NIF: s("123456789")}); err != nil {
		t.Errorf("other clinic: %v", err)
	}

	// The unique index stops a second NIF in the same clinic even without the service check
	dupRow := models.Patient{Name: "Raw insert", NIF: s("123456789"), FacilityID: first.FacilityID, IsActive: true}
	if err := patient.NewGormRepository(db).Create(ctx, &dupRow); !errors.Is(err, patient.ErrDuplicateIdentifier) {
		t.Errorf("unique index: want ErrDuplicateIdentifier, got %v", err)
	}
}
