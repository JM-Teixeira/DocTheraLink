package schedule

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Absence: a physiotherapist away (holiday, illness, training…) from one
// date to another, optionally covered by a colleague who takes over their
// patients and reads the handover.
type Absence struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ClinicID    string         `json:"clinic_id" gorm:"type:uuid;not null;index"`
	TherapistID uint           `json:"therapist_id" gorm:"not null;index"`
	CoverID     *uint          `json:"cover_id" gorm:"index"`
	From        time.Time      `json:"from" gorm:"type:date;not null"`
	To          time.Time      `json:"to" gorm:"type:date;not null"` // inclusive
	Reason      string         `json:"reason" gorm:"size:12;not null"`
	Handover    string         `json:"handover" gorm:"type:text"` // general handover for the cover
	Notes       []HandoverNote `json:"notes" gorm:"foreignKey:AbsenceID"`
	Therapist   *UserSummary   `json:"therapist,omitempty" gorm:"foreignKey:TherapistID;->;-:migration"`
	Cover       *UserSummary   `json:"cover,omitempty" gorm:"foreignKey:CoverID;->;-:migration"`
	CreatedByID uint           `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_date"`
}

// Covers reports whether the absence includes the date.
func (a Absence) Covers(d time.Time) bool {
	s := d.Format(dateLayout)
	return s >= a.From.Format(dateLayout) && s <= a.To.Format(dateLayout)
}

