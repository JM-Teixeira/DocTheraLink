package inventory

import (
	"context"
	"errors"
	"strings"

	"klyniva/internal/auth"
	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/google/uuid"
)

var (
	// ErrForbidden: the role may not do this.
	ErrForbidden = errors.New("not allowed to perform this action")
	// ErrNoClinic: the user works in no clinic, so has no stock.
	ErrNoClinic = errors.New("you are not assigned to a clinic")
)

// ValidationError reports bad input with the offending field.
type ValidationError struct{ Field, Msg string }

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

// Actor is the authenticated user.
type Actor struct {
	ID   uint
	Role models.UserRole
}

// Service holds the inventory rules:
//   - everyone sees the stock of their own clinic (admins: every clinic);
//   - any non-read-only member records use, restocks, counts and new items;
//   - only coordinators and admins remove items from the list.
type Service struct {
	repo    Repository
	clinics clinic.Resolver
}

// NewService builds the service.
func NewService(repo Repository, clinics clinic.Resolver) *Service {
	return &Service{repo: repo, clinics: clinics}
}

// clinicsFor returns the clinics the actor works with; nil = all (admins).
func (s *Service) clinicsFor(ctx context.Context, a Actor) ([]string, error) {
	if auth.IsAdmin(a.Role) {
		return nil, nil
	}
	home, err := s.clinics.HomeClinic(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	if home == nil {
		return nil, ErrNoClinic
	}
	return []string{*home}, nil
}

func canWrite(a Actor) bool { return auth.CanModifyResource(a.Role) }

// canAudit: the clinic-wide log of who used or opened what.
func canAudit(a Actor) bool { return canArchive(a) }

func canArchive(a Actor) bool {
	return auth.IsAdmin(a.Role) || a.Role == models.RoleCoordinator
}

// List returns the clinic's items, by category and name.
func (s *Service) List(ctx context.Context, a Actor, f ListFilter) ([]ItemView, error) {
	ids, err := s.clinicsFor(ctx, a)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx, ids, f)
	if err != nil {
		return nil, err
	}
	out := make([]ItemView, len(items))
	for i, it := range items {
		out[i] = view(it)
	}
	return out, nil
}

// Summary counts all items and those needing attention.
func (s *Service) Summary(ctx context.Context, a Actor) (Summary, error) {
	items, err := s.List(ctx, a, ListFilter{})
	if err != nil {
		return Summary{}, err
	}
	sum := Summary{Items: len(items)}
	for _, it := range items {
		switch it.Status {
		case "low":
			sum.Low++
		case "out":
			sum.Out++
		}
	}
	return sum, nil
}

// Get returns one visible item.
func (s *Service) Get(ctx context.Context, a Actor, id string) (*Item, error) {
	if uuid.Validate(id) != nil {
		return nil, ErrNotFound
	}
	ids, err := s.clinicsFor(ctx, a)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, ids, id)
}

// Create adds an item to the actor's clinic with its opening stock.
func (s *Service) Create(ctx context.Context, a Actor, req ItemRequest) (*ItemView, error) {
	if !canWrite(a) {
		return nil, ErrForbidden
	}
	clinicID, err := s.targetClinic(ctx, a, req.ClinicID)
	if err != nil {
		return nil, err
	}
	it := &Item{
		ClinicID: clinicID, Name: strings.TrimSpace(req.Name), Category: strings.TrimSpace(req.Category),
		Unit: unitOr(req.Unit), Quantity: req.Quantity, MinQuantity: req.MinQuantity, PerSession: req.PerSession,
	}
	if err := s.checkName(ctx, clinicID, it.Name, ""); err != nil {
		return nil, err
	}
	var opening *Movement
	if req.Quantity > 0 {
		opening = &Movement{Kind: KindAdjust, Delta: req.Quantity, QuantityAfter: req.Quantity, Note: "opening stock", UserID: a.ID}
	}
	if err := s.repo.Create(ctx, it, opening); err != nil {
		return nil, err
	}
	v := view(*it)
	return &v, nil
}

