package schedule

import (
	"context"
	"errors"
	"slices"
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
	patients   map[string]int // id → care level
	shifts     map[string]Shift
	appts      map[string]Appointment
	absences   []Absence
}

func between(d, from, to time.Time) bool {
	s := d.Format(dateLayout)
	return s >= from.Format(dateLayout) && s <= to.Format(dateLayout)
}

func (f *fakeRepo) DefaultClinic(context.Context) (*string, error) { return &f.clinic, nil }
func (f *fakeRepo) Divisions(_ context.Context, cid string) ([]Division, error) {
	var out []Division
	for _, d := range f.divisions {
		if d.ClinicID == cid && d.Active {
			out = append(out, d)
		}
	}
	slices.SortFunc(out, func(a, b Division) int { return a.SortOrder - b.SortOrder })
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
func (f *fakeRepo) PatientCareLevel(_ context.Context, cid, id string) (int, bool, error) {
	lvl, ok := f.patients[id]
	return lvl, ok && cid == f.clinic, nil
}
func (f *fakeRepo) Shifts(_ context.Context, cid string, from, to time.Time) ([]Shift, error) {
	var out []Shift
	for _, s := range f.shifts {
		if s.ClinicID == cid && between(s.Date, from, to) {
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
func (f *fakeRepo) Appointments(_ context.Context, cid string, from, to time.Time) ([]Appointment, error) {
	var out []Appointment
	for _, a := range f.appts {
		if a.ClinicID == cid && between(a.Date, from, to) {
			a.Patient = &PatientSummary{ID: a.PatientID, Name: "Patient " + a.PatientID[:4], CareLevel: a.Credits} // as Preload does
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
func (f *fakeRepo) Book(_ context.Context, a *Appointment, capacity, limit int) error {
	n, credits := 0, 0
	for _, o := range f.appts {
		if sameDate(o.Date, a.Date) && o.Hour == a.Hour {
			if o.PatientID == a.PatientID {
				return ErrPatientBusy
			}
			if o.TherapistID == a.TherapistID {
				n++
				credits += o.Credits
			}
		}
	}
	switch {
	case n >= capacity:
		return ErrFull
	case credits+a.Credits > limit:
		return ErrOverCredits
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
func (f *fakeRepo) DeleteSeriesFrom(_ context.Context, cid, seriesID string, from time.Time) (int64, error) {
	var n int64
	for id, a := range f.appts {
		if a.ClinicID == cid && a.SeriesID != nil && *a.SeriesID == seriesID && a.Status == StatusBooked &&
			a.Date.Format(dateLayout) >= from.Format(dateLayout) {
			delete(f.appts, id)
			n++
		}
	}
	return n, nil
}
func (f *fakeRepo) MoveAppointment(_ context.Context, a *Appointment, to uint, capacity, limit int) error {
	n, credits := 0, 0
	for _, o := range f.appts {
		if o.TherapistID == to && sameDate(o.Date, a.Date) && o.Hour == a.Hour {
			n++
			credits += o.Credits
		}
	}
	if n >= capacity {
		return ErrFull
	}
	if credits+a.Credits > limit {
		return ErrOverCredits
	}
	ap := f.appts[a.ID]
	ap.TherapistID = to
	f.appts[a.ID] = ap
	return nil
}
func (f *fakeRepo) CreateAbsence(_ context.Context, a *Absence) error {
	a.ID = uuid.NewString()
	f.absences = append(f.absences, *a)
	return nil
}
func (f *fakeRepo) Absences(_ context.Context, cid string, from, to time.Time) ([]Absence, error) {
	out := []Absence{}
	for _, a := range f.absences {
		if a.ClinicID == cid && a.From.Format(dateLayout) <= to.Format(dateLayout) && a.To.Format(dateLayout) >= from.Format(dateLayout) {
			out = append(out, a)
		}
	}
	return out, nil
}
func (f *fakeRepo) DeleteAbsence(_ context.Context, cid, id string) error {
	for i, a := range f.absences {
		if a.ID == id && a.ClinicID == cid {
			f.absences = append(f.absences[:i], f.absences[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
func (f *fakeRepo) CoveredBy(_ context.Context, coverID uint, at time.Time) ([]uint, error) {
	var out []uint
	for _, a := range f.absences {
		if a.CoverID != nil && *a.CoverID == coverID && a.Covers(at) {
			out = append(out, a.TherapistID)
		}
	}
	return out, nil
}
func (f *fakeRepo) MonthlyStats(context.Context, string, int) ([]MonthStats, error) {
	return make([]MonthStats, 12), nil
}

// recordingHook remembers session events.
type recordingHook struct{ attended, undone []string }

func (h *recordingHook) SessionAttended(_ context.Context, _, id, division string, _ uint) ([]string, error) {
	h.attended = append(h.attended, id+"@"+division)
	return []string{"Gloves: not enough in stock"}, nil
}
func (h *recordingHook) SessionUndone(_ context.Context, _, id string, _ uint) error {
	h.undone = append(h.undone, id)
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
	svc       *Service
	repo      *fakeRepo
	hook      *recordingHook
	gym, room string // divisions: gym 4 patients / 8 credits, room 1 patient / 8 credits
}

// patient adds a patient with a care level and returns its id.
func (f fixture) patient(level int) string {
	id := uuid.NewString()
	f.repo.patients[id] = level
	return id
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	cid := uuid.NewString()
	repo := &fakeRepo{clinic: cid, divisions: map[string]Division{}, therapists: []UserSummary{sarah, tiago},
		patients: map[string]int{}, shifts: map[string]Shift{}, appts: map[string]Appointment{}}
	clinics := clinic.Static{coordinator.ID: cid, reception.ID: cid, doctor.ID: cid, sarah.ID: cid}
	hook := &recordingHook{}
	f := fixture{svc: NewService(repo, clinics).WithSessionHook(hook), repo: repo, hook: hook}
	ctx := context.Background()
	gym, err := f.svc.SaveDivision(ctx, coordinator, "", DivisionRequest{Name: "Gym", Capacity: 4})
	if err != nil {
		t.Fatal(err)
	}
	if gym.CreditLimit != DefaultCreditLimit {
		t.Fatalf("default credit limit: %d", gym.CreditLimit)
	}
	room, _ := f.svc.SaveDivision(ctx, coordinator, "", DivisionRequest{Name: "Room 1", Color: "indigo", Capacity: 1, SortOrder: 1})
	f.gym, f.room = gym.ID, room.ID
	return f
}

func (f fixture) shift(t *testing.T, date string, who uint, div string, from, to int) *Shift {
	t.Helper()
	s, err := f.svc.SaveShift(context.Background(), coordinator, "", ShiftRequest{
		Date: date, StartHour: from, EndHour: to, TherapistID: who, DivisionID: div,
	})
	if err != nil {
		t.Fatalf("shift %s %d-%d: %v", date, from, to, err)
	}
	return s
}

func (f fixture) book(a Actor, who uint, hour int, patient string) (*Appointment, error) {
	return f.svc.Book(context.Background(), a, AppointmentRequest{Date: monday, Hour: hour, TherapistID: who, PatientID: patient})
}

func TestShiftRules(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.shift(t, monday, sarah.ID, f.gym, 8, 13)

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
	f.shift(t, monday, sarah.ID, f.room, 13, 18) // back-to-back is fine

	for _, who := range []Actor{reception, doctor, sarahActor} {
		if _, err := f.svc.SaveShift(ctx, who, "", ShiftRequest{Date: monday, StartHour: 18, EndHour: 19, TherapistID: tiago.ID, DivisionID: f.gym}); !errors.Is(err, ErrForbidden) {
			t.Errorf("role %s may not manage shifts: %v", who.Role, err)
		}
	}
}

func TestBookingRulesAndCredits(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	gymShift := f.shift(t, monday, sarah.ID, f.gym, 8, 13)
	f.shift(t, monday, tiago.ID, f.room, 9, 12)

	// Credits: 8 per hour. Two level-5 patients never fit; 5 + 3 does.
	neuro1, neuro2, mid, light := f.patient(5), f.patient(5), f.patient(3), f.patient(1)
	first, err := f.book(coordinator, sarah.ID, 9, neuro1)
	if err != nil || first.Credits != 5 {
		t.Fatalf("first: %+v %v", first, err)
	}
	if _, err := f.book(coordinator, sarah.ID, 9, neuro2); !errors.Is(err, ErrOverCredits) {
		t.Errorf("two level-5 patients: want ErrOverCredits, got %v", err)
	}
	if _, err := f.book(coordinator, sarah.ID, 9, mid); err != nil {
		t.Errorf("5 + 3 = 8 fits: %v", err)
	}
	if _, err := f.book(coordinator, sarah.ID, 9, light); !errors.Is(err, ErrOverCredits) {
		t.Errorf("9 credits: %v", err)
	}
	// Patient limit still applies (room takes 1 patient)
	if _, err := f.book(coordinator, tiago.ID, 10, light); err != nil {
		t.Fatal(err)
	}
	if _, err := f.book(coordinator, tiago.ID, 10, neuro2); !errors.Is(err, ErrFull) {
		t.Errorf("room is full: %v", err)
	}
	if _, err := f.book(coordinator, tiago.ID, 9, neuro1); !errors.Is(err, ErrPatientBusy) {
		t.Errorf("same patient twice at 9: %v", err)
	}
	var conflict *ConflictError
	if _, err := f.book(coordinator, sarah.ID, 14, light); !errors.As(err, &conflict) {
		t.Errorf("outside shift: %v", err)
	}
	if _, err := f.book(coordinator, sarah.ID, 9, uuid.NewString()); err == nil {
		t.Error("unknown patient should fail")
	}
	for _, who := range []Actor{reception, doctor, sarahActor} {
		if _, err := f.book(who, sarah.ID, 10, light); !errors.Is(err, ErrForbidden) {
			t.Errorf("%s does not assign patients: %v", who.Role, err)
		}
	}

	// Sarah marks her own patient attended (hook runs once, warnings pass through)
	res, err := f.svc.SetStatus(ctx, sarahActor, first.ID, StatusRequest{Status: StatusAttended})
	if err != nil || res.Status != StatusAttended || len(res.Warnings) != 1 {
		t.Fatalf("attended: %+v %v", res, err)
	}
	if f.hook.attended[0] != first.ID+"@"+f.gym {
		t.Errorf("the hook gets the session's division: %v", f.hook.attended)
	}
	if _, err := f.svc.SetStatus(ctx, sarahActor, first.ID, StatusRequest{Status: StatusAttended}); err != nil || len(f.hook.attended) != 1 {
		t.Errorf("repeating attended must not use stock twice: %v %v", err, f.hook.attended)
	}
	if _, err := f.svc.SetStatus(ctx, sarahActor, first.ID, StatusRequest{Status: StatusMissed}); err != nil || len(f.hook.undone) != 1 {
		t.Errorf("attended → missed gives the stock back: %v %v", err, f.hook.undone)
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
}

func TestSeries(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	// Sarah works Mon 26, Wed 28 and Fri 30 Oct, then Mon 2 and Wed 4 Nov
	for _, d := range []string{"2026-10-26", "2026-10-28", "2026-10-30", "2026-11-02", "2026-11-04"} {
		f.shift(t, d, sarah.ID, f.room, 9, 12)
	}
	other := f.patient(1)
	if _, err := f.svc.Book(ctx, coordinator, AppointmentRequest{Date: "2026-10-28", Hour: 10, TherapistID: sarah.ID, PatientID: other}); err != nil {
		t.Fatal(err)
	}
	pat := f.patient(2)
	req := SeriesRequest{PatientID: pat, TherapistID: sarah.ID, Weekdays: []int{1, 3, 5}, Hour: 10,
		StartDate: "2026-10-26", Sessions: 4, DryRun: true}

	preview, err := f.svc.Series(ctx, coordinator, req)
	if err != nil {
		t.Fatal(err)
	}
	// Wed 28 is taken (room holds 1) → skipped; four sessions need Mon 2 Nov too
	if !slices.Equal(preview.Booked, []string{"2026-10-26", "2026-10-30", "2026-11-02", "2026-11-04"}) ||
		len(preview.Skipped) != 1 || preview.Skipped[0] != (SkippedDate{"2026-10-28", SkipFull}) || preview.SeriesID != "" {
		t.Fatalf("preview: %+v", preview)
	}
	if len(f.repo.appts) != 1 {
		t.Fatal("a dry run must not book")
	}

	req.DryRun = false
	res, err := f.svc.Series(ctx, coordinator, req)
	if err != nil || len(res.Booked) != 4 || res.SeriesID == "" {
		t.Fatalf("series: %+v %v", res, err)
	}
	if len(f.repo.appts) != 5 {
		t.Fatalf("appointments: %d", len(f.repo.appts))
	}

	// Beyond the planned shifts there is nothing to book
	more := req
	more.StartDate, more.Sessions, more.Weekdays = "2026-11-05", 1, []int{4}
	if r, _ := f.svc.Series(ctx, coordinator, more); len(r.Booked) != 0 || r.Skipped[0].Reason != SkipNotOnShift {
		t.Errorf("no shift: %+v", r)
	}
	if _, err := f.svc.Series(ctx, reception, req); !errors.Is(err, ErrForbidden) {
		t.Errorf("receptionist plans: %v", err)
	}

	n, err := f.svc.CancelSeries(ctx, coordinator, res.SeriesID, "2026-11-01")
	if err != nil || n != 2 {
		t.Errorf("cancel from 1 Nov: %d %v", n, err)
	}
}

func TestBoardRangeAndRepeat(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.shift(t, monday, sarah.ID, f.gym, 7, 10)
	f.shift(t, monday, sarah.ID, f.room, 14, 16)
	pat := f.patient(3)
	ap, err := f.book(coordinator, sarah.ID, 8, pat)
	if err != nil {
		t.Fatal(err)
	}
	f.svc.SetStatus(ctx, coordinator, ap.ID, StatusRequest{Status: StatusAttended})

	b, err := f.svc.Board(ctx, doctor, monday)
	if err != nil {
		t.Fatal(err)
	}
	if b.OpenHour != 7 || len(b.Divisions) != 2 || b.Permissions.BookPatients || len(b.Therapists) != 2 {
		t.Fatalf("board header: %+v", b)
	}
	row := b.Therapists[0]
	// 3 gym hours (4 patients, 8 credits) + 2 room hours (1, 8) = 14 patients, 40 credits
	if row.ID != sarah.ID || len(row.Hours) != 5 || row.Capacity != 14 || row.CreditCapacity != 40 || row.Credits != 3 {
		t.Fatalf("sarah: %+v", row)
	}
	if h := row.Hours[1]; h.Hour != 8 || h.Credits != 3 || h.CreditLimit != 8 {
		t.Errorf("hour 8: %+v", h)
	}
	if _, err := f.svc.Board(ctx, admin, monday); err != nil {
		t.Errorf("admin board: %v", err)
	}

	// Repeat the week (Mon 26 Oct) for 2 more weeks: Sarah's 2 shifts × 2
	n, err := f.svc.RepeatWeek(ctx, coordinator, RepeatWeekRequest{WeekStart: monday, Weeks: 2})
	if err != nil || n != 4 {
		t.Fatalf("repeat week: %d %v", n, err)
	}

	v, err := f.svc.Range(ctx, coordinator, monday, "2026-11-01")
	if err != nil || len(v.Days) != 7 {
		t.Fatalf("range: %v", err)
	}
	mon := v.Days[0]
	if mon.Totals.Patients != 1 || mon.Totals.Attended != 1 || mon.Totals.Credits != 3 || len(mon.Therapists) != 1 {
		t.Errorf("monday totals: %+v", mon)
	}
	if tue := v.Days[1]; len(tue.Therapists) != 0 || tue.Totals.Capacity != 0 {
		t.Errorf("tuesday is empty: %+v", tue)
	}
	if _, err := f.svc.Range(ctx, coordinator, monday, "2026-12-31"); err == nil {
		t.Error("ranges longer than 42 days are refused")
	}
	if n, _ := f.svc.CopyDay(ctx, coordinator, CopyRequest{FromDate: monday, ToDate: "2026-11-02"}); n != 0 {
		t.Errorf("copying onto a planned day skips: %d", n)
	}
}

func TestAbsenceAndCover(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.svc.now = func() time.Time { return time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC) }
	tuesday := "2026-10-27"
	for _, d := range []string{monday, tuesday} {
		f.shift(t, d, sarah.ID, f.gym, 8, 13)
		f.shift(t, d, sarah.ID, f.room, 14, 16)
		f.shift(t, d, tiago.ID, f.gym, 9, 12)
	}
	ana, bruno := f.patient(2), f.patient(5)
	morning, _ := f.book(coordinator, sarah.ID, 9, ana)  // Tiago works at 9 → moves
	later, _ := f.book(coordinator, sarah.ID, 14, bruno) // Tiago is off at 14 → stays
	cover := tiago.ID
	req := AbsenceRequest{TherapistID: sarah.ID, CoverID: &cover, From: monday, To: tuesday, Reason: "vacation",
		Handover: "Plano de exercícios na pasta partilhada.", MovePatients: true, DryRun: true}
	req.Notes = []HandoverInput{{ana, "Progredir carga no agachamento."}}

	preview, err := f.svc.CreateAbsence(ctx, coordinator, req)
	if err != nil || len(preview.Sessions) != 2 || len(preview.Patients) != 2 || preview.Absence != nil {
		t.Fatalf("preview: %+v %v", preview, err)
	}
	if f.repo.appts[morning.ID].TherapistID != sarah.ID || len(f.repo.absences) != 0 {
		t.Fatal("a preview must not change anything")
	}

	req.DryRun = false
	res, err := f.svc.CreateAbsence(ctx, coordinator, req)
	if err != nil || res.Absence == nil {
		t.Fatal(err)
	}
	byID := map[string]MovedSession{}
	for _, s := range res.Sessions {
		byID[s.AppointmentID] = s
	}
	if !byID[morning.ID].Moved || byID[later.ID].Moved || byID[later.ID].Reason != SkipNotOnShift {
		t.Fatalf("sessions: %+v", res.Sessions)
	}
	if f.repo.appts[morning.ID].TherapistID != tiago.ID || f.repo.appts[later.ID].TherapistID != sarah.ID {
		t.Error("the 9:00 session goes to Tiago, the 14:00 one stays to be placed by hand")
	}
	// Sarah's empty shifts are gone; the one still holding Bruno stays
	left := 0
	for _, s := range f.repo.shifts {
		if s.TherapistID == sarah.ID {
			left++
		}
	}
	if left != 1 {
		t.Errorf("sarah's shifts left: %d", left)
	}
	var conflict *ConflictError
	if _, err := f.svc.Book(ctx, coordinator, AppointmentRequest{Date: monday, Hour: 14, TherapistID: sarah.ID, PatientID: f.patient(1)}); !errors.As(err, &conflict) {
		t.Errorf("booking an absent physio: %v", err)
	}
	if _, err := f.svc.CreateAbsence(ctx, coordinator, AbsenceRequest{TherapistID: sarah.ID, From: tuesday, To: tuesday, Reason: "sick"}); !errors.As(err, &conflict) {
		t.Errorf("overlapping absence: %v", err)
	}
	if _, err := f.svc.CreateAbsence(ctx, sarahActor, req); !errors.Is(err, ErrForbidden) {
		t.Errorf("physios do not record absences: %v", err)
	}

	notes, err := f.svc.Handover(ctx, sarahActor, ana)
	if err != nil || len(notes) != 1 || notes[0].Note != "Progredir carga no agachamento." {
		t.Fatalf("handover: %+v %v", notes, err)
	}
	covered, _ := f.svc.CoveredBy(ctx, tiago.ID, time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC))
	if !slices.Equal(covered, []uint{sarah.ID}) {
		t.Errorf("tiago covers sarah on monday: %v", covered)
	}
	b, _ := f.svc.Board(ctx, coordinator, monday)
	if len(b.Absences) != 1 {
		t.Errorf("the board shows the absence: %d", len(b.Absences))
	}
	if err := f.svc.DeleteAbsence(ctx, coordinator, res.Absence.ID); err != nil {
		t.Errorf("delete: %v", err)
	}
}
