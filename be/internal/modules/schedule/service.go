package schedule

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

var (
	// ErrForbidden: the role may not do this.
	ErrForbidden = errors.New("not allowed to perform this action")
	// ErrNoClinic: the user works in no clinic.
	ErrNoClinic = errors.New("you are not assigned to a clinic")
)

// ValidationError reports bad input with the offending field.
type ValidationError struct{ Field, Msg string }

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

// ConflictError explains why a change clashes with the board.
type ConflictError struct{ Msg string }

func (e *ConflictError) Error() string { return e.Msg }

// Actor is the authenticated user.
type Actor struct {
	ID   uint
	Role models.UserRole
}

// Who may do what on the time board:
//   - everyone in the clinic sees the board (who is where, how busy);
//   - coordinators and admins plan divisions and shifts and assign patients
//     to physiotherapists;
//   - a physiotherapist marks their own patients attended or missed.
func canManage(a Actor) bool { return auth.IsAdmin(a.Role) || a.Role == models.RoleCoordinator }

// Service holds the time-board rules.
type Service struct {
	repo    Repository
	clinics clinic.Resolver
	hook    SessionHook
	now     func() time.Time
}

// NewService builds the service.
func NewService(repo Repository, clinics clinic.Resolver) *Service {
	return &Service{repo: repo, clinics: clinics, hook: noHook{}, now: time.Now}
}

// WithSessionHook sets what happens when a session is held (e.g. using
// consumables) and returns the service.
func (s *Service) WithSessionHook(h SessionHook) *Service {
	s.hook = h
	return s
}

// clinicOf is the actor's clinic; admins without one get the default clinic.
func (s *Service) clinicOf(ctx context.Context, a Actor) (string, error) {
	home, err := s.clinics.HomeClinic(ctx, a.ID)
	if err != nil {
		return "", err
	}
	if home == nil && auth.IsAdmin(a.Role) {
		if home, err = s.repo.DefaultClinic(ctx); err != nil {
			return "", err
		}
	}
	if home == nil {
		return "", ErrNoClinic
	}
	return *home, nil
}

func parseDate(field, v string) (time.Time, error) {
	t, err := time.Parse(dateLayout, v)
	if err != nil {
		return time.Time{}, &ValidationError{field, "must be a date in YYYY-MM-DD format"}
	}
	return t, nil
}

func creditLimit(d *Division) int {
	if d.CreditLimit <= 0 {
		return DefaultCreditLimit
	}
	return d.CreditLimit
}

// ---- Divisions ----

// Divisions lists the clinic's active divisions.
func (s *Service) Divisions(ctx context.Context, a Actor) ([]Division, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	return s.repo.Divisions(ctx, cid)
}

// SaveDivision creates (id == "") or edits a division.
func (s *Service) SaveDivision(ctx context.Context, a Actor, id string, req DivisionRequest) (*Division, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	d := &Division{ClinicID: cid, Active: true}
	if id != "" {
		if d, err = s.division(ctx, cid, id); err != nil {
			return nil, err
		}
	}
	d.Name = strings.TrimSpace(req.Name)
	if len(d.Name) < 2 {
		return nil, &ValidationError{"name", "must have at least 2 characters"}
	}
	d.Capacity, d.SortOrder, d.Floor = req.Capacity, req.SortOrder, req.Floor
	if d.Floor == 0 {
		d.Floor = 1
	}
	d.CreditLimit = req.CreditLimit
	if d.CreditLimit == 0 {
		d.CreditLimit = DefaultCreditLimit
	}
	if req.Color != "" {
		d.Color = req.Color
	} else if d.Color == "" {
		d.Color = Colors[0]
	}
	return d, s.repo.SaveDivision(ctx, d)
}

// RemoveDivision hides a division from new shifts; past shifts keep it.
func (s *Service) RemoveDivision(ctx context.Context, a Actor, id string) error {
	if !canManage(a) {
		return ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return err
	}
	d, err := s.division(ctx, cid, id)
	if err != nil {
		return err
	}
	d.Active = false
	return s.repo.SaveDivision(ctx, d)
}

func (s *Service) division(ctx context.Context, cid, id string) (*Division, error) {
	if uuid.Validate(id) != nil {
		return nil, ErrNotFound
	}
	return s.repo.GetDivision(ctx, cid, id)
}

// ---- Shifts ----

