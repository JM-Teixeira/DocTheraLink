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
	// ErrOverCredits: the patients' care levels would exceed the hour's credits.
	ErrOverCredits = errors.New("this patient needs more care than is left in this hour")
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
	// PatientCareLevel returns the patient's care level; ok is false when
	// the patient is not in the clinic.
	PatientCareLevel(ctx context.Context, clinicID, patientID string) (level int, ok bool, err error)

	Shifts(ctx context.Context, clinicID string, from, to time.Time) ([]Shift, error) // inclusive dates
	GetShift(ctx context.Context, clinicID, id string) (*Shift, error)
	SaveShift(ctx context.Context, s *Shift) error
	DeleteShift(ctx context.Context, id string) error

	Appointments(ctx context.Context, clinicID string, from, to time.Time) ([]Appointment, error) // inclusive dates
	GetAppointment(ctx context.Context, clinicID, id string) (*Appointment, error)
	// Book inserts the appointment unless the hour already holds capacity
	// patients or the credits would exceed creditLimit, or the patient is
	// busy then; it is safe under concurrency.
	Book(ctx context.Context, a *Appointment, capacity, creditLimit int) error
	UpdateAppointment(ctx context.Context, id string, fields map[string]any) error
	DeleteAppointment(ctx context.Context, id string) error
	// DeleteSeriesFrom removes the still-booked appointments of a series
	// from a date on, returning how many.
	DeleteSeriesFrom(ctx context.Context, clinicID, seriesID string, from time.Time) (int64, error)

	MonthlyStats(ctx context.Context, clinicID string, year int) ([]MonthStats, error)

	// MoveAppointment gives a booked session to another physio at the same
	// hour, under the same limits and lock as Book.
	MoveAppointment(ctx context.Context, a *Appointment, to uint, capacity, creditLimit int) error
	CreateAbsence(ctx context.Context, a *Absence) error
	// Absences overlapping from..to, with handover notes, therapist and cover.
	Absences(ctx context.Context, clinicID string, from, to time.Time) ([]Absence, error)
	DeleteAbsence(ctx context.Context, clinicID, id string) error
	// CoveredBy lists the physios a cover physio is covering on a date.
	CoveredBy(ctx context.Context, coverID uint, at time.Time) ([]uint, error)
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

func (g *GormRepository) PatientCareLevel(ctx context.Context, clinicID, patientID string) (int, bool, error) {
	var levels []int
	err := g.db.WithContext(ctx).Table("patients").Where("id = ? AND facility_id = ? AND deleted_at IS NULL",
		patientID, clinicID).Pluck("care_level", &levels).Error
	if err != nil || len(levels) == 0 {
		return 0, false, err
	}
	return max(levels[0], 1), true, nil
}

func (g *GormRepository) Shifts(ctx context.Context, clinicID string, from, to time.Time) ([]Shift, error) {
	var ss []Shift
	err := g.db.WithContext(ctx).Where("clinic_id = ? AND date BETWEEN ? AND ?", clinicID, day(from), day(to)).
		Order("date, therapist_id, start_hour").Find(&ss).Error
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

func (g *GormRepository) Appointments(ctx context.Context, clinicID string, from, to time.Time) ([]Appointment, error) {
	var as []Appointment
	err := g.db.WithContext(ctx).Preload("Patient").
		Where("clinic_id = ? AND date BETWEEN ? AND ?", clinicID, day(from), day(to)).
		Order("date, hour, created_at").Find(&as).Error
	return as, err
}

func (g *GormRepository) GetAppointment(ctx context.Context, clinicID, id string) (*Appointment, error) {
	var a Appointment
	err := g.db.WithContext(ctx).Where("clinic_id = ?", clinicID).First(&a, "id = ?", id).Error
	return &a, notFound(err)
}

// Book serialises bookings of the same physiotherapist-hour (and of the
// same patient-hour) with transaction-scoped advisory locks, then re-checks
// patients, credits and the patient.
func (g *GormRepository) Book(ctx context.Context, a *Appointment, capacity, creditLimit int) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, key := range []string{
			fmt.Sprintf("appt:%s:%d:%s:%d", a.ClinicID, a.TherapistID, day(a.Date), a.Hour),
			fmt.Sprintf("patient:%s:%s:%d", a.PatientID, day(a.Date), a.Hour),
		} {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", key).Error; err != nil {
				return err
			}
		}
		var load struct {
			N       int
			Credits int
		}
		if err := tx.Model(&Appointment{}).Select("COUNT(*) AS n, COALESCE(SUM(credits), 0) AS credits").
			Where("therapist_id = ? AND date = ? AND hour = ?", a.TherapistID, day(a.Date), a.Hour).
			Scan(&load).Error; err != nil {
			return err
		}
		if load.N >= capacity {
			return ErrFull
		}
		if load.Credits+a.Credits > creditLimit {
			return ErrOverCredits
		}
		var busy int64
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

