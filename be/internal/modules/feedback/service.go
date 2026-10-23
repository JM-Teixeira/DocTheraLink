// Package feedback collects suggestions about the app from anyone signed in
// and gives admins one place to read and triage them.
//
//	service.go    - model and rules
//	repository.go - persistence
//	handler.go    - HTTP
package feedback

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"klyniva/internal/auth"
	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrForbidden = errors.New("not allowed to perform this action")
	ErrNotFound  = errors.New("suggestion not found")
)

// ValidationError reports bad input with the offending field.
type ValidationError struct{ Field, Msg string }

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

// Kinds and statuses (labels live in the frontend).
var (
	kinds    = map[string]bool{"idea": true, "problem": true, "other": true}
	statuses = map[string]bool{"new": true, "read": true, "planned": true, "done": true, "declined": true}
)

const maxText = 2000

// Suggestion is one idea or problem someone sent.
type Suggestion struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	AuthorID  uint      `json:"author_id" gorm:"not null;index"`
	Author    *author   `json:"author,omitempty" gorm:"foreignKey:AuthorID;->;-:migration"`
	ClinicID  *string   `json:"clinic_id" gorm:"type:uuid"`
	Kind      string    `json:"kind" gorm:"size:10;not null"`
	Page      string    `json:"page" gorm:"size:100"` // where in the app it was about
	Text      string    `json:"text" gorm:"type:text;not null"`
	Status    string    `json:"status" gorm:"size:10;not null;default:new;index"`
	AdminNote string    `json:"admin_note" gorm:"size:1000"` // the answer the author sees
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

func (Suggestion) TableName() string { return "suggestions" }

// BeforeCreate assigns a UUID.
func (s *Suggestion) BeforeCreate(*gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	return nil
}

// author is who sent it (read-only view of users).
type author struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserType  string `json:"user_type"`
}

func (author) TableName() string { return "users" }

// Models lists the module's tables.
func Models() []any { return []any{&Suggestion{}} }

// Actor is the signed-in user.
type Actor struct {
	ID   uint
	Role models.UserRole
}

// SendRequest is a new suggestion.
type SendRequest struct {
	Kind string `json:"kind"`
	Page string `json:"page"`
	Text string `json:"text"`
}

// ReviewRequest is an admin's triage.
type ReviewRequest struct {
	Status    string  `json:"status"`
	AdminNote *string `json:"admin_note"`
}

// Repository is the persistence boundary.
type Repository interface {
	Add(ctx context.Context, s *Suggestion) error
	List(ctx context.Context, authorID uint, status string) ([]Suggestion, error) // authorID 0 = everyone
	Get(ctx context.Context, id string) (*Suggestion, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	CountNew(ctx context.Context) (int64, error)
}

// Service holds the rules.
type Service struct {
	repo    Repository
	clinics clinic.Resolver
}

// NewService builds a Service.
func NewService(repo Repository, clinics clinic.Resolver) *Service {
	return &Service{repo: repo, clinics: clinics}
}

// Send stores a suggestion from anyone signed in.
func (s *Service) Send(ctx context.Context, a Actor, req SendRequest) (*Suggestion, error) {
	text := strings.TrimSpace(req.Text)
	if utf8.RuneCountInString(text) < 5 {
		return nil, &ValidationError{"text", "write at least a few words"}
	}
	if utf8.RuneCountInString(text) > maxText {
		return nil, &ValidationError{"text", "at most 2000 characters"}
	}
	if !kinds[req.Kind] {
		return nil, &ValidationError{"kind", "must be idea, problem or other"}
	}
	home, err := s.clinics.HomeClinic(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	page := strings.TrimSpace(req.Page)
	if len(page) > 100 {
		page = page[:100]
	}
	sg := &Suggestion{AuthorID: a.ID, ClinicID: home, Kind: req.Kind, Page: page, Text: text, Status: "new"}
	return sg, s.repo.Add(ctx, sg)
}

// Mine lists the actor's own suggestions with the admins' answers.
func (s *Service) Mine(ctx context.Context, a Actor) ([]Suggestion, error) {
	return nonNil(s.repo.List(ctx, a.ID, ""))
}

// All lists every suggestion (admins), optionally by status.
func (s *Service) All(ctx context.Context, a Actor, status string) ([]Suggestion, error) {
	if !auth.IsAdmin(a.Role) {
		return nil, ErrForbidden
	}
	if status != "" && !statuses[status] {
		return nil, &ValidationError{"status", "unknown status"}
	}
	return nonNil(s.repo.List(ctx, 0, status))
}

// CountNew is the number of suggestions nobody has read yet (admins).
func (s *Service) CountNew(ctx context.Context, a Actor) (int64, error) {
	if !auth.IsAdmin(a.Role) {
		return 0, ErrForbidden
	}
	return s.repo.CountNew(ctx)
}

// Review sets the status and/or the answer (admins).
func (s *Service) Review(ctx context.Context, a Actor, id string, req ReviewRequest) (*Suggestion, error) {
	if !auth.IsAdmin(a.Role) {
		return nil, ErrForbidden
	}
	if uuid.Validate(id) != nil {
		return nil, ErrNotFound
	}
	if _, err := s.repo.Get(ctx, id); err != nil {
		return nil, err
	}
	fields := map[string]any{}
	if req.Status != "" {
		if !statuses[req.Status] {
			return nil, &ValidationError{"status", "unknown status"}
		}
		fields["status"] = req.Status
	}
	if req.AdminNote != nil {
		note := strings.TrimSpace(*req.AdminNote)
		if utf8.RuneCountInString(note) > 1000 {
			return nil, &ValidationError{"admin_note", "at most 1000 characters"}
		}
		fields["admin_note"] = note
	}
	if len(fields) > 0 {
		if err := s.repo.Update(ctx, id, fields); err != nil {
			return nil, err
		}
	}
	return s.repo.Get(ctx, id)
}

func nonNil(items []Suggestion, err error) ([]Suggestion, error) {
	if items == nil {
		items = []Suggestion{}
	}
	return items, err
}
