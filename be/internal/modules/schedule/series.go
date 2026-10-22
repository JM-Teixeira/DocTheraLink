package schedule

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Reasons a plan date is skipped.
const (
	SkipNotOnShift  = "not_on_shift"
	SkipFull        = "full"
	SkipOverCredits = "over_credits"
	SkipPatientBusy = "patient_busy"
	SkipAbsent      = "absent"
)

// Planning looks at most this far ahead and gives up after this many skips.
const (
	seriesHorizonDays = 366
	seriesMaxSkips    = 60
)

// Series books a weekly treatment plan (e.g. Mon/Wed/Fri at 10:00 until 12
// sessions). Dates that do not fit are skipped and reported, and the plan
// continues until the number of sessions is reached. With DryRun nothing is
// saved: the result is a preview.
func (s *Service) Series(ctx context.Context, a Actor, req SeriesRequest) (*SeriesResult, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	start, err := parseDate("start_date", req.StartDate)
	if err != nil {
		return nil, err
	}
	level, ok, err := s.repo.PatientCareLevel(ctx, cid, req.PatientID)
	if err != nil || !ok {
		return nil, firstErr(err, &ValidationError{"patient_id", "is not a patient of this clinic"})
	}
	if ok, err := s.repo.IsTherapist(ctx, cid, req.TherapistID); err != nil || !ok {
		return nil, firstErr(err, &ValidationError{"therapist_id", "is not a physiotherapist of this clinic"})
	}
	weekdays := map[time.Weekday]bool{}
	for _, d := range req.Weekdays {
		weekdays[time.Weekday(d)] = true
	}

	end := start.AddDate(0, 0, seriesHorizonDays)
	data, err := s.load(ctx, cid, start, end)
	if err != nil {
		return nil, err
	}
	plan := planner{data: data, level: level}

	res := &SeriesResult{Booked: []string{}, Skipped: []SkippedDate{}}
	if !req.DryRun {
		res.SeriesID = uuid.NewString()
	}
	for d := start; d.Before(end) && len(res.Booked) < req.Sessions && len(res.Skipped) < seriesMaxSkips; d = d.AddDate(0, 0, 1) {
		if !weekdays[d.Weekday()] {
			continue
		}
		date := d.Format(dateLayout)
		reason, div := plan.problem(d, req.Hour, req.TherapistID, req.PatientID)
		if reason == "" && !req.DryRun {
			ap := &Appointment{ClinicID: cid, Date: d, Hour: req.Hour, TherapistID: req.TherapistID,
				PatientID: req.PatientID, Credits: level, ReferralID: req.ReferralID, SeriesID: &res.SeriesID,
				Status: StatusBooked, Notes: strings.TrimSpace(req.Notes), CreatedByID: a.ID}
			if err := s.repo.Book(ctx, ap, div.Capacity, creditLimit(div)); err != nil {
				if reason = skipReason(err); reason == "" {
					return nil, err
				}
			}
		}
		if reason != "" {
			res.Skipped = append(res.Skipped, SkippedDate{Date: date, Reason: reason})
			continue
		}
		res.Booked = append(res.Booked, date)
		plan.add(d, req.Hour, req.TherapistID, req.PatientID)
	}
	return res, nil
}

// CancelSeries removes a plan's still-booked sessions from a date on
// (attended and missed ones are kept as history).
func (s *Service) CancelSeries(ctx context.Context, a Actor, id, from string) (int64, error) {
	if !canManage(a) {
		return 0, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return 0, err
	}
	if uuid.Validate(id) != nil {
		return 0, ErrNotFound
	}
	d, err := parseDate("from", from)
	if err != nil {
		return 0, err
	}
	return s.repo.DeleteSeriesFrom(ctx, cid, id, d)
}

func skipReason(err error) string {
	switch {
	case errors.Is(err, ErrFull):
		return SkipFull
	case errors.Is(err, ErrOverCredits):
		return SkipOverCredits
	case errors.Is(err, ErrPatientBusy):
		return SkipPatientBusy
	}
	return ""
}

// planner checks plan dates against the loaded board data, counting the
// sessions it has planned so far (so a preview is accurate too).
type planner struct {
	data    *boardData
	level   int
	exclude map[string]bool // appointments left out of the checks (they are moving)
}

func (p *planner) problem(d time.Time, hour int, therapistID uint, patientID string) (string, *Division) {
	var on *Shift
	for i := range p.data.shifts {
		sh := &p.data.shifts[i]
		if sameDate(sh.Date, d) && sh.TherapistID == therapistID && sh.Covers(hour) {
			on = sh
		}
	}
	if on == nil {
		return SkipNotOnShift, nil
	}
	if absentOn(p.data.absences, therapistID, d) {
		return SkipAbsent, nil
	}
	var div *Division
	for i := range p.data.divisions {
		if p.data.divisions[i].ID == on.DivisionID {
			div = &p.data.divisions[i]
		}
	}
	if div == nil {
		return SkipNotOnShift, nil
	}
	n, credits := 0, 0
	for _, ap := range p.data.appts {
		if p.exclude[ap.ID] && ap.ID != "" {
			continue
		}
		if !sameDate(ap.Date, d) || ap.Hour != hour {
			continue
		}
		if ap.PatientID == patientID {
			return SkipPatientBusy, div
		}
		if ap.TherapistID == therapistID {
			n++
			credits += ap.Credits
		}
	}
	switch {
	case n >= div.Capacity:
		return SkipFull, div
	case credits+p.level > creditLimit(div):
		return SkipOverCredits, div
	}
	return "", div
}

// room is a physio's free places and credits at an hour (ok = on shift and
// not absent), counting what this plan already gave them.
func (p *planner) room(d time.Time, hour int, therapistID uint) (places, credits int, ok bool) {
	_, div := p.problem(d, hour, therapistID, "")
	if div == nil || absentOn(p.data.absences, therapistID, d) {
		return 0, 0, false
	}
	n, used := 0, 0
	for _, ap := range p.data.appts {
		if (p.exclude[ap.ID] && ap.ID != "") || !sameDate(ap.Date, d) || ap.Hour != hour || ap.TherapistID != therapistID {
			continue
		}
		n++
		used += ap.Credits
	}
	return max(div.Capacity-n, 0), max(creditLimit(div)-used, 0), true
}

func (p *planner) add(d time.Time, hour int, therapistID uint, patientID string) {
	p.data.appts = append(p.data.appts, Appointment{Date: d, Hour: hour, TherapistID: therapistID,
		PatientID: patientID, Credits: p.level})
}
