package attachment

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// ErrNotFound is returned when an attachment does not exist (or is hidden).
var ErrNotFound = errors.New("attachment not found")

// Repository is the persistence boundary of the module.
type Repository interface {
	Create(ctx context.Context, a *Attachment) error
	Get(ctx context.Context, id string) (*Attachment, error)
	ListByReferral(ctx context.Context, referralID string) ([]Attachment, error)
	Delete(ctx context.Context, id string) error
}

// GormRepository implements Repository on PostgreSQL via GORM.
type GormRepository struct{ db *gorm.DB }

// NewGormRepository builds a repository on the given connection.
func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (g *GormRepository) Create(ctx context.Context, a *Attachment) error {
	if err := g.db.WithContext(ctx).Omit("Uploader").Create(a).Error; err != nil {
		return err
	}
	return g.db.WithContext(ctx).Preload("Uploader").First(a, "id = ?", a.ID).Error
}

func (g *GormRepository) Get(ctx context.Context, id string) (*Attachment, error) {
	var a Attachment
	err := g.db.WithContext(ctx).Preload("Uploader").First(&a, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (g *GormRepository) ListByReferral(ctx context.Context, referralID string) ([]Attachment, error) {
	var out []Attachment
	err := g.db.WithContext(ctx).Preload("Uploader").
		Where("referral_id = ?", referralID).Order("created_at ASC").Find(&out).Error
	return out, err
}

func (g *GormRepository) Delete(ctx context.Context, id string) error {
	return g.db.WithContext(ctx).Delete(&Attachment{}, "id = ?", id).Error
}
