// Package clinic resolves which clinic a user works in.
//
// Klyniva runs as a single clinic today but is built for several: a clinic is
// a row in `facilities`, and a user's `home_facility_id` is the clinic they
// work in. Modules scope their data by clinic through this package, so adding
// multi-clinic organisations later means changing the resolver, not every query.
package clinic

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Resolver returns the clinic ID of a user (nil when the user has no clinic,
// e.g. organisation-wide admins or not-yet-configured accounts).
type Resolver interface {
	HomeClinic(ctx context.Context, userID uint) (*string, error)
}

// GormResolver reads users.home_facility_id.
type GormResolver struct{ db *gorm.DB }

// NewGormResolver builds a resolver on the given connection.
func NewGormResolver(db *gorm.DB) *GormResolver { return &GormResolver{db: db} }

// HomeClinic returns the user's home clinic, or nil if unset/unknown user.
func (g *GormResolver) HomeClinic(ctx context.Context, userID uint) (*string, error) {
	var row struct{ HomeFacilityID *string }
	err := g.db.WithContext(ctx).Table("users").Select("home_facility_id").
		Where("id = ?", userID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return row.HomeFacilityID, err
}

// SameClinic reports whether two clinic IDs may share data: true when they
// match or either is unset (single-clinic setups without clinic assignment).
func SameClinic(a, b *string) bool {
	return a == nil || b == nil || *a == *b
}

// Static is a Resolver backed by a map, for tests.
type Static map[uint]string

// HomeClinic implements Resolver.
func (s Static) HomeClinic(_ context.Context, userID uint) (*string, error) {
	if c, ok := s[userID]; ok {
		return &c, nil
	}
	return nil, nil
}
