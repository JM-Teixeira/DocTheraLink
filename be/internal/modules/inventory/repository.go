package inventory

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// ErrNotFound: no such item in the user's clinic.
	ErrNotFound = errors.New("item not found")
	// ErrInsufficient: using more than is in stock.
	ErrInsufficient = errors.New("not enough in stock")
)

// Repository is the persistence boundary of the module.
type Repository interface {
	List(ctx context.Context, clinicIDs []string, f ListFilter) ([]Item, error)
	Get(ctx context.Context, clinicIDs []string, id string) (*Item, error)
	Create(ctx context.Context, it *Item, opening *Movement) error
	Update(ctx context.Context, id string, fields map[string]any) error
	// Apply changes the stock atomically and records the movement; it
	// returns the item after the change.
	Apply(ctx context.Context, itemID string, m *Movement) (*Item, error)
	Movements(ctx context.Context, itemID string, limit int) ([]Movement, error)
	// Audit lists the movements of the clinics' items, newest first.
	Audit(ctx context.Context, clinicIDs []string, f AuditFilter, limit int) ([]Movement, error)
	// SessionItems are the clinic's items used automatically per session.
	SessionItems(ctx context.Context, clinicID string) ([]Item, error)
	// NetForAppointment sums the stock change per item recorded for an appointment.
	NetForAppointment(ctx context.Context, appointmentID string) (map[string]int, error)
}

// GormRepository implements Repository on PostgreSQL via GORM.
type GormRepository struct{ db *gorm.DB }

// NewGormRepository builds a repository on the given connection.
func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

// inClinics restricts to the given clinics; nil means every clinic (admins).
func inClinics(db *gorm.DB, clinicIDs []string) *gorm.DB {
	if clinicIDs == nil {
		return db
	}
	return db.Where("clinic_id IN ?", clinicIDs)
}

func (g *GormRepository) List(ctx context.Context, clinicIDs []string, f ListFilter) ([]Item, error) {
	db := inClinics(g.db.WithContext(ctx).Model(&Item{}), clinicIDs).Where("archived = false")
	if s := strings.TrimSpace(f.Search); s != "" {
		db = db.Where("name ILIKE ?", "%"+escapeLike(s)+"%")
	}
	if f.Category != "" {
		db = db.Where("category = ?", f.Category)
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
	err := inClinics(g.db.WithContext(ctx), clinicIDs).
		Where("archived = false").First(&it, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &it, err
}

func (g *GormRepository) Create(ctx context.Context, it *Item, opening *Movement) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(it).Error; err != nil {
			return err
		}
		if opening == nil {
			return nil
		}
		opening.ItemID = it.ID
		return tx.Omit("User").Create(opening).Error
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
		next, delta, err := nextQuantity(out.Quantity, m.Kind, m.Delta)
		if err != nil {
			return err
		}
		if err := tx.Model(&Item{}).Where("id = ?", itemID).Update("quantity", next).Error; err != nil {
			return err
		}
		out.Quantity = next
		m.ItemID, m.Delta, m.QuantityAfter = itemID, delta, next
		return tx.Omit("User").Create(m).Error
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
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
	err := g.db.WithContext(ctx).Where("clinic_id = ? AND per_session > 0 AND archived = false", clinicID).
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

// nextQuantity applies a movement to the current stock. For restock/use the
// amount is the (positive) quantity; for adjust it is the counted stock.
// It returns the new stock and the signed change.
func nextQuantity(current int, kind string, amount int) (next, delta int, err error) {
	switch kind {
	case KindRestock:
		next = current + amount
	case KindUse, KindOpen:
		next = current - amount
	case KindAdjust:
		next = amount
	default:
		return 0, 0, errors.New("unknown movement kind")
	}
	if next < 0 {
		return 0, 0, ErrInsufficient
	}
	return next, next - current, nil
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
