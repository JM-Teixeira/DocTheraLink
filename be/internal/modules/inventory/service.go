package inventory

import (
	"context"
	"errors"
	"slices"
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
//   - everyone sees the stock of their own clinic (admins: every clinic,
//     or one with the clinic filter);
//   - coordinators and admins change any item; other people only the items
//     their grants cover (floors × categories) - use, open, finish, restock,
//     count, add and edit; without a grant the stock is read-only;
//   - coordinators and admins remove items, set per-division session use,
//     give grants and see the audit log;
//   - sessions marked attended use their per-session items automatically,
//     whoever marks them.
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

func canManage(a Actor) bool {
	return auth.IsAdmin(a.Role) || a.Role == models.RoleCoordinator
}

// narrow applies an admin's clinic filter; staff may only name their own.
func narrow(ids []string, want string) ([]string, error) {
	switch {
	case want == "":
		return ids, nil
	case ids == nil || slices.Contains(ids, want):
		return []string{want}, nil
	}
	return nil, ErrForbidden
}

// access is what the actor may change: everything (managers), what their
// grants cover, or nothing (read-only roles, no grant).
func (s *Service) access(ctx context.Context, a Actor) (Access, error) {
	switch {
	case auth.IsReadOnlyUser(a.Role):
		return Access{Grants: []Grant{}}, nil
	case canManage(a):
		return Access{Manage: true, Grants: []Grant{}}, nil
	}
	ids, err := s.clinicsFor(ctx, a)
	if err != nil {
		return Access{}, err
	}
	grants, err := s.repo.Grants(ctx, ids, a.ID)
	if err != nil {
		return Access{}, err
	}
	if grants == nil {
		grants = []Grant{}
	}
	return Access{Grants: grants}, nil
}

// Allows reports whether an item on that floor and in that category may be changed.
func (acc Access) Allows(floor *int, category string) bool {
	return acc.Manage || slices.ContainsFunc(acc.Grants, func(g Grant) bool { return g.Covers(floor, category) })
}

func (s *Service) views(items []Item, acc Access) []ItemView {
	now := s.now()
	out := make([]ItemView, len(items))
	for i, it := range items {
		out[i] = view(it, now)
		out[i].CanChange = acc.Allows(it.Floor, it.Category)
	}
	return out
}

// MyAccess is what the viewer may change (for the page's buttons).
func (s *Service) MyAccess(ctx context.Context, a Actor) (Access, error) { return s.access(ctx, a) }

// List returns the clinic's items, by category and name.
func (s *Service) List(ctx context.Context, a Actor, f ListFilter) ([]ItemView, error) {
	ids, err := s.clinicsFor(ctx, a)
	if err != nil {
		return nil, err
	}
	if ids, err = narrow(ids, f.Clinic); err != nil {
		return nil, err
	}
	acc, err := s.access(ctx, a)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx, ids, f)
	if err != nil {
		return nil, err
	}
	return s.views(items, acc), nil
}

// Summary counts all items and those needing attention.
func (s *Service) Summary(ctx context.Context, a Actor, clinicFilter string) (Summary, error) {
	items, err := s.List(ctx, a, ListFilter{Clinic: clinicFilter})
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
func (s *Service) Reorder(ctx context.Context, a Actor, clinicFilter string) ([]ItemView, error) {
	items, err := s.List(ctx, a, ListFilter{Status: "attention", Clinic: clinicFilter})
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
	acc, err := s.access(ctx, a)
	if err != nil {
		return nil, err
	}
	if !acc.Allows(req.Floor, req.Category) {
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
		Floor: req.Floor,
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
	v.CanChange = true
	return &v, nil
}

// Update edits the item's settings (not its stock). The kind can only
// change while no container is open. Moving it to another floor or
// category needs a grant for both the old and the new place.
func (s *Service) Update(ctx context.Context, a Actor, id string, req ItemRequest) (*ItemView, error) {
	acc, err := s.access(ctx, a)
	if err != nil {
		return nil, err
	}
	it, err := s.Get(ctx, a, id)
	if err != nil {
		return nil, err
	}
	if !acc.Allows(it.Floor, it.Category) || !acc.Allows(req.Floor, req.Category) {
		return nil, ErrForbidden
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
		"per_session": req.PerSession, "open_life_days": req.OpenLifeDays, "floor": req.Floor,
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
	v.CanChange = true
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
	v.CanChange = true
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
	if req.Kind != KindAdjust && req.Quantity < 1 {
		return nil, &ValidationError{"quantity", "must be at least 1"}
	}
	acc, err := s.access(ctx, a)
	if err != nil {
		return nil, err
	}
	it, err := s.Get(ctx, a, id)
	if err != nil {
		return nil, err
	}
	if !acc.Allows(it.Floor, it.Category) {
		return nil, ErrForbidden
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
	v.CanChange = true
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
	if ids, err = narrow(ids, f.Clinic); err != nil {
		return nil, err
	}
	return s.repo.Audit(ctx, ids, f, 1000)
}

// Grants lists who may change which stock in the clinic(s) (managers).
func (s *Service) Grants(ctx context.Context, a Actor, clinicFilter string) ([]Grant, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	ids, err := s.clinicsFor(ctx, a)
	if err != nil {
		return nil, err
	}
	if ids, err = narrow(ids, clinicFilter); err != nil {
		return nil, err
	}
	out, err := s.repo.Grants(ctx, ids, 0)
	if out == nil {
		out = []Grant{}
	}
	return out, err
}

// SetGrants replaces what a person may change in their clinic's stock:
// coordinators for people of their own clinic, admins for anyone.
func (s *Service) SetGrants(ctx context.Context, a Actor, userID uint, req GrantsRequest) ([]Grant, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	home, err := s.clinics.HomeClinic(ctx, userID)
	if err != nil {
		return nil, err
	}
	if home == nil {
		return nil, &ValidationError{"user", "works in no clinic"}
	}
	if !auth.IsAdmin(a.Role) {
		mine, err := s.clinicsFor(ctx, a)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(mine, *home) {
			return nil, ErrNotFound
		}
	}
	grants := make([]Grant, 0, len(req.Rules))
	for _, r := range req.Rules {
		grants = append(grants, Grant{ClinicID: *home, UserID: userID, CreatedByID: a.ID,
			Floors: uniqueFloors(r.Floors), Categories: uniqueCategories(r.Categories)})
	}
	if err := s.repo.ReplaceGrants(ctx, *home, userID, grants); err != nil {
		return nil, err
	}
	out, err := s.repo.Grants(ctx, []string{*home}, userID)
	if out == nil {
		out = []Grant{}
	}
	return out, err
}

func uniqueFloors(in []int) []int {
	out := []int{}
	for _, f := range in {
		if !slices.Contains(out, f) {
			out = append(out, f)
		}
	}
	slices.Sort(out)
	return out
}

func uniqueCategories(in []string) []string {
	out := []string{}
	for _, c := range in {
		c = strings.TrimSpace(c)
		if c != "" && !slices.ContainsFunc(out, func(x string) bool { return strings.EqualFold(x, c) }) {
			out = append(out, c)
		}
	}
	return out
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
