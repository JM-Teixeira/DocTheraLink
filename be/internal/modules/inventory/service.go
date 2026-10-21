package inventory

import (
	"context"
	"errors"
	"strings"
	"time"

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
//   - any non-read-only member uses, opens, finishes, restocks, counts and
//     adds items;
//   - coordinators and admins remove items, set per-division session use
//     and see the audit log.
type Service struct {
	repo    Repository
	clinics clinic.Resolver
	now     func() time.Time
}

// NewService builds the service.
func NewService(repo Repository, clinics clinic.Resolver) *Service {
	return &Service{repo: repo, clinics: clinics, now: time.Now}
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

func canManage(a Actor) bool {
	return auth.IsAdmin(a.Role) || a.Role == models.RoleCoordinator
}

func (s *Service) views(items []Item) []ItemView {
	now := s.now()
	out := make([]ItemView, len(items))
	for i, it := range items {
		out[i] = view(it, now)
	}
	return out
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
	return s.views(items), nil
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
		if it.OpenExpired {
			sum.OpenExpired++
		}
	}
	return sum, nil
}

// Reorder lists the items at or below their alert level with how many to
// order to get back to their target.
func (s *Service) Reorder(ctx context.Context, a Actor) ([]ItemView, error) {
	items, err := s.List(ctx, a, ListFilter{Status: "attention"})
	if err != nil {
		return nil, err
	}
	out := []ItemView{}
	for _, it := range items {
		if it.Reorder > 0 {
			out = append(out, it)
		}
	}
	return out, nil
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

// checkKind validates the kind-specific settings of an item request.
func checkKind(req *ItemRequest) error {
	if req.Tracking == "" {
		req.Tracking = TrackUnit
	}
	if req.Tracking == TrackContainer && req.PerSession > 0 {
		return &ValidationError{"per_session", "only single-use items are used per session"}
	}
	if req.Tracking == TrackUnit {
		req.OpenLifeDays = 0
	}
	return nil
}

// Create adds an item to the actor's clinic with its opening stock.
func (s *Service) Create(ctx context.Context, a Actor, req ItemRequest) (*ItemView, error) {
	if !canWrite(a) {
		return nil, ErrForbidden
	}
	if err := checkKind(&req); err != nil {
		return nil, err
	}
	clinicID, err := s.targetClinic(ctx, a, req.ClinicID)
	if err != nil {
		return nil, err
	}
	it := &Item{
		ClinicID: clinicID, Name: strings.TrimSpace(req.Name), Category: strings.TrimSpace(req.Category),
		Unit: unitOr(req.Unit), Tracking: req.Tracking, Quantity: req.Quantity, MinQuantity: req.MinQuantity,
		TargetQuantity: req.TargetQuantity, PerSession: req.PerSession, OpenLifeDays: req.OpenLifeDays,
	}
	if err := s.checkName(ctx, clinicID, it.Name, ""); err != nil {
		return nil, err
	}
	var opening *Movement
	if req.Quantity > 0 {
		opening = &Movement{Kind: KindAdjust, Amount: req.Quantity, Delta: req.Quantity, QuantityAfter: req.Quantity,
			Note: "opening stock", UserID: a.ID}
	}
	if err := s.repo.Create(ctx, it, opening); err != nil {
		return nil, err
	}
	v := view(*it, s.now())
	return &v, nil
}

// Update edits the item's settings (not its stock). The kind can only
// change while no container is open.
func (s *Service) Update(ctx context.Context, a Actor, id string, req ItemRequest) (*ItemView, error) {
	if !canWrite(a) {
		return nil, ErrForbidden
	}
	it, err := s.Get(ctx, a, id)
	if err != nil {
		return nil, err
	}
	if req.Tracking == "" {
		req.Tracking = it.Tracking
	}
	if err := checkKind(&req); err != nil {
		return nil, err
	}
	if req.Tracking != it.Tracking && it.InUse > 0 {
		return nil, &ValidationError{"tracking", "finish the open containers first"}
	}
	name := strings.TrimSpace(req.Name)
	if err := s.checkName(ctx, it.ClinicID, name, it.ID); err != nil {
		return nil, err
	}
	fields := map[string]any{
		"name": name, "category": strings.TrimSpace(req.Category), "unit": unitOr(req.Unit),
		"tracking": req.Tracking, "min_quantity": req.MinQuantity, "target_quantity": req.TargetQuantity,
		"per_session": req.PerSession, "open_life_days": req.OpenLifeDays,
	}
	if err := s.repo.Update(ctx, id, fields); err != nil {
		return nil, err
	}
	if req.Tracking == TrackContainer && len(it.Rules) > 0 {
		if err := s.repo.ReplaceRules(ctx, id, nil); err != nil {
			return nil, err
		}
	}
	if it, err = s.Get(ctx, a, id); err != nil {
		return nil, err
	}
	v := view(*it, s.now())
	return &v, nil
}

// SetSessionRules sets how many of a single-use item a session uses in
// particular divisions (others use the item's default).
func (s *Service) SetSessionRules(ctx context.Context, a Actor, id string, req SessionRulesRequest) (*ItemView, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	it, err := s.Get(ctx, a, id)
	if err != nil {
		return nil, err
	}
	if it.Tracking != TrackUnit {
		return nil, &ValidationError{"rules", "only single-use items are used per session"}
	}
	seen := map[string]bool{}
	rules := []SessionRule{}
	for _, r := range req.Rules {
		if seen[r.DivisionID] {
			return nil, &ValidationError{"rules", "a division appears twice"}
		}
		seen[r.DivisionID] = true
		rules = append(rules, SessionRule{DivisionID: r.DivisionID, Quantity: r.Quantity})
	}
	if err := s.repo.ReplaceRules(ctx, id, rules); err != nil {
		return nil, err
	}
	if it, err = s.Get(ctx, a, id); err != nil {
		return nil, err
	}
	v := view(*it, s.now())
	return &v, nil
}

// Archive removes an item from the list; its history is kept.
func (s *Service) Archive(ctx context.Context, a Actor, id string) error {
	if !canManage(a) {
		return ErrForbidden
	}
	if _, err := s.Get(ctx, a, id); err != nil {
		return err
	}
	return s.repo.Update(ctx, id, map[string]any{"archived": true})
}

// Move records a movement and returns the updated item. Single-use items
// are used; containers are opened and finished; both are restocked and
// counted.
func (s *Service) Move(ctx context.Context, a Actor, id string, req MovementRequest) (*ItemView, error) {
	if !canWrite(a) {
		return nil, ErrForbidden
	}
	if req.Kind != KindAdjust && req.Quantity < 1 {
		return nil, &ValidationError{"quantity", "must be at least 1"}
	}
	it, err := s.Get(ctx, a, id)
	if err != nil {
		return nil, err
	}
	switch {
	case req.Kind == KindUse && it.Tracking != TrackUnit:
		return nil, &ValidationError{"kind", "containers are opened and finished, not used"}
	case (req.Kind == KindOpen || req.Kind == KindFinish) && it.Tracking != TrackContainer:
		return nil, &ValidationError{"kind", "single-use items are used, not opened"}
	}
	updated, err := s.repo.Apply(ctx, id, &Movement{
		Kind: req.Kind, Amount: req.Quantity, Note: strings.TrimSpace(req.Note), UserID: a.ID,
	})
	if err != nil {
		return nil, err
	}
	updated.Rules = it.Rules
	v := view(*updated, s.now())
	return &v, nil
}

// History lists the latest movements of an item, newest first.
func (s *Service) History(ctx context.Context, a Actor, id string) ([]Movement, error) {
	if _, err := s.Get(ctx, a, id); err != nil {
		return nil, err
	}
	return s.repo.Movements(ctx, id, 100)
}

// Audit is the clinic's log of movements (who used, opened, finished,
// restocked or counted what, and when) for coordinators and admins; at
// most 1000 rows.
func (s *Service) Audit(ctx context.Context, a Actor, f AuditFilter) ([]Movement, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	ids, err := s.clinicsFor(ctx, a)
	if err != nil {
		return nil, err
	}
	return s.repo.Audit(ctx, ids, f, 1000)
}

// SessionAttended uses, once per appointment, what a session in the
// division needs (each single-use item's quantity for that division).
// Items short of stock are skipped and reported; the session is never refused.
func (s *Service) SessionAttended(ctx context.Context, clinicID, appointmentID, divisionID string, actorID uint) ([]string, error) {
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
		qty := it.PerSessionIn(divisionID)
		if qty <= 0 || done[it.ID] < 0 {
			continue // not used in this division, or already used for this session
		}
		_, err := s.repo.Apply(ctx, it.ID, &Movement{Kind: KindUse, Amount: qty,
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
		if _, err := s.repo.Apply(ctx, itemID, &Movement{Kind: KindRestock, Amount: -n,
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
