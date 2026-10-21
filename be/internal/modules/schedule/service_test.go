package schedule

import (
	"context"
	"errors"
	"testing"
	"time"

	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory Repository with the same booking rules.
type fakeRepo struct {
	clinic     string
	divisions  map[string]Division
	therapists []UserSummary
	patients   map[string]bool
	shifts     map[string]Shift
	appts      map[string]Appointment
}

func (f *fakeRepo) DefaultClinic(context.Context) (*string, error) { return &f.clinic, nil }
func (f *fakeRepo) Divisions(_ context.Context, cid string) ([]Division, error) {
	var out []Division
	for _, d := range f.divisions {
		if d.ClinicID == cid && d.Active {
			out = append(out, d)
		}
	}
	return out, nil
}
func (f *fakeRepo) GetDivision(_ context.Context, cid, id string) (*Division, error) {
	d, ok := f.divisions[id]
	if !ok || d.ClinicID != cid {
		return nil, ErrNotFound
	}
	return &d, nil
}
func (f *fakeRepo) SaveDivision(_ context.Context, d *Division) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	f.divisions[d.ID] = *d
	return nil
}
func (f *fakeRepo) Therapists(context.Context, string) ([]UserSummary, error) {
	return f.therapists, nil
}
func (f *fakeRepo) IsTherapist(_ context.Context, cid string, id uint) (bool, error) {
	for _, t := range f.therapists {
		if t.ID == id {
			return cid == f.clinic, nil
		}
	}
	return false, nil
}
func (f *fakeRepo) PatientInClinic(_ context.Context, cid, id string) (bool, error) {
	return cid == f.clinic && f.patients[id], nil
}
func (f *fakeRepo) Shifts(_ context.Context, cid string, date time.Time) ([]Shift, error) {
	var out []Shift
	for _, s := range f.shifts {
		if s.ClinicID == cid && s.Date.Equal(date) {
			out = append(out, s)
		}
	}
	return out, nil
}
func (f *fakeRepo) GetShift(_ context.Context, cid, id string) (*Shift, error) {
	s, ok := f.shifts[id]
	if !ok || s.ClinicID != cid {
		return nil, ErrNotFound
	}
	return &s, nil
}
func (f *fakeRepo) SaveShift(_ context.Context, s *Shift) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	f.shifts[s.ID] = *s
	return nil
}
func (f *fakeRepo) DeleteShift(_ context.Context, id string) error { delete(f.shifts, id); return nil }
func (f *fakeRepo) Appointments(_ context.Context, cid string, date time.Time) ([]Appointment, error) {
	var out []Appointment
	for _, a := range f.appts {
		if a.ClinicID == cid && a.Date.Equal(date) {
			out = append(out, a)
		}
	}
	return out, nil
}
func (f *fakeRepo) GetAppointment(_ context.Context, cid, id string) (*Appointment, error) {
	a, ok := f.appts[id]
	if !ok || a.ClinicID != cid {
		return nil, ErrNotFound
	}
	return &a, nil
}
func (f *fakeRepo) Book(_ context.Context, a *Appointment, capacity int) error {
	booked := 0
	for _, o := range f.appts {
		if o.Date.Equal(a.Date) && o.Hour == a.Hour {
			if o.TherapistID == a.TherapistID {
				booked++
			}
			if o.PatientID == a.PatientID {
				return ErrPatientBusy
			}
		}
	}
	if booked >= capacity {
		return ErrFull
	}
	a.ID = uuid.NewString()
	f.appts[a.ID] = *a
	return nil
}
func (f *fakeRepo) UpdateAppointment(_ context.Context, id string, fields map[string]any) error {
	a := f.appts[id]
	a.Status = fields["status"].(string)
	f.appts[id] = a
	return nil
}
func (f *fakeRepo) DeleteAppointment(_ context.Context, id string) error {
	delete(f.appts, id)
	return nil
}

var (
	sarah       = UserSummary{ID: 10, FirstName: "Sarah", LastName: "Johnson"}
	tiago       = UserSummary{ID: 11, FirstName: "Tiago", LastName: "Martins"}
	coordinator = Actor{ID: 1, Role: models.RoleCoordinator}
	reception   = Actor{ID: 2, Role: models.RoleReceptionist}
	doctor      = Actor{ID: 3, Role: models.RoleDoctor}
	sarahActor  = Actor{ID: sarah.ID, Role: models.RolePhysioTherapist}
	admin       = Actor{ID: 9, Role: models.RoleStateAdmin} // no clinic
	monday      = "2026-10-26"
)

