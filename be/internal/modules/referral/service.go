package referral

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/models"
)

var (
	// ErrForbidden means the actor may not perform the action on this referral.
	ErrForbidden = errors.New("not allowed to perform this action")
	// ErrInvalidTransition means the action is not valid from the current status.
	ErrInvalidTransition = errors.New("action not allowed in the current status")
)

// ValidationError reports bad input that passed binding but fails business rules.
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

// Actor is the authenticated user performing an action.
type Actor struct {
	ID   uint
	Role models.UserRole
}

func (a Actor) isAdmin() bool     { return auth.IsAdmin(a.Role) }
func (a Actor) isDoctor() bool    { return a.Role == models.RoleDoctor }
func (a Actor) isTherapist() bool { return a.Role == models.RolePhysioTherapist }

// Pagination bounds for List.
const (
	defaultLimit = 20
	maxLimit     = 100
)

// Service holds the referral business rules.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService builds a Service. now may be nil to use time.Now.
func NewService(repo Repository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{repo: repo, now: now}
}

// Create opens a new referral. Only doctors and admins can refer patients.
func (s *Service) Create(ctx context.Context, actor Actor, req CreateRequest) (*Referral, error) {
	if !actor.isDoctor() && !actor.isAdmin() {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(req.Diagnosis) == "" || strings.TrimSpace(req.Reason) == "" {
		return nil, &ValidationError{"diagnosis and reason are required"}
	}

	ok, err := s.repo.PatientExists(ctx, req.PatientID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &ValidationError{"patient does not exist"}
	}

	// A doctor may address the referral to a specific therapist; otherwise
	// it lands in every therapist's inbox until someone accepts it.
	if req.TherapistID != nil {
		role, active, err := s.repo.UserRole(ctx, *req.TherapistID)
		if err != nil {
			return nil, err
		}
		if !active || models.UserRole(role) != models.RolePhysioTherapist {
			return nil, &ValidationError{"therapist_id must be an active physiotherapist"}
		}
	}

	priority := req.Priority
	if priority == "" {
		priority = PriorityRoutine
	}

	r := &Referral{
		PatientID:           req.PatientID,
		FacilityID:          req.FacilityID,
		DoctorID:            actor.ID,
		TherapistID:         req.TherapistID,
		Diagnosis:           strings.TrimSpace(req.Diagnosis),
		ICDCode:             req.ICDCode,
		Reason:              strings.TrimSpace(req.Reason),
		Precautions:         req.Precautions,
		Goals:               req.Goals,
		RecommendedSessions: req.RecommendedSessions,
		Priority:            priority,
		Status:              StatusPending,
	}
	if err := s.repo.Create(ctx, r); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, r.ID)
}

// Get returns a referral the actor is allowed to see.
func (s *Service) Get(ctx context.Context, actor Actor, id string) (*Referral, error) {
	r, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canView(actor, r) {
		// Hide existence from users who are not involved.
		return nil, ErrNotFound
	}
	return r, nil
}

// List returns the referrals visible to the actor, scoped by role:
// admins see everything, doctors their own referrals, therapists their inbox.
func (s *Service) List(ctx context.Context, actor Actor, f ListFilter) ([]Referral, int64, error) {
	switch {
	case actor.isAdmin():
		// no scoping
	case actor.isDoctor():
		f.DoctorID = actor.ID
	case actor.isTherapist():
		f.TherapistID = actor.ID
		f.InboxForTherapist = true
	default:
		return nil, 0, ErrForbidden
	}
	if f.Limit <= 0 {
		f.Limit = defaultLimit
	}
	if f.Limit > maxLimit {
		f.Limit = maxLimit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.repo.List(ctx, f)
}

// Accept lets a physiotherapist take a pending referral. Unassigned referrals
// are claimed by whoever accepts first.
func (s *Service) Accept(ctx context.Context, actor Actor, id string) (*Referral, error) {
	return s.transition(ctx, actor, id, StatusPending, func(r *Referral) error {
		if !actor.isTherapist() || (r.TherapistID != nil && *r.TherapistID != actor.ID) {
			return ErrForbidden
		}
		now := s.now()
		r.TherapistID = &actor.ID
		r.Status = StatusAccepted
		r.AcceptedAt = &now
		return nil
	}, "Referral accepted")
}

// Decline lets the assigned therapist refuse a pending referral.
func (s *Service) Decline(ctx context.Context, actor Actor, id, reason string) (*Referral, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, &ValidationError{"a reason is required to decline"}
	}
	return s.transition(ctx, actor, id, StatusPending, func(r *Referral) error {
		if !isAssignedTherapist(actor, r) {
			return ErrForbidden
		}
		r.Status = StatusDeclined
		r.DeclineReason = &reason
		return nil
	}, "Referral declined: "+reason)
}

