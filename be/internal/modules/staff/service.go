// Package staff manages the people of a clinic beyond their account: the
// skills they have, the clinic they work in and their employment status
// (active, on leave, left — people who left are kept as a record but can no
// longer sign in).
//
//	service.go    - permissions and rules (this file)
//	repository.go - persistence
//	handler.go    - HTTP handlers, on the paths the user pages already use
//	seed.go       - the physiotherapy skills catalogue
package staff

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/models"

	"github.com/google/uuid"
)

// Employment statuses.
const (
	StatusActive  = "active"
	StatusOnLeave = "on_leave"
	StatusLeft    = "left"
)

var (
	// ErrForbidden: the actor may not change this user.
	ErrForbidden = errors.New("not allowed to perform this action")
	// ErrNotFound: no such user, skill or clinic.
	ErrNotFound = errors.New("not found")
)

// ValidationError reports bad input with the offending field.
type ValidationError struct{ Field, Msg string }

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

// Actor is the authenticated user.
type Actor struct {
	ID   uint
	Role models.UserRole
}

// EmploymentRequest changes a user's employment status.
type EmploymentRequest struct {
	Status string `json:"status" binding:"required,oneof=active on_leave left"`
	Since  string `json:"since"` // YYYY-MM-DD, today by default
	Note   string `json:"note" binding:"max=300"`
	// JoinedOn (YYYY-MM-DD) corrects when the person started, optional
	JoinedOn string `json:"joined_on"`
}

// PreferencesRequest changes how the app is shown to the person.
// PreferencesRequest changes the settings sent; empty ones stay as they are.
type PreferencesRequest struct {
	DateFormat string `json:"date_format"`
	Theme      string `json:"theme"`
	TextSize   string `json:"text_size"`
	StartPage  string `json:"start_page"`
	Language   string `json:"language"`
}

// preferenceValues are the accepted values of each setting.
var preferenceValues = map[string][]string{
	"date_format": {"dmy", "mdy", "ymd", "long"},
	"theme":       {"light", "dark", "system"},
	"text_size":   {"normal", "large"},
	"start_page":  {"home", "patients", "referrals", "time-board", "inventory"},
	"language":    {"pt", "en"},
}

// Service holds the staff rules.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService builds the service.
func NewService(repo Repository) *Service { return &Service{repo: repo, now: time.Now} }

// manages reports whether the actor manages the target: admins everyone,
// coordinators the people of their own clinic.
func (s *Service) manages(ctx context.Context, a Actor, target *models.User) (bool, error) {
	if auth.IsAdmin(a.Role) {
		return true, nil
	}
	if a.Role != models.RoleCoordinator || target.HomeFacilityID == nil {
		return false, nil
	}
	me, err := s.repo.UserByID(ctx, a.ID)
	if err != nil {
		return false, err
	}
	return me.HomeFacilityID != nil && *me.HomeFacilityID == *target.HomeFacilityID, nil
}

// canEditSkills: the person themself or whoever manages them.
func (s *Service) canEditSkills(ctx context.Context, a Actor, target *models.User) (bool, error) {
	if target.ID == a.ID {
		return true, nil
	}
	return s.manages(ctx, a, target)
}

// canView: admins, the person themself and the people of the same clinic.
func (s *Service) canView(ctx context.Context, a Actor, target *models.User) (bool, error) {
	if auth.IsAdmin(a.Role) || target.ID == a.ID {
		return true, nil
	}
	if target.HomeFacilityID == nil {
		return false, nil
	}
	me, err := s.repo.UserByID(ctx, a.ID)
	if err != nil {
		return false, err
	}
	return me.HomeFacilityID != nil && *me.HomeFacilityID == *target.HomeFacilityID, nil
}

// visibleUser loads a user the actor may see; others look like they don't
// exist (no hint that the username is taken).
func (s *Service) visibleUser(ctx context.Context, a Actor, username string) (*models.User, error) {
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if ok, err := s.canView(ctx, a, u); err != nil || !ok {
		return nil, firstErr(err, ErrNotFound)
	}
	return u, nil
}

// Skills lists the catalogue.
func (s *Service) Skills(ctx context.Context) ([]models.Skill, error) { return s.repo.Skills(ctx) }

// UserSkills lists a user's skills (to people who may see that user).
func (s *Service) UserSkills(ctx context.Context, a Actor, username string) ([]models.UserSkill, error) {
	u, err := s.visibleUser(ctx, a, username)
	if err != nil {
		return nil, err
	}
	return s.repo.UserSkills(ctx, u.ID)
}