func (g *GormRepository) DeleteSeriesFrom(ctx context.Context, clinicID, seriesID string, from time.Time) (int64, error) {
	res := g.db.WithContext(ctx).Where("clinic_id = ? AND series_id = ? AND date >= ? AND status = ?",
		clinicID, seriesID, day(from), StatusBooked).Delete(&Appointment{})
	return res.RowsAffected, res.Error
}

func (g *GormRepository) MoveAppointment(ctx context.Context, a *Appointment, to uint, capacity, creditLimit int) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		key := fmt.Sprintf("appt:%s:%d:%s:%d", a.ClinicID, to, day(a.Date), a.Hour)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", key).Error; err != nil {
			return err
		}
		var load struct {
			N       int
			Credits int
		}
		if err := tx.Model(&Appointment{}).Select("COUNT(*) AS n, COALESCE(SUM(credits), 0) AS credits").
			Where("therapist_id = ? AND date = ? AND hour = ?", to, day(a.Date), a.Hour).Scan(&load).Error; err != nil {
			return err
		}
		switch {
		case load.N >= capacity:
			return ErrFull
		case load.Credits+a.Credits > creditLimit:
			return ErrOverCredits
		}
		return tx.Model(&Appointment{}).Where("id = ?", a.ID).Update("therapist_id", to).Error
	})
}

func (g *GormRepository) CreateAbsence(ctx context.Context, a *Absence) error {
	return g.db.WithContext(ctx).Omit("Therapist", "Cover", "Notes.Patient").Create(a).Error
}

func (g *GormRepository) Absences(ctx context.Context, clinicID string, from, to time.Time) ([]Absence, error) {
	out := []Absence{}
	err := g.db.WithContext(ctx).Preload("Notes.Patient").Preload("Therapist").Preload("Cover").
		Where("clinic_id = ? AND \"from\" <= ? AND \"to\" >= ?", clinicID, day(to), day(from)).
		Order("\"from\"").Find(&out).Error
	return out, err
}

func (g *GormRepository) DeleteAbsence(ctx context.Context, clinicID, id string) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("clinic_id = ? AND id = ?", clinicID, id).Delete(&Absence{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return tx.Where("absence_id = ?", id).Delete(&HandoverNote{}).Error
	})
}

func (g *GormRepository) CoveredBy(ctx context.Context, coverID uint, at time.Time) ([]uint, error) {
	var ids []uint
	err := g.db.WithContext(ctx).Model(&Absence{}).Where("cover_id = ? AND \"from\" <= ? AND \"to\" >= ?", coverID, day(at), day(at)).
		Pluck("therapist_id", &ids).Error
	return ids, err
}

// MonthlyStats aggregates a year of sessions (appointments) and referrals.
func (g *GormRepository) MonthlyStats(ctx context.Context, clinicID string, year int) ([]MonthStats, error) {
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(1, 0, 0)
	stats := make([]MonthStats, 12)
	for i := range stats {
		stats[i].Month = i + 1
	}

	var sessions []struct {
		Month            int
		Booked, Attended int
		Missed, Patients int
	}
	err := g.db.WithContext(ctx).Model(&Appointment{}).
		Select(`EXTRACT(MONTH FROM date)::int AS month,
			COUNT(*) FILTER (WHERE status = 'booked') AS booked,
			COUNT(*) FILTER (WHERE status = 'attended') AS attended,
			COUNT(*) FILTER (WHERE status = 'missed') AS missed,
			COUNT(DISTINCT patient_id) AS patients`).
		Where("clinic_id = ? AND date >= ? AND date < ?", clinicID, day(from), day(to)).
		Group("month").Scan(&sessions).Error
	if err != nil {
		return nil, err
	}
	for _, s := range sessions {
		m := &stats[s.Month-1]
		m.Booked, m.Attended, m.Missed, m.Patients = s.Booked, s.Attended, s.Missed, s.Patients
	}

	var refs []struct{ Month, Referred, Discharged int }
	err = g.db.WithContext(ctx).Raw(`
		SELECT m AS month, SUM(referred) AS referred, SUM(discharged) AS discharged FROM (
			SELECT EXTRACT(MONTH FROM created_at)::int AS m, 1 AS referred, 0 AS discharged
			  FROM referrals WHERE deleted_at IS NULL AND facility_id = ? AND created_at >= ? AND created_at < ?
			UNION ALL
			SELECT EXTRACT(MONTH FROM completed_at)::int, 0, 1
			  FROM referrals WHERE deleted_at IS NULL AND facility_id = ? AND status = 'completed'
			   AND completed_at >= ? AND completed_at < ?
		) x GROUP BY m`, clinicID, from, to, clinicID, from, to).Scan(&refs).Error
	if err != nil {
		return nil, err
	}
	for _, r := range refs {
		stats[r.Month-1].Referred, stats[r.Month-1].Discharged = r.Referred, r.Discharged
	}
	return stats, nil
}

// day formats a date for the SQL date columns.
func day(t time.Time) string { return t.Format(dateLayout) }