type fixture struct {
	svc        *Service
	repo       *fakeRepo
	gym, room  string // division ids (capacity 2 and 1)
	ana, bruno string // patients
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	cid := uuid.NewString()
	repo := &fakeRepo{clinic: cid, divisions: map[string]Division{}, therapists: []UserSummary{sarah, tiago},
		patients: map[string]bool{}, shifts: map[string]Shift{}, appts: map[string]Appointment{}}
	clinics := clinic.Static{coordinator.ID: cid, reception.ID: cid, doctor.ID: cid, sarah.ID: cid}
	f := fixture{svc: NewService(repo, clinics), repo: repo, ana: uuid.NewString(), bruno: uuid.NewString()}
	repo.patients[f.ana], repo.patients[f.bruno] = true, true
	ctx := context.Background()
	gym, err := f.svc.SaveDivision(ctx, coordinator, "", DivisionRequest{Name: "Gym", Capacity: 2})
	if err != nil {
		t.Fatal(err)
	}
	room, _ := f.svc.SaveDivision(ctx, coordinator, "", DivisionRequest{Name: "Room 1", Color: "indigo", Capacity: 1})
	f.gym, f.room = gym.ID, room.ID
	return f
}

func (f fixture) shift(t *testing.T, who uint, div string, from, to int) *Shift {
	t.Helper()
	s, err := f.svc.SaveShift(context.Background(), coordinator, "", ShiftRequest{
		Date: monday, StartHour: from, EndHour: to, TherapistID: who, DivisionID: div,
	})
	if err != nil {
		t.Fatalf("shift %d-%d: %v", from, to, err)
	}
	return s
}

func TestShiftRules(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.shift(t, sarah.ID, f.gym, 8, 13)

	cases := []struct {
		name string
		req  ShiftRequest
	}{
		{"overlap", ShiftRequest{Date: monday, StartHour: 12, EndHour: 15, TherapistID: sarah.ID, DivisionID: f.room}},
		{"end before start", ShiftRequest{Date: monday, StartHour: 15, EndHour: 14, TherapistID: sarah.ID, DivisionID: f.room}},
		{"not a physio", ShiftRequest{Date: monday, StartHour: 14, EndHour: 18, TherapistID: 999, DivisionID: f.room}},
		{"unknown division", ShiftRequest{Date: monday, StartHour: 14, EndHour: 18, TherapistID: sarah.ID, DivisionID: uuid.NewString()}},
		{"bad date", ShiftRequest{Date: "26/10/2026", StartHour: 14, EndHour: 18, TherapistID: sarah.ID, DivisionID: f.room}},
	}
	for _, tc := range cases {
		if _, err := f.svc.SaveShift(ctx, coordinator, "", tc.req); err == nil {
			t.Errorf("%s: expected an error", tc.name)
		}
	}
	// Back-to-back is fine
	f.shift(t, sarah.ID, f.room, 13, 18)

	for _, who := range []Actor{reception, doctor, sarahActor} {
		if _, err := f.svc.SaveShift(ctx, who, "", ShiftRequest{Date: monday, StartHour: 18, EndHour: 19, TherapistID: tiago.ID, DivisionID: f.gym}); !errors.Is(err, ErrForbidden) {
			t.Errorf("role %s may not manage shifts: %v", who.Role, err)
		}
	}
}

