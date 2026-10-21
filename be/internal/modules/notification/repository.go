package notification

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a notification does not exist for the user.
var ErrNotFound = errors.New("notification not found")

// Repository is the persistence boundary of the module.
type Repository interface {
	CreateMany(ctx context.Context, ns []Notification) error
	List(ctx context.Context, userID uint, q ListQuery) ([]Notification, int64, error)
	UnreadCount(ctx context.Context, userID uint) (int64, error)
	MarkRead(ctx context.Context, userID uint, id string, at time.Time) error
	MarkAllRead(ctx context.Context, userID uint, at time.Time) error
	// People returns active users by ID (unknown or inactive IDs are skipped).
	People(ctx context.Context, ids []uint) ([]Person, error)
	// Therapists returns the active physiotherapists of a clinic (all of them
	// when clinicID is nil).
	Therapists(ctx context.Context, clinicID *string) ([]uint, error)
}

// GormRepository implements Repository on PostgreSQL via GORM.
type GormRepository struct{ db *gorm.DB }

// NewGormRepository builds a repository on the given connection.
func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (g *GormRepository) CreateMany(ctx context.Context, ns []Notification) error {
	if len(ns) == 0 {
		return nil
	}
	return g.db.WithContext(ctx).Create(&ns).Error
}

func (g *GormRepository) List(ctx context.Context, userID uint, q ListQuery) ([]Notification, int64, error) {
	db := g.db.WithContext(ctx).Model(&Notification{}).Where("user_id = ?", userID)
	if q.Unread {
		db = db.Where("read_at IS NULL")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []Notification
	err := db.Order("created_at DESC").Limit(q.Limit).Offset(q.Offset).Find(&items).Error
	return items, total, err
}

func (g *GormRepository) UnreadCount(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := g.db.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).Count(&n).Error
	return n, err
}

func (g *GormRepository) MarkRead(ctx context.Context, userID uint, id string, at time.Time) error {
	// Scoped by user: nobody can touch someone else's notifications.
	res := g.db.WithContext(ctx).Model(&Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read_at", gorm.Expr("COALESCE(read_at, ?)", at))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (g *GormRepository) MarkAllRead(ctx context.Context, userID uint, at time.Time) error {
	return g.db.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).Update("read_at", at).Error
}

func (g *GormRepository) People(ctx context.Context, ids []uint) ([]Person, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var out []Person
	err := g.db.WithContext(ctx).Table("users").
		Select("id, email, first_name, last_name").
		Where("id IN ? AND is_active AND deleted_at IS NULL", ids).
		Order("id").Scan(&out).Error
	return out, err
}

func (g *GormRepository) Therapists(ctx context.Context, clinicID *string) ([]uint, error) {
	q := g.db.WithContext(ctx).Table("users").
		Where("user_type = ? AND is_active AND deleted_at IS NULL", "PhysioTherapist")
	if clinicID != nil {
		q = q.Where("home_facility_id = ?", *clinicID)
	}
	var ids []uint
	err := q.Order("id").Pluck("id", &ids).Error
	return ids, err
}
