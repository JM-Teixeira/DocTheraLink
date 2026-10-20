// Package patient implements patient registration, lookup and editing.
//
// Layout mirrors the referral module:
//
//	repository.go - persistence (Repository interface + GORM implementation)
//	service.go    - access scoping and validation
//	handler.go    - HTTP handlers
//	routes.go     - route registration
package patient

import (
	"context"
	"errors"
	"strings"

	"klyniva/internal/models"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a patient does not exist or is not visible.
var ErrNotFound = errors.New("patient not found")

// Scope restricts which patients a query may return.
type Scope struct {
	All      bool    // admins: no restriction
	UserID   uint    // the requesting user
	ClinicID *string // the user's clinic (see internal/clinic), if any
}

// ListQuery holds the list/search parameters.
type ListQuery struct {
	Search     string `form:"search"`
	Phone      string `form:"phone_number"`
	FacilityID string `form:"facility"`
	IsActive   *bool  `form:"is_active"`
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
}

// Repository is the persistence boundary of the module.
type Repository interface {
	List(ctx context.Context, s Scope, q ListQuery) ([]models.Patient, int64, error)
	Get(ctx context.Context, s Scope, id string) (*models.Patient, error)
	Create(ctx context.Context, p *models.Patient) error
	Update(ctx context.Context, id string, fields map[string]any) error
}

// GormRepository implements Repository on PostgreSQL via GORM.
type GormRepository struct{ db *gorm.DB }

// NewGormRepository builds a repository on the given connection.
func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

// scoped applies the visibility rules: a non-admin sees patients of their
// clinic, patients they created or are assigned to, and patients linked to
// them through a referral (as referring doctor or treating therapist).
func scoped(db *gorm.DB, s Scope) *gorm.DB {
	if s.All {
		return db
	}
	// Grouped conditions must be built on a fresh session, not the outer query.
	fresh := db.Session(&gorm.Session{NewDB: true})
	cond := fresh.Where("patients.created_by_id = ?", s.UserID).
		Or("patients.assigned_to_id = ?", s.UserID).
		Or("patients.id IN (?)", fresh.Table("referrals").
			Select("patient_id").
			Where("deleted_at IS NULL AND (doctor_id = ? OR therapist_id = ?)", s.UserID, s.UserID))
	if s.ClinicID != nil {
		cond = cond.Or("patients.facility_id = ?", *s.ClinicID)
	}
	return db.Where(cond)
}

func (g *GormRepository) List(ctx context.Context, s Scope, q ListQuery) ([]models.Patient, int64, error) {
	db := scoped(g.db.WithContext(ctx).Model(&models.Patient{}), s)
	if term := strings.TrimSpace(q.Search); term != "" {
		db = db.Where("name ILIKE ?", "%"+escapeLike(term)+"%")
	}
	if q.Phone != "" {
		db = db.Where("phone_number = ?", q.Phone)
	}
	if q.FacilityID != "" {
		db = db.Where("facility_id = ?", q.FacilityID)
	}
	if q.IsActive != nil {
		db = db.Where("is_active = ?", *q.IsActive)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.Patient
	err := db.Order("created_at DESC").Limit(q.Limit).Offset(q.Offset).Find(&items).Error
	return items, total, err
}

func (g *GormRepository) Get(ctx context.Context, s Scope, id string) (*models.Patient, error) {
	var p models.Patient
	err := scoped(g.db.WithContext(ctx).Model(&models.Patient{}), s).
		Preload("Facility").
		First(&p, "patients.id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (g *GormRepository) Create(ctx context.Context, p *models.Patient) error {
	// Omit associations so a client-supplied nested object never writes other tables.
	return g.db.WithContext(ctx).Omit("Facility", "State", "District", "LocalBody", "Ward",
		"AssignedTo", "CreatedBy", "LastEdited").Create(p).Error
}

func (g *GormRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	res := g.db.WithContext(ctx).Model(&models.Patient{}).Where("id = ?", id).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// escapeLike escapes LIKE wildcards so user input is matched literally.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
