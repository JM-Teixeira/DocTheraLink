//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"care-backend/internal/database"
	"care-backend/internal/models"
	"care-backend/internal/modules/patient"
	"care-backend/internal/modules/referral"
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

// TestPatientVisibility checks the SQL scoping: a user sees patients they
// created, and a therapist gains access once a referral links them.
func TestPatientVisibility(t *testing.T) {
	startDB(t)
	ctx := context.Background()
	db := database.GetDB()
	svc := patient.NewService(patient.NewGormRepository(db), nil)

	doctor := userID(t, "doctor")
	physio := userID(t, "physiotherapist")
	nurse := userID(t, "nurse")
	asPatientActor := func(a referral.Actor) patient.Actor { return patient.Actor{ID: a.ID, Role: a.Role} }

	p, err := svc.Create(ctx, asPatientActor(doctor), patient.CreateRequest{Name: "Rui Almeida"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Search matches case-insensitively; wildcards in input are literal.
	items, total, err := svc.List(ctx, asPatientActor(doctor), patient.ListQuery{Search: "rui"})
	if err != nil || total != 1 || items[0].ID != p.ID {
		t.Fatalf("doctor search: total=%d err=%v", total, err)
	}
	if _, total, _ := svc.List(ctx, asPatientActor(doctor), patient.ListQuery{Search: "%"}); total != 0 {
		t.Errorf("literal %% should not match everything, got %d", total)
	}

	// Unrelated staff cannot see the patient.
	if _, err := svc.Get(ctx, asPatientActor(nurse), p.ID); !errors.Is(err, patient.ErrNotFound) {
		t.Errorf("nurse should not see patient, got %v", err)
	}
	if _, err := svc.Get(ctx, asPatientActor(physio), p.ID); !errors.Is(err, patient.ErrNotFound) {
		t.Errorf("physio should not see patient before referral, got %v", err)
	}

	// After the doctor refers the patient to the physio, the physio can see them.
	refSvc := referral.NewService(referral.NewGormRepository(db), nil)
	if _, err := refSvc.Create(ctx, doctor, referral.CreateRequest{
		PatientID: p.ID, TherapistID: &physio.ID, Diagnosis: "Ankle sprain", Reason: "Proprioception",
	}); err != nil {
		t.Fatalf("referral: %v", err)
	}
	if _, err := svc.Get(ctx, asPatientActor(physio), p.ID); err != nil {
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
