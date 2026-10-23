package staff

import (
	"context"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmploymentEvent is one change of a person's work status (active, on leave,
// left) from a date, kept so the profile can show the whole story:
// "joined 1 Oct · leave from 10 Oct · left 31 Dec".
type EmploymentEvent struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	Status       string    `json:"status" gorm:"size:10;not null"`
	Since        time.Time `json:"since" gorm:"type:date;not null"`
	Note         string    `json:"note" gorm:"size:300"`
	RecordedByID uint      `json:"recorded_by_id"`
	CreatedAt    time.Time `json:"created_date"`
}

func (EmploymentEvent) TableName() string { return "employment_events" }

// BeforeCreate assigns a UUID.
func (e *EmploymentEvent) BeforeCreate(*gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

// Models lists the module's own tables.
func Models() []any { return []any{&EmploymentEvent{}} }

// History is a person's work story: when they joined and every status
// change since, oldest first.
type History struct {
	JoinedOn time.Time         `json:"joined_on"`
	Events   []EmploymentEvent `json:"events"`
}

// EmploymentHistory returns the story of someone the actor may see (the
// person, people of the same clinic, admins).
func (s *Service) EmploymentHistory(ctx context.Context, a Actor, username string) (*History, error) {
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if ok, err := s.sees(ctx, a, u); err != nil || !ok {
		return nil, firstErr(err, ErrNotFound)
	}
	events, err := s.repo.EmploymentEvents(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if events == nil {
		events = []EmploymentEvent{}
	}
	return &History{JoinedOn: joinedOn(u), Events: events}, nil
}

// joinedOn is the date set by the manager, or when the account was made.
func joinedOn(u *models.User) time.Time {
	if u.JoinedOn != nil {
		return *u.JoinedOn
	}
	return u.CreatedAt
}

// sees: admins, the person and people of the same clinic.
func (s *Service) sees(ctx context.Context, a Actor, target *models.User) (bool, error) {
	if auth.IsAdmin(a.Role) || a.ID == target.ID {
		return true, nil
	}
	me, err := s.repo.UserByID(ctx, a.ID)
	if err != nil {
		return false, err
	}
	return me.HomeFacilityID != nil && target.HomeFacilityID != nil && *me.HomeFacilityID == *target.HomeFacilityID, nil
}
