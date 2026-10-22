package referral

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Treatment plan of a referral: the steps of a session in order ("10 min
// ultrasound on the shoulder", "Codman exercises 3x10"), so a colleague
// covering the therapist sees exactly what was being done. Every save is a
// new version; the newest is the current plan.

// PlanItem is one step of a session.
type PlanItem struct {
	Activity  string `json:"activity"`            // what (required)
	Minutes   *int   `json:"minutes,omitempty"`   // how long
	Dose      string `json:"dose,omitempty"`      // sets x repetitions, intensity, ...
	Equipment string `json:"equipment,omitempty"` // machine or room
	Notes     string `json:"notes,omitempty"`
}

// Plan is one saved version of a referral's plan.
type Plan struct {
	ID         string       `json:"id" gorm:"primaryKey;type:uuid"`
	ReferralID string       `json:"referral_id" gorm:"type:uuid;not null;index"`
	Items      []PlanItem   `json:"items" gorm:"serializer:json;type:text;not null"`
	AuthorID   uint         `json:"author_id" gorm:"not null"`
	Author     *UserSummary `json:"author,omitempty" gorm:"foreignKey:AuthorID;->;-:migration"`
	CreatedAt  time.Time    `json:"created_date"`
}

func (Plan) TableName() string { return "referral_plans" }

// BeforeCreate assigns a UUID.
func (p *Plan) BeforeCreate(*gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}

// PlanView is the current plan (nil when none yet), every earlier version
// (newest first) and whether the viewer may change it.
type PlanView struct {
	Current  *Plan  `json:"current"`
	Versions []Plan `json:"versions"`
	CanEdit  bool   `json:"can_edit"`
}

// PlanRequest replaces the plan with these steps.
type PlanRequest struct {
	Items []PlanItem `json:"items" binding:"required"`
}

// Limits of a plan.
const (
	maxPlanItems   = 30
	maxPlanMinutes = 240
)

// Plan returns a visible referral's plan.
func (s *Service) Plan(ctx context.Context, actor Actor, id string) (*PlanView, error) {
	r, err := s.Get(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	versions, err := s.repo.Plans(ctx, id)
	if err != nil {
		return nil, err
	}
	canEdit, err := s.canEditPlan(ctx, actor, r)
	if err != nil {
		return nil, err
	}
	v := &PlanView{Versions: versions, CanEdit: canEdit}
	if v.Versions == nil {
		v.Versions = []Plan{}
	}
	if len(versions) > 0 {
		v.Current = &versions[0]
	}
	return v, nil
}

// SavePlan stores a new version of the plan.
func (s *Service) SavePlan(ctx context.Context, actor Actor, id string, req PlanRequest) (*PlanView, error) {
	r, err := s.Get(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	ok, err := s.canEditPlan(ctx, actor, r)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	items, err := cleanPlan(req.Items)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AddPlan(ctx, &Plan{ReferralID: id, Items: items, AuthorID: actor.ID}); err != nil {
		return nil, err
	}
	return s.Plan(ctx, actor, id)
}

// canEditPlan: while the referral is open, its physio (or whoever covers
// them), the clinic's coordinators and admins.
func (s *Service) canEditPlan(ctx context.Context, a Actor, r *Referral) (bool, error) {
	if r.Status != StatusAccepted && r.Status != StatusInProgress {
		return false, nil
	}
	switch {
	case a.isAdmin():
		return true, nil
	case a.isCoordinator():
		mine, err := s.clinics.HomeClinic(ctx, a.ID)
		return err == nil && mine != nil && r.FacilityID != nil && *mine == *r.FacilityID, err
	case a.isTherapist():
		return s.participates(ctx, a, r)
	}
	return false, nil
}

// cleanPlan trims the steps, drops empty ones and checks the limits.
func cleanPlan(in []PlanItem) ([]PlanItem, error) {
	out := []PlanItem{}
	for _, it := range in {
		it.Activity = strings.TrimSpace(it.Activity)
		it.Dose = strings.TrimSpace(it.Dose)
		it.Equipment = strings.TrimSpace(it.Equipment)
		it.Notes = strings.TrimSpace(it.Notes)
		if it.Activity == "" {
			continue
		}
		if it.Minutes != nil && (*it.Minutes < 1 || *it.Minutes > maxPlanMinutes) {
			return nil, &ValidationError{fmt.Sprintf("minutes must be between 1 and %d", maxPlanMinutes)}
		}
		for _, f := range []struct {
			s   string
			max int
		}{{it.Activity, 120}, {it.Dose, 40}, {it.Equipment, 80}, {it.Notes, 300}} {
			if utf8.RuneCountInString(f.s) > f.max {
				return nil, &ValidationError{fmt.Sprintf("a text is longer than %d characters", f.max)}
			}
		}
		out = append(out, it)
	}
	if len(out) > maxPlanItems {
		return nil, &ValidationError{fmt.Sprintf("at most %d steps", maxPlanItems)}
	}
	return out, nil
}
