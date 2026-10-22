package schedule

import (
	"context"
	"strings"
	"time"

	"klyniva/internal/models"

	"github.com/google/uuid"
)

// Waiting-list statuses.
const (
	WaitingOpen      = "waiting"
	WaitingScheduled = "scheduled" // a plan was booked
	WaitingRemoved   = "removed"   // no longer needed (kept as history)
)

// WaitingEntry is a patient waiting for a slot: urgent ones first, then by
// how long they have waited, with their preferred days and time of day.
type WaitingEntry struct {
	ID                   string          `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ClinicID             string          `json:"clinic_id" gorm:"type:uuid;not null;index"`
	PatientID            string          `json:"patient_id" gorm:"type:uuid;not null;index"`
	Patient              *PatientSummary `json:"patient,omitempty" gorm:"foreignKey:PatientID;->;-:migration"`
	ReferralID           *string         `json:"referral_id" gorm:"type:uuid"`
	Urgent               bool            `json:"urgent" gorm:"not null;default:false"`
	Weekdays             []int           `json:"weekdays" gorm:"serializer:json;type:text"`  // preferred (0 = Sunday)
	Period               string          `json:"period" gorm:"size:10;not null;default:any"` // morning | afternoon | any
	PreferredTherapistID *uint           `json:"preferred_therapist_id"`
	Note                 string          `json:"note" gorm:"size:300"`
	Status               string          `json:"status" gorm:"size:12;not null;default:waiting;index"`
	SeriesID             *string         `json:"series_id" gorm:"type:uuid"` // the plan booked for it
	CreatedByID          uint            `json:"created_by_id"`
	CreatedAt            time.Time       `json:"created_date"`
	ClosedAt             *time.Time      `json:"closed_at"`
}

func (WaitingEntry) TableName() string { return "waiting_list" }

// WaitingRequest adds a patient to the waiting list.
type WaitingRequest struct {
	PatientID            string  `json:"patient_id" binding:"required,uuid"`
	ReferralID           *string `json:"referral_id" binding:"omitempty,uuid"`
	Urgent               bool    `json:"urgent"`
	Weekdays             []int   `json:"weekdays" binding:"max=7,dive,min=0,max=6"`
	Period               string  `json:"period" binding:"omitempty,oneof=morning afternoon any"`
	PreferredTherapistID *uint   `json:"preferred_therapist_id"`
	Note                 string  `json:"note" binding:"max=300"`
}

// WaitingUpdate closes an entry (scheduled / removed) or edits it.
type WaitingUpdate struct {
	Status   string  `json:"status" binding:"omitempty,oneof=waiting scheduled removed"`
	SeriesID *string `json:"series_id" binding:"omitempty,uuid"`
	Urgent   *bool   `json:"urgent"`
	Note     *string `json:"note" binding:"omitempty,max=300"`
}

// canQueue: who manages the waiting list - coordinators, admins and the front desk.
func canQueue(a Actor) bool { return canManage(a) || a.Role == models.RoleReceptionist }

// WaitingList lists the clinic's entries with a status (default: waiting),
// urgent first and then oldest first.
func (s *Service) WaitingList(ctx context.Context, a Actor, status string) ([]WaitingEntry, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if status == "" {
		status = WaitingOpen
	}
	return s.repo.Waiting(ctx, cid, status)
}

// AddWaiting puts a patient of the clinic on the list (once while waiting).
func (s *Service) AddWaiting(ctx context.Context, a Actor, req WaitingRequest) (*WaitingEntry, error) {
	if !canQueue(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if _, ok, err := s.repo.PatientCareLevel(ctx, cid, req.PatientID); err != nil || !ok {
		return nil, firstErr(err, &ValidationError{"patient_id", "is not a patient of this clinic"})
	}
	if req.PreferredTherapistID != nil {
		if ok, err := s.repo.IsTherapist(ctx, cid, *req.PreferredTherapistID); err != nil || !ok {
			return nil, firstErr(err, &ValidationError{"preferred_therapist_id", "is not a physiotherapist of this clinic"})
		}
	}
	open, err := s.repo.Waiting(ctx, cid, WaitingOpen)
	if err != nil {
		return nil, err
	}
	for _, e := range open {
		if e.PatientID == req.PatientID {
			return nil, &ConflictError{"this patient is already on the waiting list"}
		}
	}
	period := req.Period
	if period == "" {
		period = "any"
	}
	e := &WaitingEntry{ClinicID: cid, PatientID: req.PatientID, ReferralID: req.ReferralID, Urgent: req.Urgent,
		Weekdays: req.Weekdays, Period: period, PreferredTherapistID: req.PreferredTherapistID,
		Note: strings.TrimSpace(req.Note), Status: WaitingOpen, CreatedByID: a.ID}
	if e.Weekdays == nil {
		e.Weekdays = []int{}
	}
	if err := s.repo.SaveWaiting(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// UpdateWaiting marks an entry scheduled (with its plan) or removed, or
// edits urgency and note. Closed entries are kept as history.
func (s *Service) UpdateWaiting(ctx context.Context, a Actor, id string, req WaitingUpdate) (*WaitingEntry, error) {
	if !canQueue(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if uuid.Validate(id) != nil {
		return nil, ErrNotFound
	}
	e, err := s.repo.GetWaiting(ctx, cid, id)
	if err != nil {
		return nil, err
	}
	fields := map[string]any{}
	if req.Status != "" && req.Status != e.Status {
		fields["status"] = req.Status
		if req.Status == WaitingOpen {
			fields["closed_at"] = nil
		} else {
			fields["closed_at"] = s.now()
		}
	}
	if req.SeriesID != nil {
		fields["series_id"] = *req.SeriesID
	}
	if req.Urgent != nil {
		fields["urgent"] = *req.Urgent
	}
	if req.Note != nil {
		fields["note"] = strings.TrimSpace(*req.Note)
	}
	if len(fields) > 0 {
		if err := s.repo.UpdateWaiting(ctx, id, fields); err != nil {
			return nil, err
		}
	}
	return s.repo.GetWaiting(ctx, cid, id)
}