// SaveShift creates (id == "") or edits a shift.
func (s *Service) SaveShift(ctx context.Context, a Actor, id string, req ShiftRequest) (*Shift, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	date, err := parseDate("date", req.Date)
	if err != nil {
		return nil, err
	}
	if req.EndHour <= req.StartHour {
		return nil, &ValidationError{"end_hour", "must be after the start"}
	}
	if ok, err := s.repo.IsTherapist(ctx, cid, req.TherapistID); err != nil || !ok {
		return nil, firstErr(err, &ValidationError{"therapist_id", "is not a physiotherapist of this clinic"})
	}
	div, err := s.division(ctx, cid, req.DivisionID)
	if err != nil || !div.Active {
		return nil, firstErr(notFoundOnly(err), &ValidationError{"division_id", "is not an active division of this clinic"})
	}

	sh := &Shift{ClinicID: cid, CreatedByID: a.ID}
	var old *Shift
	if id != "" {
		if old, err = s.shift(ctx, cid, id); err != nil {
			return nil, err
		}
		sh = old
		copyOld := *old
		old = &copyOld
	}
	sh.Date, sh.StartHour, sh.EndHour = date, req.StartHour, req.EndHour
	sh.TherapistID, sh.DivisionID, sh.Notes = req.TherapistID, req.DivisionID, strings.TrimSpace(req.Notes)

	// No overlap with the physiotherapist's other shifts that day
	day, err := s.repo.Shifts(ctx, cid, date, date)
	if err != nil {
		return nil, err
	}
	for _, o := range day {
		if o.ID != sh.ID && o.TherapistID == sh.TherapistID && o.StartHour < sh.EndHour && sh.StartHour < o.EndHour {
			return nil, &ConflictError{fmt.Sprintf("overlaps another shift (%02d:00-%02d:00)", o.StartHour, o.EndHour)}
		}
	}
	// Editing must not strand patients already booked in the old shift
	if old != nil {
		if err := s.checkStranded(ctx, cid, *old, sh); err != nil {
			return nil, err
		}
	}
	return sh, s.repo.SaveShift(ctx, sh)
}

// DeleteShift removes a shift that has no patients booked in it.
func (s *Service) DeleteShift(ctx context.Context, a Actor, id string) error {
	if !canManage(a) {
		return ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return err
	}
	sh, err := s.shift(ctx, cid, id)
	if err != nil {
		return err
	}
	if err := s.checkStranded(ctx, cid, *sh, nil); err != nil {
		return err
	}
	return s.repo.DeleteShift(ctx, id)
}

// CopyDay copies every shift of from onto to, skipping physiotherapists
// who already have a shift that day. It returns how many were copied.
func (s *Service) CopyDay(ctx context.Context, a Actor, req CopyRequest) (int, error) {
	if !canManage(a) {
		return 0, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return 0, err
	}
	from, err := parseDate("from_date", req.FromDate)
	if err != nil {
		return 0, err
	}
	to, err := parseDate("to_date", req.ToDate)
	if err != nil {
		return 0, err
	}
	if from.Equal(to) {
		return 0, &ValidationError{"to_date", "must differ from from_date"}
	}
	return s.copyDay(ctx, cid, a.ID, from, to)
}

// RepeatWeek copies the 7 days starting at WeekStart onto each of the next
// Weeks weeks (same weekday), skipping physiotherapists already scheduled.
func (s *Service) RepeatWeek(ctx context.Context, a Actor, req RepeatWeekRequest) (int, error) {
	if !canManage(a) {
		return 0, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return 0, err
	}
	start, err := parseDate("week_start", req.WeekStart)
	if err != nil {
		return 0, err
	}
	total := 0
	for w := 1; w <= req.Weeks; w++ {
		for d := 0; d < 7; d++ {
			from := start.AddDate(0, 0, d)
			n, err := s.copyDay(ctx, cid, a.ID, from, from.AddDate(0, 0, 7*w))
			total += n
			if err != nil {
				return total, err
			}
		}
	}
	return total, nil
}