func TestBookingRules(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	gymShift := f.shift(t, sarah.ID, f.gym, 8, 13) // capacity 2 per hour
	f.shift(t, tiago.ID, f.room, 9, 12)            // capacity 1 per hour
	book := func(a Actor, who uint, hour int, patient string) (*Appointment, error) {
		return f.svc.Book(ctx, a, AppointmentRequest{Date: monday, Hour: hour, TherapistID: who, PatientID: patient})
	}
	third := uuid.NewString()
	f.repo.patients[third] = true

	first, err := book(reception, sarah.ID, 9, f.ana)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := book(coordinator, sarah.ID, 9, f.bruno); err != nil {
		t.Fatalf("second patient fits (capacity 2): %v", err)
	}
	if _, err := book(reception, sarah.ID, 9, third); !errors.Is(err, ErrFull) {
		t.Errorf("third patient: want ErrFull, got %v", err)
	}
	if _, err := book(reception, tiago.ID, 9, f.ana); !errors.Is(err, ErrPatientBusy) {
		t.Errorf("same patient twice at 9: %v", err)
	}
	var conflict *ConflictError
	if _, err := book(reception, sarah.ID, 14, third); !errors.As(err, &conflict) {
		t.Errorf("outside shift: %v", err)
	}
	if _, err := book(reception, sarah.ID, 9, uuid.NewString()); err == nil {
		t.Error("unknown patient should fail")
	}
	if _, err := book(doctor, sarah.ID, 10, third); !errors.Is(err, ErrForbidden) {
		t.Errorf("doctors do not book: %v", err)
	}

	// Sarah marks her own patient attended; may not remove bookings
	if ap, err := f.svc.SetStatus(ctx, sarahActor, first.ID, StatusRequest{Status: StatusAttended}); err != nil || ap.Status != StatusAttended {
		t.Errorf("own status: %v %v", ap, err)
	}
	if err := f.svc.Unbook(ctx, sarahActor, first.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("physio unbook: %v", err)
	}

	// A shift with patients cannot be deleted or shrunk past them
	if err := f.svc.DeleteShift(ctx, coordinator, gymShift.ID); !errors.As(err, &conflict) {
		t.Errorf("delete busy shift: %v", err)
	}
	if _, err := f.svc.SaveShift(ctx, coordinator, gymShift.ID, ShiftRequest{Date: monday, StartHour: 10, EndHour: 13, TherapistID: sarah.ID, DivisionID: f.gym}); !errors.As(err, &conflict) {
		t.Errorf("shrinking past booked hour 9: %v", err)
	}
	if _, err := f.svc.SaveShift(ctx, coordinator, gymShift.ID, ShiftRequest{Date: monday, StartHour: 8, EndHour: 11, TherapistID: sarah.ID, DivisionID: f.gym}); err != nil {
		t.Errorf("shrinking the end keeps hour 9: %v", err)
	}
}

func TestBoardAndCopy(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.shift(t, sarah.ID, f.gym, 7, 10)
	f.shift(t, sarah.ID, f.room, 14, 16)
	if _, err := f.svc.Book(ctx, reception, AppointmentRequest{Date: monday, Hour: 8, TherapistID: sarah.ID, PatientID: f.ana}); err != nil {
		t.Fatal(err)
	}

	b, err := f.svc.Board(ctx, doctor, monday)
	if err != nil {
		t.Fatal(err)
	}
	if b.OpenHour != 7 || b.CloseHour != defaultClose || len(b.Divisions) != 2 || b.Permissions.BookPatients {
		t.Fatalf("board header: %+v", b)
	}
	var row BoardTherapist
	for _, r := range b.Therapists {
		if r.ID == sarah.ID {
			row = r
		}
	}
	// 3 gym hours (capacity 2) + 2 room hours (capacity 1) = 5 hours, capacity 8, 1 patient
	if len(row.Hours) != 5 || row.Capacity != 8 || row.Patients != 1 {
		t.Fatalf("sarah row: hours=%d capacity=%d patients=%d", len(row.Hours), row.Capacity, row.Patients)
	}
	if h := row.Hours[1]; h.Hour != 8 || h.DivisionID != f.gym || len(h.Appointments) != 1 {
		t.Errorf("hour 8: %+v", h)
	}
	if len(b.Therapists) != 2 {
		t.Errorf("therapists without shifts still appear: %d", len(b.Therapists))
	}

	// Admin without a clinic sees the default clinic
	if _, err := f.svc.Board(ctx, admin, monday); err != nil {
		t.Errorf("admin board: %v", err)
	}

	n, err := f.svc.CopyDay(ctx, coordinator, CopyRequest{FromDate: monday, ToDate: "2026-10-27"})
	if err != nil || n != 2 {
		t.Fatalf("copy: %d %v", n, err)
	}
	if n, _ := f.svc.CopyDay(ctx, coordinator, CopyRequest{FromDate: monday, ToDate: "2026-10-27"}); n != 0 {
		t.Errorf("copying again skips physios already scheduled: %d", n)
	}
}
