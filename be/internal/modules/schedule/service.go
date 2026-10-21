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
//   - coordinators and admins manage divisions and shifts;
//   - they and receptionists assign patients to a physiotherapist's hour;
//   - a physiotherapist can mark their own patients attended or missed.
func canManage(a Actor) bool { return auth.IsAdmin(a.Role) || a.Role == models.RoleCoordinator }
func canBook(a Actor) bool   { return canManage(a) || a.Role == models.RoleReceptionist }

// Service holds the time-board rules.
type Service struct {
	repo    Repository
	clinics clinic.Resolver
}

// NewService builds the service.
func NewService(repo Repository, clinics clinic.Resolver) *Service {
	return &Service{repo: repo, clinics: clinics}
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
	d.Capacity, d.SortOrder = req.Capacity, req.SortOrder
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
	day, err := s.repo.Shifts(ctx, cid, date)
	if err != nil {
		return nil, err
	}
	for _, o := range day {
		if o.ID != sh.ID && o.TherapistID == sh.TherapistID && o.StartHour < sh.EndHour && sh.StartHour < o.EndHour {
			return nil, &ConflictError{fmt.Sprintf("overlaps another shift (%02d:00–%02d:00)", o.StartHour, o.EndHour)}
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
	src, err := s.repo.Shifts(ctx, cid, from)
	if err != nil {
		return 0, err
	}
	dst, err := s.repo.Shifts(ctx, cid, to)
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
			TherapistID: sh.TherapistID, DivisionID: sh.DivisionID, Notes: sh.Notes, CreatedByID: a.ID}
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
	appts, err := s.repo.Appointments(ctx, cid, old.Date)
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
// must be on shift then and the division's capacity not yet reached.
func (s *Service) Book(ctx context.Context, a Actor, req AppointmentRequest) (*Appointment, error) {
	if !canBook(a) {
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
	if ok, err := s.repo.PatientInClinic(ctx, cid, req.PatientID); err != nil || !ok {
		return nil, firstErr(err, &ValidationError{"patient_id", "is not a patient of this clinic"})
	}
	shifts, err := s.repo.Shifts(ctx, cid, date)
	if err != nil {
		return nil, err
	}
	var on *Shift
	for i := range shifts {
		if shifts[i].TherapistID == req.TherapistID && shifts[i].Covers(req.Hour) {
			on = &shifts[i]
		}
	}
	if on == nil {
		return nil, &ConflictError{"the physiotherapist is not on shift at this hour"}
	}
	div, err := s.repo.GetDivision(ctx, cid, on.DivisionID)
	if err != nil {
		return nil, err
	}
	ap := &Appointment{
		ClinicID: cid, Date: date, Hour: req.Hour, TherapistID: req.TherapistID, PatientID: req.PatientID,
		ReferralID: req.ReferralID, Status: StatusBooked, Notes: strings.TrimSpace(req.Notes), CreatedByID: a.ID,
	}
	if err := s.repo.Book(ctx, ap, div.Capacity); err != nil {
		return nil, err
	}
	return ap, nil
}

// SetStatus marks attended / missed: board managers, front desk or the
// physiotherapist of the appointment.
func (s *Service) SetStatus(ctx context.Context, a Actor, id string, req StatusRequest) (*Appointment, error) {
	ap, err := s.appointment(ctx, a, id)
	if err != nil {
		return nil, err
	}
	if !canBook(a) && ap.TherapistID != a.ID {
		return nil, ErrForbidden
	}
	if err := s.repo.UpdateAppointment(ctx, id, map[string]any{"status": req.Status}); err != nil {
		return nil, err
	}
	ap.Status = req.Status
	return ap, nil
}

// Unbook removes a patient from the hour.
func (s *Service) Unbook(ctx context.Context, a Actor, id string) error {
	if !canBook(a) {
		return ErrForbidden
	}
	if _, err := s.appointment(ctx, a, id); err != nil {
		return err
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
	divs, err := s.repo.Divisions(ctx, cid)
	if err != nil {
		return nil, err
	}
	therapists, err := s.repo.Therapists(ctx, cid)
	if err != nil {
		return nil, err
	}
	shifts, err := s.repo.Shifts(ctx, cid, d)
	if err != nil {
		return nil, err
	}
	appts, err := s.repo.Appointments(ctx, cid, d)
	if err != nil {
		return nil, err
	}
	// Shifts may use divisions removed since; include them for display
	for _, sh := range shifts {
		if !hasDivision(divs, sh.DivisionID) {
			if old, err := s.repo.GetDivision(ctx, cid, sh.DivisionID); err == nil {
				divs = append(divs, *old)
			}
		}
	}
	b := buildBoard(d, divs, therapists, shifts, appts)
	b.Permissions = Permissions{ManageShifts: canManage(a), BookPatients: canBook(a)}
	return b, nil
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