// HandoverNote is what the absent physio leaves about one patient.
type HandoverNote struct {
	ID        string          `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AbsenceID string          `json:"absence_id" gorm:"type:uuid;not null;index"`
	PatientID string          `json:"patient_id" gorm:"type:uuid;not null;index"`
	Patient   *PatientSummary `json:"patient,omitempty" gorm:"foreignKey:PatientID;->;-:migration"`
	Note      string          `json:"note" gorm:"type:text;not null"`
}

// Absence reasons.
var absenceReasons = map[string]bool{"vacation": true, "sick": true, "training": true, "other": true}

// AbsenceRequest records an absence; with MovePatients the absent physio's
// booked sessions in the period go to the cover physio where possible.
type AbsenceRequest struct {
	TherapistID  uint            `json:"therapist_id" binding:"required"`
	CoverID      *uint           `json:"cover_id"`
	From         string          `json:"from" binding:"required"`
	To           string          `json:"to" binding:"required"`
	Reason       string          `json:"reason" binding:"required,oneof=vacation sick training other"`
	Handover     string          `json:"handover" binding:"max=10000"`
	Notes        []HandoverInput `json:"notes" binding:"dive"`
	MovePatients bool            `json:"move_patients"`
	DryRun       bool            `json:"dry_run"`
	// Per session: who takes it (therapist 0 = stays with the absent
	// physio). Sessions not listed go to the cover physio.
	Assignments []Assignment `json:"assignments" binding:"max=1000,dive"`
}

// Assignment sends one session of the absent physio to a chosen physio.
type Assignment struct {
	AppointmentID string `json:"appointment_id" binding:"required,uuid"`
	TherapistID   uint   `json:"therapist_id"`
}

// Candidate is a physio who works at a session's hour, with the room they
// have left (after the moves already planned in this request).
type Candidate struct {
	TherapistID uint `json:"therapist_id"`
	Places      int  `json:"places"`  // free patient places that hour
	Credits     int  `json:"credits"` // free care credits that hour
	Fits        bool `json:"fits"`    // this patient fits (places, credits, not busy)
}

// HandoverInput is the note left about one patient.
type HandoverInput struct {
	PatientID string `json:"patient_id" binding:"required,uuid"`
	Note      string `json:"note" binding:"max=2000"`
}

// MovedSession is one session of the absent physio and what happens to it.
type MovedSession struct {
	AppointmentID string `json:"appointment_id"`
	Date          string `json:"date"`
	Hour          int    `json:"hour"`
	PatientID     string `json:"patient_id"`
	PatientName   string `json:"patient_name"`
	Credits       int    `json:"credits"` // the patient's care level
	Moved         bool   `json:"moved"`
	To            uint   `json:"to,omitempty"`     // who takes it
	Reason        string `json:"reason,omitempty"` // not_on_shift | full | over_credits | patient_busy | no_cover | kept
	// Who could take it instead (on shift at that hour), for the coordinator to choose
	Candidates []Candidate `json:"candidates"`
}

// AbsenceResult is the absence and the fate of each session in the period.
type AbsenceResult struct {
	Absence  *Absence       `json:"absence,omitempty"`
	Sessions []MovedSession `json:"sessions"`
	// Patients of the absent physio in the period (for the handover form).
	Patients []PatientSummary `json:"patients"`
}

// CreateAbsence records an absence (coordinators/admins). With MovePatients
// each booked session moves to the cover physio at the same hour when they
// are working then and the hour has room; the others are reported so the
// coordinator can place them by hand. The absent physio's shifts left
// empty in the period are removed. DryRun only previews.
func (s *Service) CreateAbsence(ctx context.Context, a Actor, req AbsenceRequest) (*AbsenceResult, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	from, err := parseDate("from", req.From)
	if err != nil {
		return nil, err
	}
	to, err := parseDate("to", req.To)
	if err != nil {
		return nil, err
	}
	if to.Before(from) || to.Sub(from) > 366*24*time.Hour {
		return nil, &ValidationError{"to", "must be after from and within a year"}
	}
	if !absenceReasons[req.Reason] {
		return nil, &ValidationError{"reason", "is not valid"}
	}
	if ok, err := s.repo.IsTherapist(ctx, cid, req.TherapistID); err != nil || !ok {
		return nil, firstErr(err, &ValidationError{"therapist_id", "is not a physiotherapist of this clinic"})
	}
	if req.CoverID != nil {
		if *req.CoverID == req.TherapistID {
			return nil, &ValidationError{"cover_id", "must be another physiotherapist"}
		}
		if ok, err := s.repo.IsTherapist(ctx, cid, *req.CoverID); err != nil || !ok {
			return nil, firstErr(err, &ValidationError{"cover_id", "is not a physiotherapist of this clinic"})
		}
	}
	existing, err := s.repo.Absences(ctx, cid, from, to)
	if err != nil {
		return nil, err
	}
	for _, ab := range existing {
		if ab.TherapistID == req.TherapistID {
			return nil, &ConflictError{"this physiotherapist already has an absence in this period"}
		}
	}

	data, err := s.load(ctx, cid, from, to)
	if err != nil {
		return nil, err
	}
	res := &AbsenceResult{Sessions: []MovedSession{}, Patients: []PatientSummary{}}
	seen := map[string]bool{}
	// The absent physio's sessions are the ones moving: leave them out of the checks
	plan := planner{data: data, exclude: map[string]bool{}}
	for _, ap := range data.appts {
		if ap.TherapistID == req.TherapistID {
			plan.exclude[ap.ID] = true
		}
	}
	assigned := map[string]uint{}
	for _, as := range req.Assignments {
		if as.TherapistID != 0 {
			if as.TherapistID == req.TherapistID {
				return nil, &ValidationError{"assignments", "sessions cannot go to the absent physiotherapist"}
			}
			if ok, err := s.repo.IsTherapist(ctx, cid, as.TherapistID); err != nil || !ok {
				return nil, firstErr(err, &ValidationError{"assignments", "is not a physiotherapist of this clinic"})
			}
		}
		assigned[as.AppointmentID] = as.TherapistID
	}
	var moves []struct {
		ap  Appointment
		div *Division
	}
	sorted := slices.Clone(data.appts)
	slices.SortStableFunc(sorted, func(x, y Appointment) int {
		if c := x.Date.Compare(y.Date); c != 0 {
			return c
		}
		return x.Hour - y.Hour
	})
	for _, ap := range sorted {
		if ap.TherapistID != req.TherapistID || ap.Status != StatusBooked {
			continue
		}
		name := ""
		if ap.Patient != nil {
			name = ap.Patient.Name
			if !seen[ap.PatientID] {
				seen[ap.PatientID] = true
				res.Patients = append(res.Patients, *ap.Patient)
			}
		}
		ms := MovedSession{AppointmentID: ap.ID, Date: ap.Date.Format(dateLayout), Hour: ap.Hour, PatientID: ap.PatientID,
			PatientName: name, Credits: ap.Credits}
		plan.level = ap.Credits
		// Everyone else working that hour, with their room left so far
		ms.Candidates = s.candidates(&plan, data, ap, req.TherapistID)
		to, chosen := assigned[ap.ID]
		if !chosen && req.CoverID != nil {
			to = *req.CoverID
		}
		switch {
		case !req.MovePatients || (chosen && to == 0):
			ms.Reason = "kept"
		case to == 0:
			ms.Reason = "no_cover"
		default:
			reason, div := plan.problem(ap.Date, ap.Hour, to, ap.PatientID)
			if reason == "" {
				ms.Moved, ms.To = true, to
				moves = append(moves, struct {
					ap  Appointment
					div *Division
				}{ap, div})
				plan.add(ap.Date, ap.Hour, to, ap.PatientID)
			} else {
				ms.Reason = reason
			}
		}
		res.Sessions = append(res.Sessions, ms)
	}
	if req.DryRun {
		return res, nil
	}

	ab := &Absence{ClinicID: cid, TherapistID: req.TherapistID, CoverID: req.CoverID, From: from, To: to,
		Reason: req.Reason, Handover: strings.TrimSpace(req.Handover), CreatedByID: a.ID}
	for _, n := range req.Notes {
		if note := strings.TrimSpace(n.Note); note != "" {
			ab.Notes = append(ab.Notes, HandoverNote{PatientID: n.PatientID, Note: note})
		}
	}
	if err := s.repo.CreateAbsence(ctx, ab); err != nil {
		return nil, err
	}
	for i, ms := range res.Sessions {
		if !ms.Moved {
			continue
		}
		mv := moves[0]
		moves = moves[1:]
		if err := s.repo.MoveAppointment(ctx, &mv.ap, ms.To, mv.div.Capacity, creditLimit(mv.div)); err != nil {
			if res.Sessions[i].Reason = skipReason(err); res.Sessions[i].Reason == "" {
				return nil, err
			}
			res.Sessions[i].Moved = false
		}
	}
	// The absent physio's shifts that no longer hold patients are removed
	if err := s.clearShifts(ctx, cid, req.TherapistID, from, to); err != nil {
		return nil, err
	}
	res.Absence = ab
	return res, nil
}

// candidates are the physios (not the absent one) on shift at the session's
// hour, with the places and credits they have left; Fits says whether this
// patient could go to them.
func (s *Service) candidates(plan *planner, data *boardData, ap Appointment, absent uint) []Candidate {
	out := []Candidate{}
	for _, t := range data.therapists {
		if t.ID == absent {
			continue
		}
		places, credits, ok := plan.room(ap.Date, ap.Hour, t.ID)
		if !ok {
			continue
		}
		reason, _ := plan.problem(ap.Date, ap.Hour, t.ID, ap.PatientID)
		out = append(out, Candidate{TherapistID: t.ID, Places: places, Credits: credits, Fits: reason == ""})
	}
	return out
}

// clearShifts deletes the physio's shifts in the period that hold no patients.
func (s *Service) clearShifts(ctx context.Context, cid string, therapist uint, from, to time.Time) error {
	shifts, err := s.repo.Shifts(ctx, cid, from, to)
	if err != nil {
		return err
	}
	for _, sh := range shifts {
		if sh.TherapistID != therapist {
			continue
		}
		if err := s.checkStranded(ctx, cid, sh, nil); err == nil {
			if err := s.repo.DeleteShift(ctx, sh.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// Absences lists the clinic's absences overlapping from..to, with handovers.
func (s *Service) Absences(ctx context.Context, a Actor, fromS, toS string) ([]Absence, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	from, err := parseDate("from", fromS)
	if err != nil {
		return nil, err
	}
	to, err := parseDate("to", toS)
	if err != nil {
		return nil, err
	}
	return s.repo.Absences(ctx, cid, from, to)
}

// DeleteAbsence removes an absence record (sessions already moved stay with
// the cover; shifts are not re-created).
func (s *Service) DeleteAbsence(ctx context.Context, a Actor, id string) error {
	if !canManage(a) {
		return ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return err
	}
	if uuid.Validate(id) != nil {
		return ErrNotFound
	}
	return s.repo.DeleteAbsence(ctx, cid, id)
}

// Handover returns the handover notes about a patient from absences that
// are current or upcoming (for the patient pop-up and the cover physio).
func (s *Service) Handover(ctx context.Context, a Actor, patientID string) ([]PatientHandover, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if uuid.Validate(patientID) != nil {
		return nil, ErrNotFound
	}
	today := s.today()
	absences, err := s.repo.Absences(ctx, cid, today, today.AddDate(0, 3, 0))
	if err != nil {
		return nil, err
	}
	out := []PatientHandover{}
	for _, ab := range absences {
		for _, n := range ab.Notes {
			if n.PatientID == patientID {
				out = append(out, PatientHandover{Absence: ab, Note: n.Note})
			}
		}
	}
	return out, nil
}

// PatientHandover is one handover note about a patient, with its absence.
type PatientHandover struct {
	Absence Absence `json:"absence"`
	Note    string  `json:"note"`
}

// CoveredBy returns the physios the cover physio is covering on the date
// (used by referrals so the cover can read and write on their threads).
func (s *Service) CoveredBy(ctx context.Context, coverID uint, at time.Time) ([]uint, error) {
	return s.repo.CoveredBy(ctx, coverID, at)
}

func (s *Service) today() time.Time {
	n := s.now()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

// absentOn reports whether the physio is absent on the date.
func absentOn(absences []Absence, therapist uint, d time.Time) bool {
	for _, ab := range absences {
		if ab.TherapistID == therapist && ab.Covers(d) {
			return true
		}
	}
	return false
}

// errAbsent explains a booking refused because the physio is away.
var errAbsent = errors.New("the physiotherapist is absent on this date")
