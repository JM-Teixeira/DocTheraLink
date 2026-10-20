//go:build integration

// Package integration runs tests against a real PostgreSQL started with
// testcontainers. They need Docker and run with: go test -tags integration ./...
package integration

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"klyniva/internal/clinic"
	"klyniva/internal/config"
	"klyniva/internal/database"
	"klyniva/internal/models"
	"klyniva/internal/modules/referral"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// startDB boots PostgreSQL, runs every migration and seeds demo data once per test.
func startDB(t *testing.T) {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()

	pg, err := postgres.Run(ctx, "postgres:18-alpine",
		postgres.WithDatabase("care"), postgres.WithUsername("care"), postgres.WithPassword("care"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	host, _ := pg.Host(ctx)
	port, _ := pg.MappedPort(ctx, "5432/tcp")
	cfg := &config.Config{Server: config.ServerConfig{Debug: os.Getenv("SQL_DEBUG") != ""}, Database: config.DatabaseConfig{
		Host: host, Port: port.Port(), User: "care", Password: "care", Name: "care",
		SSLMode: "disable", Timezone: "UTC",
	}}
	if err := database.Connect(cfg); err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := database.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := database.SeedData(true); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

// userID looks up a seeded demo user.
func userID(t *testing.T, username string) referral.Actor {
	t.Helper()
	var u models.User
	if err := database.GetDB().Where("username = ?", username).First(&u).Error; err != nil {
		t.Fatalf("user %s: %v", username, err)
	}
	return referral.Actor{ID: u.ID, Role: u.UserType}
}

func TestReferralLifecycleOnPostgres(t *testing.T) {
	startDB(t)
	ctx := context.Background()
	db := database.GetDB()

	patient := models.Patient{Name: "Maria Silva"}
	if err := db.Create(&patient).Error; err != nil {
		t.Fatalf("create patient: %v", err)
	}

	doctor := userID(t, "doctor")
	physio := userID(t, "physiotherapist")
	svc := referral.NewService(referral.NewGormRepository(db), clinic.NewGormResolver(db), nil)

	r, err := svc.Create(ctx, doctor, referral.CreateRequest{
		PatientID: patient.ID, Diagnosis: "Lumbar disc herniation", Reason: "Core strengthening",
		Priority: referral.PriorityUrgent,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.Patient == nil || r.Patient.Name != "Maria Silva" || r.Doctor == nil {
		t.Fatalf("relations not preloaded: %+v", r)
	}

	// The therapist inbox includes the unassigned referral.
	if _, total, err := svc.List(ctx, physio, referral.ListFilter{}); err != nil || total != 1 {
		t.Fatalf("inbox: total=%d err=%v", total, err)
	}

	for _, step := range []func() (*referral.Referral, error){
		func() (*referral.Referral, error) { return svc.Accept(ctx, physio, r.ID) },
		func() (*referral.Referral, error) { return svc.Start(ctx, physio, r.ID) },
		func() (*referral.Referral, error) { return svc.Complete(ctx, physio, r.ID, "Pain-free at 6 weeks") },
	} {
		if r, err = step(); err != nil {
			t.Fatalf("transition: %v", err)
		}
	}
	if r.Status != referral.StatusCompleted || r.Therapist == nil {
		t.Fatalf("final state: %+v", r)
	}

	notes, err := svc.Notes(ctx, doctor, r.ID)
	if err != nil || len(notes) != 3 || notes[0].Author == nil {
		t.Fatalf("system notes: len=%d err=%v", len(notes), err)
	}
}

// TestConcurrentAcceptHasOneWinner checks the compare-and-swap in Transition:
// when two therapists accept the same open referral, exactly one succeeds.
func TestConcurrentAcceptHasOneWinner(t *testing.T) {
	startDB(t)
	ctx := context.Background()
	db := database.GetDB()

	patient := models.Patient{Name: "João Costa"}
	db.Create(&patient)
	second := models.User{Username: "physio2", Email: "physio2@example.com", FirstName: "P",
		LastName: "Two", Password: "x", UserType: models.RolePhysioTherapist, IsActive: true}
	if err := db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}

	svc := referral.NewService(referral.NewGormRepository(db), clinic.NewGormResolver(db), nil)
	r, err := svc.Create(ctx, userID(t, "doctor"), referral.CreateRequest{
		PatientID: patient.ID, Diagnosis: "Frozen shoulder", Reason: "Mobilisation",
	})
	if err != nil {
		t.Fatal(err)
	}

	actors := []referral.Actor{userID(t, "physiotherapist"), {ID: second.ID, Role: models.RolePhysioTherapist}}
	errs := make([]error, len(actors))
	var wg sync.WaitGroup
	for i, a := range actors {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = svc.Accept(ctx, a, r.ID)
		}()
	}
	wg.Wait()

	ok := 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, referral.ErrConflict), errors.Is(err, referral.ErrInvalidTransition),
			errors.Is(err, referral.ErrNotFound):
			// expected for the loser
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if ok != 1 {
		t.Fatalf("want exactly one successful accept, got %d", ok)
	}
}
