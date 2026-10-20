package referral

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	// ErrNotFound is returned when a referral does not exist.
	ErrNotFound = errors.New("referral not found")
	// ErrConflict is returned when the referral changed status concurrently.
	ErrConflict = errors.New("referral was modified by someone else")
)

// Repository is the persistence boundary of the module. The service only
// depends on this interface, which keeps business rules testable without a database.
type Repository interface {
	Create(ctx context.Context, r *Referral) error
	Get(ctx context.Context, id string) (*Referral, error)
	List(ctx context.Context, f ListFilter) ([]Referral, int64, error)
	// Transition saves r only if its stored status still equals from, and
	// appends note in the same transaction. Returns ErrConflict otherwise.
	Transition(ctx context.Context, r *Referral, from Status, note *Note) error
	AddNote(ctx context.Context, n *Note) error
	ListNotes(ctx context.Context, referralID string) ([]Note, error)
	PatientExists(ctx context.Context, id string) (bool, error)
	// UserRole returns the user's role and whether the account is active.
	UserRole(ctx context.Context, id uint) (role string, active bool, err error)
}

// GormRepository implements Repository on PostgreSQL via GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository builds a repository on the given connection.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// withRelations preloads the summaries embedded in API responses.
func withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Patient").Preload("Doctor").Preload("Therapist")
}

func (g *GormRepository) Create(ctx context.Context, r *Referral) error {
	return g.db.WithContext(ctx).Create(r).Error
}

func (g *GormRepository) Get(ctx context.Context, id string) (*Referral, error) {
	var r Referral
	err := withRelations(g.db.WithContext(ctx)).First(&r, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (g *GormRepository) List(ctx context.Context, f ListFilter) ([]Referral, int64, error) {
	q := g.db.WithContext(ctx).Model(&Referral{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.PatientID != "" {
		q = q.Where("patient_id = ?", f.PatientID)
	}
	if f.DoctorID != 0 {
		q = q.Where("doctor_id = ?", f.DoctorID)
	}
	switch {
	case f.TherapistID != 0 && f.InboxForTherapist && f.ClinicID != nil:
		// A therapist's inbox: everything assigned to them plus open, unassigned
		// referrals of their clinic (or without a clinic).
		q = q.Where("therapist_id = ? OR (therapist_id IS NULL AND status = ? AND (facility_id IS NULL OR facility_id = ?))",
			f.TherapistID, StatusPending, *f.ClinicID)
	case f.TherapistID != 0 && f.InboxForTherapist:
		q = q.Where("therapist_id = ? OR (therapist_id IS NULL AND status = ?)", f.TherapistID, StatusPending)
	case f.TherapistID != 0:
		q = q.Where("therapist_id = ?", f.TherapistID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []Referral
	err := withRelations(q).
		Order("CASE WHEN priority = 'urgent' THEN 0 ELSE 1 END, created_at DESC").
		Limit(f.Limit).Offset(f.Offset).
		Find(&items).Error
	return items, total, err
}

func (g *GormRepository) Transition(ctx context.Context, r *Referral, from Status, note *Note) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Referral{}).
			Where("id = ? AND status = ?", r.ID, from).
			Select("therapist_id", "status", "decline_reason", "cancel_reason",
				"outcome_summary", "accepted_at", "started_at", "completed_at", "updated_at").
			Updates(r)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrConflict
		}
		if note != nil {
			return tx.Create(note).Error
		}
		return nil
	})
}

func (g *GormRepository) AddNote(ctx context.Context, n *Note) error {
	return g.db.WithContext(ctx).Create(n).Error
}

func (g *GormRepository) ListNotes(ctx context.Context, referralID string) ([]Note, error) {
	var notes []Note
	err := g.db.WithContext(ctx).Preload("Author").
		Where("referral_id = ?", referralID).
		Order("created_at ASC").
		Find(&notes).Error
	return notes, err
}

func (g *GormRepository) PatientExists(ctx context.Context, id string) (bool, error) {
	var n int64
	err := g.db.WithContext(ctx).Table("patients").
		Where("id = ? AND deleted_at IS NULL", id).Count(&n).Error
	return n > 0, err
}

func (g *GormRepository) UserRole(ctx context.Context, id uint) (string, bool, error) {
	var row struct {
		UserType string
		IsActive bool
	}
	err := g.db.WithContext(ctx).Table("users").Select("user_type, is_active").
		Where("id = ? AND deleted_at IS NULL", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	return row.UserType, row.IsActive, err
}
