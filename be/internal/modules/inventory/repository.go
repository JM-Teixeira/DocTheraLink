package inventory

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// ErrNotFound: no such item in the user's clinic.
	ErrNotFound = errors.New("item not found")
	// ErrInsufficient: using or opening more than is in stock.
	ErrInsufficient = errors.New("not enough in stock")
	// ErrNothingOpen: finishing more containers than are open.
	ErrNothingOpen = errors.New("not that many are open")
)

// Repository is the persistence boundary of the module.
type Repository interface {
	List(ctx context.Context, clinicIDs []string, f ListFilter) ([]Item, error)
	Get(ctx context.Context, clinicIDs []string, id string) (*Item, error)
	Create(ctx context.Context, it *Item, opening *Movement) error
	Update(ctx context.Context, id string, fields map[string]any) error
	// Apply changes the stock atomically and records the movement (m.Amount
	// is the requested amount); it returns the item after the change.
	Apply(ctx context.Context, itemID string, m *Movement) (*Item, error)
	Movements(ctx context.Context, itemID string, limit int) ([]Movement, error)
	// Audit lists the movements of the clinics' items, newest first.
	Audit(ctx context.Context, clinicIDs []string, f AuditFilter, limit int) ([]Movement, error)
	// SessionItems are the clinic's unit items used per session (by default
	// or in some division), with their rules.
	SessionItems(ctx context.Context, clinicID string) ([]Item, error)
	// NetForAppointment sums the stock change per item recorded for an appointment.
	NetForAppointment(ctx context.Context, appointmentID string) (map[string]int, error)
	ReplaceRules(ctx context.Context, itemID string, rules []SessionRule) error
	// Grants of the clinics (nil = all) for one user (0 = everyone).
	Grants(ctx context.Context, clinicIDs []string, userID uint) ([]Grant, error)
	// ReplaceGrants sets all the grants of a user in a clinic.
	ReplaceGrants(ctx context.Context, clinicID string, userID uint, grants []Grant) error
}

// GormRepository implements Repository on PostgreSQL via GORM.
type GormRepository struct {
	db  *gorm.DB
	now func() time.Time
}

// NewGormRepository builds a repository on the given connection.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db, now: time.Now}
}

// inClinics restricts to the given clinics; nil means every clinic (admins).
func inClinics(db *gorm.DB, clinicIDs []string) *gorm.DB {
	if clinicIDs == nil {
		return db
	}
	return db.Where("clinic_id IN ?", clinicIDs)
}

func (g *GormRepository) Grants(ctx context.Context, clinicIDs []string, userID uint) ([]Grant, error) {
	db := inClinics(g.db.WithContext(ctx).Model(&Grant{}), clinicIDs)
	if userID != 0 {
		db = db.Where("user_id = ?", userID)
	}
	var out []Grant
	err := db.Order("user_id, created_at").Find(&out).Error
	return out, err
}

func (g *GormRepository) ReplaceGrants(ctx context.Context, clinicID string, userID uint, grants []Grant) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("clinic_id = ? AND user_id = ?", clinicID, userID).Delete(&Grant{}).Error; err != nil {
			return err
		}
		if len(grants) == 0 {
			return nil
		}
		return tx.Create(&grants).Error
	})
}

func (g *GormRepository) List(ctx context.Context, clinicIDs []string, f ListFilter) ([]Item, error) {
	db := inClinics(g.db.WithContext(ctx).Model(&Item{}), clinicIDs).Preload("Rules").Where("archived = false")
	if s := strings.TrimSpace(f.Search); s != "" {
		db = db.Where("name ILIKE ?", "%"+escapeLike(s)+"%")
	}
	if f.Category != "" {
		db = db.Where("category = ?", f.Category)
	}
	if f.Tracking != "" {
		db = db.Where("tracking = ?", f.Tracking)
	}
	if f.Floor != nil {
		db = db.Where("floor = ?", *f.Floor)
	}
	switch f.Status {
	case "out":
		db = db.Where("quantity <= 0")
	case "low":
		db = db.Where("quantity > 0 AND quantity <= min_quantity")
	case "attention":
		db = db.Where("quantity <= min_quantity OR quantity <= 0")
	case "ok":
		db = db.Where("quantity > min_quantity AND quantity > 0")
	}
	var items []Item
	err := db.Order("category, name").Limit(500).Find(&items).Error
	return items, err
}

func (g *GormRepository) Get(ctx context.Context, clinicIDs []string, id string) (*Item, error) {
	var it Item
	err := inClinics(g.db.WithContext(ctx), clinicIDs).Preload("Rules").
		Where("archived = false").First(&it, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &it, err
}

func (g *GormRepository) Create(ctx context.Context, it *Item, opening *Movement) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Rules").Create(it).Error; err != nil {
			return err
		}
		if opening == nil {
			return nil
		}
		opening.ItemID = it.ID
		return tx.Omit("User", "Item").Create(opening).Error
	})
}

func (g *GormRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	res := g.db.WithContext(ctx).Model(&Item{}).Where("id = ?", id).Updates(fields)
	if res.Error == nil && res.RowsAffected == 0 {
		return ErrNotFound
	}
	return res.Error
}