// Update edits name, category, unit and minimum (not the stock).
func (s *Service) Update(ctx context.Context, a Actor, id string, req ItemRequest) (*ItemView, error) {
	if !canWrite(a) {
		return nil, ErrForbidden
	}
	it, err := s.Get(ctx, a, id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if err := s.checkName(ctx, it.ClinicID, name, it.ID); err != nil {
		return nil, err
	}
	fields := map[string]any{
		"name": name, "category": strings.TrimSpace(req.Category),
		"unit": unitOr(req.Unit), "min_quantity": req.MinQuantity, "per_session": req.PerSession,
	}
	if err := s.repo.Update(ctx, id, fields); err != nil {
		return nil, err
	}
	if it, err = s.Get(ctx, a, id); err != nil {
		return nil, err
	}
	v := view(*it)
	return &v, nil
}

// Archive removes an item from the list; its history is kept.
func (s *Service) Archive(ctx context.Context, a Actor, id string) error {
	if !canArchive(a) {
		return ErrForbidden
	}
	if _, err := s.Get(ctx, a, id); err != nil {
		return err
	}
	return s.repo.Update(ctx, id, map[string]any{"archived": true})
}

// Move records a restock, use or count and returns the updated item.
func (s *Service) Move(ctx context.Context, a Actor, id string, req MovementRequest) (*ItemView, error) {
	if !canWrite(a) {
		return nil, ErrForbidden
	}
	if req.Kind != KindAdjust && req.Quantity < 1 {
		return nil, &ValidationError{"quantity", "must be at least 1"}
	}
	if _, err := s.Get(ctx, a, id); err != nil {
		return nil, err
	}
	it, err := s.repo.Apply(ctx, id, &Movement{
		Kind: req.Kind, Delta: req.Quantity, Note: strings.TrimSpace(req.Note), UserID: a.ID,
	})
	if err != nil {
		return nil, err
	}
	v := view(*it)
	return &v, nil
}

// History lists the latest movements of an item, newest first.
func (s *Service) History(ctx context.Context, a Actor, id string) ([]Movement, error) {
	if _, err := s.Get(ctx, a, id); err != nil {
		return nil, err
	}
	return s.repo.Movements(ctx, id, 100)
}

// Audit is the clinic's log of movements (who used, opened, restocked or
// counted what, and when) for coordinators and admins; at most 1000 rows.
func (s *Service) Audit(ctx context.Context, a Actor, f AuditFilter) ([]Movement, error) {
	if !canAudit(a) {
		return nil, ErrForbidden
	}
	ids, err := s.clinicsFor(ctx, a)
	if err != nil {
		return nil, err
	}
	return s.repo.Audit(ctx, ids, f, 1000)
}

// SessionAttended uses the per-session quantity of every such item for an
// attended appointment (once). Items short of stock are skipped and
// reported as warnings; the session itself is never refused.
func (s *Service) SessionAttended(ctx context.Context, clinicID, appointmentID string, actorID uint) ([]string, error) {
	items, err := s.repo.SessionItems(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	done, err := s.repo.NetForAppointment(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	warnings := []string{}
	for _, it := range items {
		if done[it.ID] < 0 {
			continue // already used for this session
		}
		_, err := s.repo.Apply(ctx, it.ID, &Movement{Kind: KindUse, Delta: it.PerSession,
			Note: "session", UserID: actorID, AppointmentID: &appointmentID})
		switch {
		case errors.Is(err, ErrInsufficient):
			warnings = append(warnings, it.Name+": not enough in stock")
		case err != nil:
			return warnings, err
		}
	}
	return warnings, nil
}

// SessionUndone gives back what SessionAttended used for the appointment.
func (s *Service) SessionUndone(ctx context.Context, _ string, appointmentID string, actorID uint) error {
	net, err := s.repo.NetForAppointment(ctx, appointmentID)
	if err != nil {
		return err
	}
	for itemID, n := range net {
		if n >= 0 {
			continue
		}
		if _, err := s.repo.Apply(ctx, itemID, &Movement{Kind: KindRestock, Delta: -n,
			Note: "session undone", UserID: actorID, AppointmentID: &appointmentID}); err != nil {
			return err
		}
	}
	return nil
}

// targetClinic: staff always use their own clinic; admins must name one.
func (s *Service) targetClinic(ctx context.Context, a Actor, requested *string) (string, error) {
	if auth.IsAdmin(a.Role) {
		if requested == nil {
			return "", &ValidationError{"clinic_id", "is required for administrators"}
		}
		return *requested, nil
	}
	ids, err := s.clinicsFor(ctx, a)
	if err != nil {
		return "", err
	}
	if requested != nil && *requested != ids[0] {
		return "", ErrForbidden
	}
	return ids[0], nil
}

// checkName rejects an empty name or one already used in the clinic.
func (s *Service) checkName(ctx context.Context, clinicID, name, selfID string) error {
	if len(name) < 2 {
		return &ValidationError{"name", "must have at least 2 characters"}
	}
	same, err := s.repo.List(ctx, []string{clinicID}, ListFilter{Search: name})
	if err != nil {
		return err
	}
	for _, it := range same {
		if it.ID != selfID && strings.EqualFold(it.Name, name) {
			return &ValidationError{"name", "an item with this name already exists"}
		}
	}
	return nil
}

func unitOr(u string) string {
	if u = strings.TrimSpace(u); u == "" {
		return "units"
	}
	return u
}
