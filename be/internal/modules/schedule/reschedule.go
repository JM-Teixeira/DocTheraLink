package schedule

import (
	"context"

	"github.com/google/uuid"
)

// MoveRequest moves one booked session to another date, hour and/or physio.
type MoveRequest struct {
	Date        string `json:"date" binding:"required"`
	Hour        int    `json:"hour" binding:"min=0,max=23"`
	TherapistID uint   `json:"therapist_id" binding:"required"`
}

// PlanRescheduleRequest gives the rest of a weekly plan a new pattern
// (e.g. Mon/Wed/Fri 9:00 becomes Tue/Thu 17:00 with another physio) from a
// date on, keeping the number of sessions still to come.
type PlanRescheduleRequest struct {
	From        string `json:"from" binding:"required"`
	Weekdays    []int  `json:"weekdays" binding:"required,min=1,max=7,dive,min=0,max=6"`
	Hour        int    `json:"hour" binding:"min=0,max=23"`
	TherapistID uint   `json:"therapist_id" binding:"required"`
	DryRun      bool   `json:"dry_run"`
}

// PlanRescheduleResult is the new plan and how many sessions it replaces.
type PlanRescheduleResult struct {
	SeriesResult
	Replaced int `json:"replaced"`
}

// conflictFor turns a planner problem into the error a booking would give.
func conflictFor(reason string) error {
	switch reason {
	case SkipFull:
		return ErrFull
	case SkipOverCredits:
		return ErrOverCredits
	case SkipPatientBusy:
		return ErrPatientBusy
	case SkipAbsent:
		return &ConflictError{errAbsent.Error()}
	}
	return &ConflictError{"the physiotherapist is not working at that hour"}
}

// Move takes a booked session to another slot (drag and drop on the board):
// same rules as booking - the physio works then, the hour has room and care
// credits, the patient is not elsewhere and the physio is not away.
func (s *Service) Move(ctx context.Context, a Actor, id string, req MoveRequest) (*Appointment, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	ap, err := s.appointment(ctx, a, id)
	if err != nil {
		return nil, err
	}
	if ap.Status != StatusBooked {
		return nil, &ValidationError{"status", "only booked sessions can be moved"}
	}
	d, err := parseDate("date", req.Date)
	if err != nil {
		return nil, err
	}
	if sameDate(d, ap.Date) && req.Hour == ap.Hour && req.TherapistID == ap.TherapistID {
		return ap, nil
	}
	data, err := s.load(ctx, ap.ClinicID, d, d)
	if err != nil {
		return nil, err
	}
	plan := planner{data: data, level: ap.Credits, exclude: map[string]bool{ap.ID: true}}
	reason, div := plan.problem(d, req.Hour, req.TherapistID, ap.PatientID)
	if reason != "" {
		return nil, conflictFor(reason)
	}
	if err := s.repo.Reschedule(ctx, ap, d, req.Hour, req.TherapistID, div.Capacity, creditLimit(div)); err != nil {
		return nil, err
	}
	return s.appointment(ctx, a, id)
}

// ReschedulePlan moves the still-booked sessions of a plan from a date on to
// a new weekly pattern. DryRun previews (the plan's own sessions do not get
// in the way); otherwise they are replaced and the plan keeps its id.
func (s *Service) ReschedulePlan(ctx context.Context, a Actor, seriesID string, req PlanRescheduleRequest) (*PlanRescheduleResult, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if uuid.Validate(seriesID) != nil {
		return nil, ErrNotFound
	}
	from, err := parseDate("from", req.From)
	if err != nil {
		return nil, err
	}
	rest, err := s.repo.SeriesAppointments(ctx, cid, seriesID, from)
	if err != nil {
		return nil, err
	}
	if len(rest) == 0 {
		return nil, ErrNotFound
	}
	exclude := map[string]bool{}
	for _, ap := range rest {
		exclude[ap.ID] = true
	}
	req2 := SeriesRequest{PatientID: rest[0].PatientID, TherapistID: req.TherapistID, Weekdays: req.Weekdays,
		Hour: req.Hour, StartDate: req.From, Sessions: len(rest), ReferralID: rest[0].ReferralID,
		Notes: rest[0].Notes, DryRun: true}
	preview, err := s.bookSeries(ctx, a, cid, req2, exclude, "")
	if err != nil {
		return nil, err
	}
	if req.DryRun {
		return &PlanRescheduleResult{SeriesResult: *preview, Replaced: len(rest)}, nil
	}
	if _, err := s.repo.DeleteSeriesFrom(ctx, cid, seriesID, from); err != nil {
		return nil, err
	}
	req2.DryRun = false
	res, err := s.bookSeries(ctx, a, cid, req2, nil, seriesID)
	if err != nil {
		return nil, err
	}
	return &PlanRescheduleResult{SeriesResult: *res, Replaced: len(rest)}, nil
}

// PatientSessions are a patient's sessions in the user's clinic (newest
// first, at most 500), for the patient timeline.
func (s *Service) PatientSessions(ctx context.Context, a Actor, patientID string) ([]Appointment, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if uuid.Validate(patientID) != nil {
		return nil, ErrNotFound
	}
	return s.repo.PatientAppointments(ctx, cid, patientID, 500)
}

// SeriesSessions lists a plan's still-booked sessions from today (for the
// "reschedule plan" dialog).
func (s *Service) SeriesSessions(ctx context.Context, a Actor, seriesID string) ([]Appointment, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if uuid.Validate(seriesID) != nil {
		return nil, ErrNotFound
	}
	out, err := s.repo.SeriesAppointments(ctx, cid, seriesID, s.today())
	if out == nil {
		out = []Appointment{}
	}
	return out, err
}