func (s *Service) copyDay(ctx context.Context, cid string, actorID uint, from, to time.Time) (int, error) {
	src, err := s.repo.Shifts(ctx, cid, from, from)
	if err != nil {
		return 0, err
	}
	dst, err := s.repo.Shifts(ctx, cid, to, to)
	if err != nil {
		return 0, err
	}
	taken := map[uint]bool{}
	for _, d := range dst {
		taken[d.TherapistID] = true
	}
	n := 0
	for _, sh := range src {
		if taken[sh.TherapistID] {
			continue
		}
		c := Shift{ClinicID: cid, Date: to, StartHour: sh.StartHour, EndHour: sh.EndHour,
			TherapistID: sh.TherapistID, DivisionID: sh.DivisionID, Notes: sh.Notes, CreatedByID: actorID}
		if err := s.repo.SaveShift(ctx, &c); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (s *Service) shift(ctx context.Context, cid, id string) (*Shift, error) {
	if uuid.Validate(id) != nil {
		return nil, ErrNotFound
	}
	return s.repo.GetShift(ctx, cid, id)
}

// checkStranded fails if appointments inside old would fall outside next
// (nil = the shift is deleted).
func (s *Service) checkStranded(ctx context.Context, cid string, old Shift, next *Shift) error {
	appts, err := s.repo.Appointments(ctx, cid, old.Date, old.Date)
	if err != nil {
		return err
	}
	n := 0
	for _, ap := range appts {
		if ap.TherapistID != old.TherapistID || !old.Covers(ap.Hour) {
			continue
		}
		if next == nil || next.TherapistID != ap.TherapistID || !next.Date.Equal(ap.Date) || !next.Covers(ap.Hour) {
			n++
		}
	}
	if n > 0 {
		return &ConflictError{fmt.Sprintf("%d patient(s) are booked in this shift; move or remove them first", n)}
	}
	return nil
}

// ---- Appointments ----

// Book assigns a patient to a physiotherapist's hour. The physiotherapist
// must be on shift then and the division's patients and credits not yet used up.
func (s *Service) Book(ctx context.Context, a Actor, req AppointmentRequest) (*Appointment, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	date, err := parseDate("date", req.Date)
	if err != nil {
		return nil, err
	}
	level, ok, err := s.repo.PatientCareLevel(ctx, cid, req.PatientID)
	if err != nil || !ok {
		return nil, firstErr(err, &ValidationError{"patient_id", "is not a patient of this clinic"})
	}
	shifts, err := s.repo.Shifts(ctx, cid, date, date)
	if err != nil {
		return nil, err
	}
	on := shiftAt(shifts, req.TherapistID, req.Hour)
	if on == nil {
		return nil, &ConflictError{"the physiotherapist is not on shift at this hour"}
	}
	absences, err := s.repo.Absences(ctx, cid, date, date)
	if err != nil {
		return nil, err
	}
	if absentOn(absences, req.TherapistID, date) {
		return nil, &ConflictError{errAbsent.Error()}
	}
	div, err := s.repo.GetDivision(ctx, cid, on.DivisionID)
	if err != nil {
		return nil, err
	}
	ap := &Appointment{
		ClinicID: cid, Date: date, Hour: req.Hour, TherapistID: req.TherapistID, PatientID: req.PatientID,
		Credits: level, ReferralID: req.ReferralID, Status: StatusBooked, Notes: strings.TrimSpace(req.Notes), CreatedByID: a.ID,
	}
	if err := s.repo.Book(ctx, ap, div.Capacity, creditLimit(div)); err != nil {
		return nil, err
	}
	return ap, nil
}

func shiftAt(shifts []Shift, therapistID uint, hour int) *Shift {
	for i := range shifts {
		if shifts[i].TherapistID == therapistID && shifts[i].Covers(hour) {
			return &shifts[i]
		}
	}
	return nil
}

// SetStatus marks attended / missed: board managers or the physiotherapist
// of the appointment. Becoming attended triggers the session hook (e.g.
// gloves used automatically); leaving attended undoes it.
func (s *Service) SetStatus(ctx context.Context, a Actor, id string, req StatusRequest) (*StatusResult, error) {
	ap, err := s.appointment(ctx, a, id)
	if err != nil {
		return nil, err
	}
	if !canManage(a) && ap.TherapistID != a.ID {
		return nil, ErrForbidden
	}
	res := &StatusResult{Appointment: *ap, Warnings: []string{}}
	fields := missFields(req)
	// Same status: only a missed session's justification can still change
	if ap.Status == req.Status {
		if req.Status == StatusMissed {
			if err := s.repo.UpdateAppointment(ctx, id, fields); err != nil {
				return nil, err
			}
			j, _ := fields["miss_justified"].(bool)
			res.MissJustified, res.MissReason = &j, req.Reason
		}
		return res, nil
	}
	if err := s.repo.UpdateAppointment(ctx, id, fields); err != nil {
		return nil, err
	}
	if req.Status == StatusMissed {
		j, _ := fields["miss_justified"].(bool)
		res.MissJustified, res.MissReason = &j, req.Reason
	} else {
		res.MissJustified, res.MissReason = nil, ""
	}
	switch {
	case req.Status == StatusAttended:
		w, err := s.hook.SessionAttended(ctx, ap.ClinicID, ap.ID, s.divisionOf(ctx, ap), a.ID)
		if err != nil {
			return nil, err
		}
		res.Warnings = append(res.Warnings, w...)
	case ap.Status == StatusAttended:
		if err := s.hook.SessionUndone(ctx, ap.ClinicID, ap.ID, a.ID); err != nil {
			return nil, err
		}
	}
	res.Status = req.Status
	return res, nil
}

// divisionOf is the division of the shift the appointment falls in ("" if none).
func (s *Service) divisionOf(ctx context.Context, ap *Appointment) string {
	shifts, err := s.repo.Shifts(ctx, ap.ClinicID, ap.Date, ap.Date)
	if err != nil {
		return ""
	}
	if sh := shiftAt(shifts, ap.TherapistID, ap.Hour); sh != nil {
		return sh.DivisionID
	}
	return ""
}

// Unbook removes a patient from the hour.
func (s *Service) Unbook(ctx context.Context, a Actor, id string) error {
	if !canManage(a) {
		return ErrForbidden
	}
	ap, err := s.appointment(ctx, a, id)
	if err != nil {
		return err
	}
	if ap.Status == StatusAttended {
		if err := s.hook.SessionUndone(ctx, ap.ClinicID, ap.ID, a.ID); err != nil {
			return err
		}
	}
	return s.repo.DeleteAppointment(ctx, id)
}

func (s *Service) appointment(ctx context.Context, a Actor, id string) (*Appointment, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if uuid.Validate(id) != nil {
		return nil, ErrNotFound
	}
	return s.repo.GetAppointment(ctx, cid, id)
}

// ---- Board ----

// Board returns the day's board for the actor's clinic.
func (s *Service) Board(ctx context.Context, a Actor, date string) (*Board, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	d, err := parseDate("date", date)
	if err != nil {
		return nil, err
	}
	data, err := s.load(ctx, cid, d, d)
	if err != nil {
		return nil, err
	}
	b := buildBoard(d, data.divisions, data.therapists, data.shifts, data.appts)
	b.Permissions = permissions(a)
	b.Absences = data.absences
	return b, nil
}

func permissions(a Actor) Permissions {
	m := canManage(a)
	return Permissions{ManageShifts: m, BookPatients: m}
}

// boardData is everything of a date range needed to build boards.
type boardData struct {
	divisions  []Division
	therapists []UserSummary
	shifts     []Shift
	appts      []Appointment
	absences   []Absence
}

func (s *Service) load(ctx context.Context, cid string, from, to time.Time) (*boardData, error) {
	var d boardData
	var err error
	if d.divisions, err = s.repo.Divisions(ctx, cid); err != nil {
		return nil, err
	}
	if d.therapists, err = s.repo.Therapists(ctx, cid); err != nil {
		return nil, err
	}
	if d.shifts, err = s.repo.Shifts(ctx, cid, from, to); err != nil {
		return nil, err
	}
	if d.appts, err = s.repo.Appointments(ctx, cid, from, to); err != nil {
		return nil, err
	}
	if d.absences, err = s.repo.Absences(ctx, cid, from, to); err != nil {
		return nil, err
	}
	// Shifts may use divisions removed since; include them for display
	for _, sh := range d.shifts {
		if !hasDivision(d.divisions, sh.DivisionID) {
			if old, err := s.repo.GetDivision(ctx, cid, sh.DivisionID); err == nil {
				d.divisions = append(d.divisions, *old)
			}
		}
	}
	return &d, nil
}

func hasDivision(ds []Division, id string) bool {
	for _, d := range ds {
		if d.ID == id {
			return true
		}
	}
	return false
}

func firstErr(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

// notFoundOnly drops ErrNotFound so the caller reports a field error instead.
func notFoundOnly(err error) error {
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}
