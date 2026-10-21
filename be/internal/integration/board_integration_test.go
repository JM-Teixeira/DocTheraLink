//go:build integration

package integration

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"klyniva/internal/clinic"
	"klyniva/internal/database"
	"klyniva/internal/modules/inventory"
	"klyniva/internal/modules/patient"
	"klyniva/internal/modules/schedule"
)

// TestInventoryOnPostgres: seeded stock, row-locked movements that never go
// below zero under concurrency, and history.
func TestInventoryOnPostgres(t *testing.T) {
	startDB(t)
	ctx := context.Background()
	db := database.GetDB()
	svc := inventory.NewService(inventory.NewGormRepository(db), clinic.NewGormResolver(db))
	nurse := userID(t, "nurse")
	as := inventory.Actor{ID: nurse.ID, Role: nurse.Role}

	seeded, err := svc.List(ctx, as, inventory.ListFilter{Status: "attention"})
	if err != nil || len(seeded) == 0 {
		t.Fatalf("demo stock should include low/out items: %d %v", len(seeded), err)
	}

	it, err := svc.Create(ctx, as, inventory.ItemRequest{Name: "Concurrency gauze", Quantity: 5})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var ok, short int
	var mu sync.Mutex
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Move(ctx, as, it.ID, inventory.MovementRequest{Kind: inventory.KindUse, Quantity: 1})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok++
			case errors.Is(err, inventory.ErrInsufficient):
				short++
			default:
				t.Errorf("use: %v", err)
			}
		}()
	}
	wg.Wait()
	if ok != 5 || short != 3 {
		t.Fatalf("8 uses of 5 in stock: %d ok, %d refused", ok, short)
	}
	hist, err := svc.History(ctx, as, it.ID)
	if err != nil || len(hist) != 6 || hist[0].QuantityAfter != 0 || hist[0].User == nil {
		t.Fatalf("history: %d entries, err %v", len(hist), err)
	}
}

// TestTimeBoardOnPostgres: seeded divisions and shifts, the board query,
// and the advisory lock that keeps an hour from being overbooked.
func TestTimeBoardOnPostgres(t *testing.T) {
	startDB(t)
	ctx := context.Background()
	db := database.GetDB()
	svc := schedule.NewService(schedule.NewGormRepository(db), clinic.NewGormResolver(db))
	coord := userID(t, "coordinator")
	coordinator := schedule.Actor{ID: coord.ID, Role: coord.Role}
	sarah := userID(t, "physiotherapist")

	// A Sunday far away: no seeded shifts
	date := "2031-06-01"
	divs, err := svc.Divisions(ctx, coordinator)
	if err != nil || len(divs) < 4 {
		t.Fatalf("seeded divisions: %d %v", len(divs), err)
	}
	room := divs[1] // Treatment room 1, capacity 1
	if room.Capacity != 1 {
		t.Fatalf("expected a capacity-1 room, got %+v", room)
	}
	if _, err := svc.SaveShift(ctx, coordinator, "", schedule.ShiftRequest{
		Date: date, StartHour: 9, EndHour: 12, TherapistID: sarah.ID, DivisionID: room.ID,
	}); err != nil {
		t.Fatal(err)
	}

	// Patients of the clinic
	psvc := patient.NewService(patient.NewGormRepository(db), clinic.NewGormResolver(db), nil)
	var ids []string
	for i := 0; i < 4; i++ {
		p, err := psvc.Create(ctx, patient.Actor{ID: coord.ID, Role: coord.Role},
			patient.CreateRequest{Name: "Board Patient " + string(rune('A'+i)), AllowDuplicate: true})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, p.ID)
	}

	// Four receptionists race for the single place at 10:00
	var wg sync.WaitGroup
	var mu sync.Mutex
	booked, full := 0, 0
	for _, pid := range ids {
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			_, err := svc.Book(ctx, coordinator, schedule.AppointmentRequest{Date: date, Hour: 10, TherapistID: sarah.ID, PatientID: pid})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				booked++
			case errors.Is(err, schedule.ErrFull):
				full++
			default:
				t.Errorf("book: %v", err)
			}
		}(pid)
	}
	wg.Wait()
	if booked != 1 || full != 3 {
		t.Fatalf("one place, four bookings: %d booked, %d full", booked, full)
	}

	b, err := svc.Board(ctx, coordinator, date)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, row := range b.Therapists {
		if row.ID != sarah.ID {
			continue
		}
		found = true
		if row.Patients != 1 || len(row.Hours) != 3 || row.Hours[1].Appointments[0].Patient == nil {
			t.Fatalf("sarah's row: %+v", row)
		}
	}
	if !found || !b.Permissions.ManageShifts {
		t.Fatalf("board: found=%v permissions=%+v", found, b.Permissions)
	}

	// Seeded weekday shifts exist for the coming two weeks
	next := time.Now()
	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
		next = next.AddDate(0, 0, 1)
	}
	b, err = svc.Board(ctx, coordinator, next.Format("2006-01-02"))
	if err != nil {
		t.Fatal(err)
	}
	shifts := 0
	for _, row := range b.Therapists {
		shifts += len(row.Shifts)
	}
	if shifts == 0 {
		t.Error("demo shifts should be seeded on weekdays")
	}
}
