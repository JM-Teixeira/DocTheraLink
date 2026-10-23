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
	nurse := userID(t, "de.mary.williams")
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
	coord := userID(t, "de.marta.ribeiro")
	coordinator := schedule.Actor{ID: coord.ID, Role: coord.Role}
	sarah := userID(t, "de.sarah.johnson")

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

	// Attending a session uses the per-session consumables (real inventory)
	stock := inventory.NewService(inventory.NewGormRepository(db), clinic.NewGormResolver(db))
	invActor := inventory.Actor{ID: coord.ID, Role: coord.Role}
	gloves, err := stock.Create(ctx, invActor, inventory.ItemRequest{Name: "Board gloves", Quantity: 10, PerSession: 2})
	if err != nil {
		t.Fatal(err)
	}
	withHook := schedule.NewService(schedule.NewGormRepository(db), clinic.NewGormResolver(db)).WithSessionHook(stock)
	var apptID string
	for _, row := range b.Therapists {
		if row.ID == sarah.ID {
			apptID = row.Hours[1].Appointments[0].ID
		}
	}
	if _, err := withHook.SetStatus(ctx, coordinator, apptID, schedule.StatusRequest{Status: schedule.StatusAttended}); err != nil {
		t.Fatal(err)
	}
	items, _ := stock.List(ctx, invActor, inventory.ListFilter{Search: "Board gloves"})
	if len(items) != 1 || items[0].Quantity != 8 {
		t.Fatalf("2 gloves used per session: %+v", items)
	}
	auto := true
	used, err := stock.Audit(ctx, invActor, inventory.AuditFilter{ItemID: gloves.ID, Auto: &auto})
	if err != nil || len(used) != 1 || used[0].Item == nil || used[0].Item.Name != "Board gloves" || used[0].User == nil {
		t.Fatalf("audit of automatic use: %+v %v", used, err)
	}
	if _, err := withHook.SetStatus(ctx, coordinator, apptID, schedule.StatusRequest{Status: schedule.StatusMissed}); err != nil {
		t.Fatal(err)
	}
	if items, _ = stock.List(ctx, invActor, inventory.ListFilter{Search: "Board gloves"}); items[0].Quantity != 10 {
		t.Fatalf("undo gives back: %d", items[0].Quantity)
	}

	// Credits: a level-5 patient leaves 3 of 8 credits in a gym hour
	gym := divs[0]
	if gym.CreditLimit != schedule.DefaultCreditLimit {
		t.Fatalf("gym credit limit: %+v", gym)
	}
	tiago := userID(t, "de.tiago.martins")
	if _, err := svc.SaveShift(ctx, coordinator, "", schedule.ShiftRequest{
		Date: date, StartHour: 14, EndHour: 16, TherapistID: tiago.ID, DivisionID: gym.ID,
	}); err != nil {
		t.Fatal(err)
	}
	five := 5
	var neuro []string
	for i := 0; i < 2; i++ {
		p, err := psvc.Create(ctx, patient.Actor{ID: coord.ID, Role: coord.Role},
			patient.CreateRequest{Name: "Neuro Patient " + string(rune('A'+i)), CareLevel: &five, AllowDuplicate: true})
		if err != nil {
			t.Fatal(err)
		}
		neuro = append(neuro, p.ID)
	}
	if _, err := svc.Book(ctx, coordinator, schedule.AppointmentRequest{Date: date, Hour: 14, TherapistID: tiago.ID, PatientID: neuro[0]}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Book(ctx, coordinator, schedule.AppointmentRequest{Date: date, Hour: 14, TherapistID: tiago.ID, PatientID: neuro[1]}); !errors.Is(err, schedule.ErrOverCredits) {
		t.Fatalf("second level-5 patient: %v", err)
	}

	// A weekly plan books on PostgreSQL; the year view aggregates
	plan, err := svc.Series(ctx, coordinator, schedule.SeriesRequest{PatientID: neuro[1], TherapistID: tiago.ID,
		Weekdays: []int{int(time.Sunday)}, Hour: 15, StartDate: date, Sessions: 1})
	if err != nil || len(plan.Booked) != 1 || plan.SeriesID == "" {
		t.Fatalf("series: %+v %v", plan, err)
	}
	year, err := svc.Year(ctx, coordinator, 2031)
	if err != nil || year.Months[5].Booked < 2 || year.Months[5].Patients < 2 {
		t.Fatalf("June 2031 stats: %+v %v", year.Months[5], err)
	}
	if _, err := svc.Range(ctx, coordinator, "2031-05-26", "2031-06-01"); err != nil {
		t.Fatalf("range: %v", err)
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
