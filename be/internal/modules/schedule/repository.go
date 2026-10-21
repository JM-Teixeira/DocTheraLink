package schedule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrNotFound: no such record in the user's clinic.
	ErrNotFound = errors.New("not found")
	// ErrFull: the physiotherapist's hour already has as many patients as
	// the division allows.
	ErrFull = errors.New("this hour is full for this physiotherapist")
	// ErrPatientBusy: the patient is already booked at that hour.
	ErrPatientBusy = errors.New("the patient already has an appointment at this hour")
)

// Repository is the persistence boundary of the module.
type Repository interface {
	DefaultClinic(ctx context.Context) (*string, error)
	Divisions(ctx context.Context, clinicID string) ([]Division, error)
	GetDivision(ctx context.Context, clinicID, id string) (*Division, error)
	SaveDivision(ctx context.Context, d *Division) error
	Therapists(ctx context.Context, clinicID string) ([]UserSummary, error)
	IsTherapist(ctx context.Context, clinicID string, userID uint) (bool, error)
	PatientInClinic(ctx context.Context, clinicID, patientID string) (bool, error)

	Shifts(ctx context.Context, clinicID string, date time.Time) ([]Shift, error)
	GetShift(ctx context.Context, clinicID, id string) (*Shift, error)
	SaveShift(ctx context.Context, s *Shift) error
	DeleteShift(ctx context.Context, id string) error

	Appointments(ctx context.Context, clinicID string, date time.Time) ([]Appointment, error)
	GetAppointment(ctx context.Context, clinicID, id string) (*Appointment, error)
	// Book inserts the appointment unless the hour already holds capacity
	// bookings or the patient is busy then; it is safe under concurrency.
	Book(ctx context.Context, a *Appointment, capacity int) error
	UpdateAppointment(ctx context.Context, id string, fields map[string]any) error
	DeleteAppointment(ctx context.Context, id string) error
}

// GormRepository implements Repository on PostgreSQL via GORM.
type GormRepository struct{ db *gorm.DB }

// NewGormRepository builds a repository on the given connection.
func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// DefaultClinic is used for administrators who work in no clinic: the
// first clinic by name (one clinic today, see decision D2).
func (g *GormRepository) DefaultClinic(ctx context.Context) (*string, error) {
	var id string
	err := g.db.WithContext(ctx).Table("facilities").Where("deleted_at IS NULL").
		Order("name").Limit(1).Pluck("id", &id).Error
	if err != nil || id == "" {
		return nil, err
	}
	return &id, nil
}

func (g *GormRepository) Divisions(ctx context.Context, clinicID string) ([]Division, error) {
	var ds []Division
	err := g.db.WithContext(ctx).Where("clinic_id = ? AND active", clinicID).
		Order("sort_order, name").Find(&ds).Error
	return ds, err
}

func (g *GormRepository) GetDivision(ctx context.Context, clinicID, id string) (*Division, error) {
	var d Division
	err := g.db.WithContext(ctx).Where("clinic_id = ?", clinicID).First(&d, "id = ?", id).Error
	return &d, notFound(err)
}

func (g *GormRepository) SaveDivision(ctx context.Context, d *Division) error {
	return g.db.WithContext(ctx).Save(d).Error
}

func (g *GormRepository) Therapists(ctx context.Context, clinicID string) ([]UserSummary, error) {
	var us []UserSummary
	err := g.db.WithContext(ctx).Where("home_facility_id = ? AND user_type = ? AND is_active AND deleted_at IS NULL",
		clinicID, "PhysioTherapist").Order("first_name, last_name").Find(&us).Error
	return us, err
}

func (g *GormRepository) IsTherapist(ctx context.Context, clinicID string, userID uint) (bool, error) {
	var n int64
	err := g.db.WithContext(ctx).Table("users").Where("id = ? AND home_facility_id = ? AND user_type = ? AND is_active AND deleted_at IS NULL",
		userID, clinicID, "PhysioTherapist").Count(&n).Error
	return n == 1, err
}

func (g *GormRepository) PatientInClinic(ctx context.Context, clinicID, patientID string) (bool, error) {
	var n int64
	err := g.db.WithContext(ctx).Table("patients").Where("id = ? AND facility_id = ? AND deleted_at IS NULL",
		patientID, clinicID).Count(&n).Error
	return n == 1, err
}

func (g *GormRepository) Shifts(ctx context.Context, clinicID string, date time.Time) ([]Shift, error) {
	var ss []Shift
	err := g.db.WithContext(ctx).Where("clinic_id = ? AND date = ?", clinicID, day(date)).
		Order("therapist_id, start_hour").Find(&ss).Error
	return ss, err
}

func (g *GormRepository) GetShift(ctx context.Context, clinicID, id string) (*Shift, error) {
	var s Shift
	err := g.db.WithContext(ctx).Where("clinic_id = ?", clinicID).First(&s, "id = ?", id).Error
	return &s, notFound(err)
}

func (g *GormRepository) SaveShift(ctx context.Context, s *Shift) error {
	return g.db.WithContext(ctx).Save(s).Error
}

func (g *GormRepository) DeleteShift(ctx context.Context, id string) error {
	return g.db.WithContext(ctx).Delete(&Shift{}, "id = ?", id).Error
}

func (g *GormRepository) Appointments(ctx context.Context, clinicID string, date time.Time) ([]Appointment, error) {
	var as []Appointment
	err := g.db.WithContext(ctx).Preload("Patient").Where("clinic_id = ? AND date = ?", clinicID, day(date)).
		Order("hour, created_at").Find(&as).Error
	return as, err
}

func (g *GormRepository) GetAppointment(ctx context.Context, clinicID, id string) (*Appointment, error) {
	var a Appointment
	err := g.db.WithContext(ctx).Where("clinic_id = ?", clinicID).First(&a, "id = ?", id).Error
	return &a, notFound(err)
}

// Book serialises bookings of the same physiotherapist-hour with a
// transaction-scoped advisory lock, then re-checks capacity and the patient.
func (g *GormRepository) Book(ctx context.Context, a *Appointment, capacity int) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		key := fmt.Sprintf("appt:%s:%d:%s:%d", a.ClinicID, a.TherapistID, day(a.Date), a.Hour)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", key).Error; err != nil {
			return err
		}
		patientKey := fmt.Sprintf("patient:%s:%s:%d", a.PatientID, day(a.Date), a.Hour)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", patientKey).Error; err != nil {
			return err
		}
		var booked, busy int64
		if err := tx.Model(&Appointment{}).Where("therapist_id = ? AND date = ? AND hour = ?",
			a.TherapistID, day(a.Date), a.Hour).Count(&booked).Error; err != nil {
			return err
		}
		if int(booked) >= capacity {
			return ErrFull
		}
		if err := tx.Model(&Appointment{}).Where("patient_id = ? AND date = ? AND hour = ?",
			a.PatientID, day(a.Date), a.Hour).Count(&busy).Error; err != nil {
			return err
		}
		if busy > 0 {
			return ErrPatientBusy
		}
		return tx.Omit("Patient").Create(a).Error
	})
}

func (g *GormRepository) UpdateAppointment(ctx context.Context, id string, fields map[string]any) error {
	return g.db.WithContext(ctx).Model(&Appointment{}).Where("id = ?", id).Updates(fields).Error
}

func (g *GormRepository) DeleteAppointment(ctx context.Context, id string) error {
	return g.db.WithContext(ctx).Delete(&Appointment{}, "id = ?", id).Error
}

// day formats a date for the SQL date columns.
func day(t time.Time) string { return t.Format(dateLayout) }
