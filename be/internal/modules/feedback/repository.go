package feedback

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// GormRepository stores suggestions in PostgreSQL.
type GormRepository struct{ db *gorm.DB }

// NewGormRepository builds a GormRepository.
func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (g *GormRepository) Add(ctx context.Context, s *Suggestion) error {
	return g.db.WithContext(ctx).Create(s).Error
}

// List is newest first; new ones first for admins (what still needs reading).
func (g *GormRepository) List(ctx context.Context, authorID uint, status string) ([]Suggestion, error) {
	var out []Suggestion
	q := g.db.WithContext(ctx).Preload("Author")
	if authorID != 0 {
		q = q.Where("author_id = ?", authorID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("CASE WHEN status = 'new' THEN 0 ELSE 1 END, created_at DESC").Limit(500).Find(&out).Error
	return out, err
}

func (g *GormRepository) Get(ctx context.Context, id string) (*Suggestion, error) {
	var s Suggestion
	err := g.db.WithContext(ctx).Preload("Author").First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (g *GormRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	return g.db.WithContext(ctx).Model(&Suggestion{}).Where("id = ?", id).Updates(fields).Error
}

func (g *GormRepository) CountNew(ctx context.Context) (int64, error) {
	var n int64
	err := g.db.WithContext(ctx).Model(&Suggestion{}).Where("status = ?", "new").Count(&n).Error
	return n, err
}