// Start marks the first treatment session.
func (s *Service) Start(ctx context.Context, actor Actor, id string) (*Referral, error) {
	return s.transition(ctx, actor, id, StatusAccepted, func(r *Referral) error {
		if !isAssignedTherapist(actor, r) {
			return ErrForbidden
		}
		now := s.now()
		r.Status = StatusInProgress
		r.StartedAt = &now
		return nil
	}, "Treatment started")
}

// Complete closes the episode with an outcome summary for the doctor.
func (s *Service) Complete(ctx context.Context, actor Actor, id, summary string) (*Referral, error) {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return nil, &ValidationError{"an outcome summary is required to complete"}
	}
	return s.transition(ctx, actor, id, StatusInProgress, func(r *Referral) error {
		if !isAssignedTherapist(actor, r) {
			return ErrForbidden
		}
		now := s.now()
		r.Status = StatusCompleted
		r.OutcomeSummary = &summary
		r.CompletedAt = &now
		return nil
	}, "Treatment completed: "+summary)
}

// Cancel lets the referring doctor (or an admin) withdraw an open referral.
func (s *Service) Cancel(ctx context.Context, actor Actor, id, reason string) (*Referral, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, &ValidationError{"a reason is required to cancel"}
	}
	r, err := s.Get(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	// Cancel is valid from any open status, so pass the current one as "from".
	return s.transition(ctx, actor, id, r.Status, func(r *Referral) error {
		if !(actor.isAdmin() || (actor.isDoctor() && r.DoctorID == actor.ID)) {
			return ErrForbidden
		}
		if r.Status.IsTerminal() {
			return ErrInvalidTransition
		}
		r.Status = StatusCancelled
		r.CancelReason = &reason
		return nil
	}, "Referral cancelled: "+reason)
}

// AddNote posts a message on the referral thread. Only the referring doctor,
// the assigned therapist and admins take part in the conversation.
func (s *Service) AddNote(ctx context.Context, actor Actor, id string, req NoteRequest) (*Note, error) {
	body := strings.TrimSpace(req.Body)
	if body == "" {
		return nil, &ValidationError{"note body is required"}
	}
	r, err := s.Get(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if !isParticipant(actor, r) {
		return nil, ErrForbidden
	}
	kind := req.Kind
	if kind == "" {
		kind = NoteProgress
	}
	n := &Note{ReferralID: r.ID, AuthorID: actor.ID, Kind: kind, Body: body}
	if err := s.repo.AddNote(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

// Notes returns the thread of a referral the actor can see.
func (s *Service) Notes(ctx context.Context, actor Actor, id string) ([]Note, error) {
	if _, err := s.Get(ctx, actor, id); err != nil {
		return nil, err
	}
	return s.repo.ListNotes(ctx, id)
}

// transition loads the referral, checks it is in status from, applies mutate
// (which enforces permissions and sets fields), and persists atomically with a
// system note so the thread records every status change.
func (s *Service) transition(ctx context.Context, actor Actor, id string, from Status,
	mutate func(*Referral) error, noteText string) (*Referral, error) {
	r, err := s.Get(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if r.Status != from {
		return nil, fmt.Errorf("%w: referral is %s", ErrInvalidTransition, r.Status)
	}
	if err := mutate(r); err != nil {
		return nil, err
	}
	r.UpdatedAt = s.now()
	note := &Note{ReferralID: r.ID, AuthorID: actor.ID, Kind: NoteSystem, Body: noteText}
	if err := s.repo.Transition(ctx, r, from, note); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, r.ID)
}

// canView: participants, admins, and any therapist for unassigned open referrals.
func canView(a Actor, r *Referral) bool {
	if isParticipant(a, r) {
		return true
	}
	return a.isTherapist() && r.TherapistID == nil && r.Status == StatusPending
}

func isParticipant(a Actor, r *Referral) bool {
	return a.isAdmin() || r.DoctorID == a.ID || isAssignedTherapist(a, r)
}

func isAssignedTherapist(a Actor, r *Referral) bool {
	return a.isTherapist() && r.TherapistID != nil && *r.TherapistID == a.ID
}