// AddSkill links a catalogue skill to a user (once).
func (s *Service) AddSkill(ctx context.Context, a Actor, username, skillID string) (*models.UserSkill, error) {
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if ok, err := s.canEditSkills(ctx, a, u); err != nil || !ok {
		return nil, firstErr(err, ErrForbidden)
	}
	if uuid.Validate(skillID) != nil {
		return nil, &ValidationError{"skill", "is not a valid skill"}
	}
	if _, err := s.repo.Skill(ctx, skillID); err != nil {
		return nil, &ValidationError{"skill", "is not a valid skill"}
	}
	existing, err := s.repo.UserSkills(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	for _, us := range existing {
		if us.SkillID == skillID {
			return &us, nil
		}
	}
	return s.repo.AddUserSkill(ctx, u.ID, skillID)
}

// RemoveSkill unlinks a skill (by link id or skill id).
func (s *Service) RemoveSkill(ctx context.Context, a Actor, username, id string) error {
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if ok, err := s.canEditSkills(ctx, a, u); err != nil || !ok {
		return firstErr(err, ErrForbidden)
	}
	return s.repo.RemoveUserSkill(ctx, u.ID, id)
}

// Clinics lists the clinic a user works in (one today, see decision D2).
func (s *Service) Clinics(ctx context.Context, a Actor, username string) ([]models.Facility, error) {
	u, err := s.visibleUser(ctx, a, username)
	if err != nil {
		return nil, err
	}
	if u.HomeFacilityID == nil {
		return []models.Facility{}, nil
	}
	f, err := s.repo.Clinic(ctx, *u.HomeFacilityID)
	if errors.Is(err, ErrNotFound) {
		return []models.Facility{}, nil
	}
	if err != nil {
		return nil, err
	}
	return []models.Facility{*f}, nil
}

// SetClinic moves a user to a clinic (admins only); "" clears it.
func (s *Service) SetClinic(ctx context.Context, a Actor, username, clinicID string) (*models.User, error) {
	if !auth.IsAdmin(a.Role) {
		return nil, ErrForbidden
	}
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	var value *string
	if clinicID != "" {
		if uuid.Validate(clinicID) != nil {
			return nil, &ValidationError{"facility", "is not a valid clinic"}
		}
		if _, err := s.repo.Clinic(ctx, clinicID); err != nil {
			return nil, &ValidationError{"facility", "is not a valid clinic"}
		}
		value = &clinicID
	}
	if err := s.repo.UpdateUser(ctx, u.ID, map[string]any{"home_facility_id": value}); err != nil {
		return nil, err
	}
	return s.repo.UserByID(ctx, u.ID)
}

// SetEmployment records that a user is active, on leave or has left.
// People who left keep their record but can no longer sign in.
func (s *Service) SetEmployment(ctx context.Context, a Actor, username string, req EmploymentRequest) (*models.User, error) {
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u.ID == a.ID {
		return nil, &ValidationError{"status", "you cannot change your own status"}
	}
	if ok, err := s.manages(ctx, a, u); err != nil || !ok {
		return nil, firstErr(err, ErrForbidden)
	}
	since := s.now()
	if req.Since != "" {
		if since, err = time.Parse("2006-01-02", req.Since); err != nil {
			return nil, &ValidationError{"since", "must be a date in YYYY-MM-DD format"}
		}
	}
	fields := map[string]any{
		"employment_status": req.Status, "employment_since": since,
		"employment_note": strings.TrimSpace(req.Note), "is_active": req.Status != StatusLeft,
	}
	if req.JoinedOn != "" {
		joined, err := time.Parse("2006-01-02", req.JoinedOn)
		if err != nil {
			return nil, &ValidationError{"joined_on", "must be a date in YYYY-MM-DD format"}
		}
		fields["joined_on"] = joined
	}
	if err := s.repo.UpdateUser(ctx, u.ID, fields); err != nil {
		return nil, err
	}
	// The history keeps every change (the user row only has the latest)
	if err := s.repo.AddEmploymentEvent(ctx, &EmploymentEvent{UserID: u.ID, Status: req.Status, Since: since,
		Note: strings.TrimSpace(req.Note), RecordedByID: a.ID}); err != nil {
		return nil, err
	}
	auth.ForgetAccount(u.ID) // a person who left loses access now, not in 30 s
	return s.repo.UserByID(ctx, u.ID)
}

// SetPreferences saves a person's own display preferences (nobody else's).
func (s *Service) SetPreferences(ctx context.Context, a Actor, username string, req PreferencesRequest) (*models.User, error) {
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u.ID != a.ID {
		return nil, ErrForbidden
	}
	fields := map[string]any{}
	for col, v := range map[string]string{
		"date_format": req.DateFormat, "theme": req.Theme, "text_size": req.TextSize,
		"start_page": req.StartPage, "language": req.Language,
	} {
		if v == "" {
			continue
		}
		if !slices.Contains(preferenceValues[col], v) {
			return nil, &ValidationError{col, "must be one of " + strings.Join(preferenceValues[col], ", ")}
		}
		fields[col] = v
	}
	if len(fields) > 0 {
		if err := s.repo.UpdateUser(ctx, u.ID, fields); err != nil {
			return nil, err
		}
	}
	return s.repo.UserByID(ctx, u.ID)
}

func firstErr(err, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