// Apply locks the item row, so concurrent uses can never take stock below zero.
func (g *GormRepository) Apply(ctx context.Context, itemID string, m *Movement) (*Item, error) {
	var out Item
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&out, "id = ?", itemID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		next, err := nextState(out.Quantity, out.InUse, m.Kind, m.Amount)
		if err != nil {
			return err
		}
		oldest, err := g.oldestOpen(tx, out, next, m)
		if err != nil {
			return err
		}
		if err := tx.Model(&Item{}).Where("id = ?", itemID).Updates(map[string]any{
			"quantity": next.quantity, "in_use": next.inUse, "oldest_open_at": oldest,
		}).Error; err != nil {
			return err
		}
		out.Quantity, out.InUse, out.OldestOpenAt = next.quantity, next.inUse, oldest
		m.ItemID, m.Delta, m.QuantityAfter, m.InUseAfter = itemID, next.delta, next.quantity, next.inUse
		return tx.Omit("User", "Item").Create(m).Error
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// oldestOpen keeps track of when the oldest still-open container was
// opened (containers are finished first-in, first-out).
func (g *GormRepository) oldestOpen(tx *gorm.DB, before Item, next state, m *Movement) (*time.Time, error) {
	switch {
	case next.inUse == 0:
		return nil, nil
	case m.Kind == KindOpen && before.InUse == 0:
		now := g.now()
		return &now, nil
	case m.Kind != KindFinish:
		return before.OldestOpenAt, nil
	}
	var opens []struct {
		Amount    int
		CreatedAt time.Time
	}
	if err := tx.Model(&Movement{}).Select("amount, created_at").
		Where("item_id = ? AND kind = ?", before.ID, KindOpen).Order("created_at").Scan(&opens).Error; err != nil {
		return nil, err
	}
	var finished int
	if err := tx.Model(&Movement{}).Select("COALESCE(SUM(amount), 0)").
		Where("item_id = ? AND kind = ?", before.ID, KindFinish).Scan(&finished).Error; err != nil {
		return nil, err
	}
	finished += m.Amount
	cum := 0
	for _, o := range opens {
		cum += o.Amount
		if cum > finished {
			t := o.CreatedAt
			return &t, nil
		}
	}
	return before.OldestOpenAt, nil
}

func (g *GormRepository) Movements(ctx context.Context, itemID string, limit int) ([]Movement, error) {
	var ms []Movement
	err := g.db.WithContext(ctx).Preload("User").Where("item_id = ?", itemID).
		Order("created_at DESC").Limit(limit).Find(&ms).Error
	return ms, err
}

func (g *GormRepository) Audit(ctx context.Context, clinicIDs []string, f AuditFilter, limit int) ([]Movement, error) {
	items := g.db.Model(&Item{}).Select("id")
	if clinicIDs != nil {
		items = items.Where("clinic_id IN ?", clinicIDs)
	}
	db := g.db.WithContext(ctx).Preload("User").Preload("Item").Where("item_id IN (?)", items)
	if f.ItemID != "" {
		db = db.Where("item_id = ?", f.ItemID)
	}
	if f.UserID != 0 {
		db = db.Where("user_id = ?", f.UserID)
	}
	if f.Kind != "" {
		db = db.Where("kind = ?", f.Kind)
	}
	if f.From != nil {
		db = db.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		db = db.Where("created_at < ?", f.To.AddDate(0, 0, 1)) // inclusive end date
	}
	if f.Auto != nil {
		if *f.Auto {
			db = db.Where("appointment_id IS NOT NULL")
		} else {
			db = db.Where("appointment_id IS NULL")
		}
	}
	var ms []Movement
	err := db.Order("created_at DESC").Limit(limit).Find(&ms).Error
	return ms, err
}

func (g *GormRepository) SessionItems(ctx context.Context, clinicID string) ([]Item, error) {
	var items []Item
	err := g.db.WithContext(ctx).Preload("Rules").
		Where("clinic_id = ? AND tracking = ? AND archived = false", clinicID, TrackUnit).
		Where("per_session > 0 OR id IN (SELECT item_id FROM stock_session_rules WHERE quantity > 0)").
		Order("name").Find(&items).Error
	return items, err
}

func (g *GormRepository) NetForAppointment(ctx context.Context, appointmentID string) (map[string]int, error) {
	var rows []struct {
		ItemID string
		Net    int
	}
	err := g.db.WithContext(ctx).Model(&Movement{}).Select("item_id, SUM(delta) AS net").
		Where("appointment_id = ?", appointmentID).Group("item_id").Scan(&rows).Error
	out := map[string]int{}
	for _, r := range rows {
		out[r.ItemID] = r.Net
	}
	return out, err
}

func (g *GormRepository) ReplaceRules(ctx context.Context, itemID string, rules []SessionRule) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("item_id = ?", itemID).Delete(&SessionRule{}).Error; err != nil {
			return err
		}
		if len(rules) == 0 {
			return nil
		}
		for i := range rules {
			rules[i].ItemID = itemID
		}
		return tx.Create(&rules).Error
	})
}

// state is the stock of an item after a movement.
type state struct {
	quantity, inUse, delta int
}

// nextState applies a movement (amount is positive; for adjust it is the
// counted stock). It returns the new sealed/available stock, how many are
// open, and the signed change of the stock.
func nextState(quantity, inUse int, kind string, amount int) (state, error) {
	next := state{quantity: quantity, inUse: inUse}
	switch kind {
	case KindRestock:
		next.quantity += amount
	case KindUse:
		next.quantity -= amount
	case KindOpen:
		next.quantity -= amount
		next.inUse += amount
	case KindFinish:
		next.inUse -= amount
		if next.inUse < 0 {
			return state{}, ErrNothingOpen
		}
	case KindAdjust:
		next.quantity = amount
	default:
		return state{}, errors.New("unknown movement kind")
	}
	if next.quantity < 0 {
		return state{}, ErrInsufficient
	}
	next.delta = next.quantity - quantity
	return next, nil
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
