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
	"time"

	"klyniva/internal/models"
	"klyniva/internal/textutil"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a patient does not exist or is not visible.
var ErrNotFound = errors.New("patient not found")

// ErrDuplicateIdentifier is returned when the clinic already has a patient
// with the same NIF or SNS number (unique index; races the service check).
var ErrDuplicateIdentifier = errors.New("a patient with this identifier already exists")

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
	AgeMin     *int   `form:"age_min"`
	AgeMax     *int   `form:"age_max"`
	CareLevels []int  `form:"care_level"` // any of these
	Gender     *int   `form:"gender"`
	// Patients a physio has (a referral or a session with them)
	Physio uint `form:"physio"`
	// Referral diagnosis or ICD code contains this (clinical: see Service.checkFilters)
	Diagnosis string `form:"diagnosis"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

// Stats summarises a set of patients.
type Stats struct {
	Total      int64            `json:"total"`
	CareLevels map[int]int64    `json:"care_levels"` // 1–5
	Genders    map[int]int64    `json:"genders"`     // 0 = unknown, 1–3
	AgeBands   map[string]int64 `json:"age_bands"`   // 0-17, 18-39, 40-64, 65-79, 80+, unknown
	AverageAge float64          `json:"average_age"`
	// Registrations per month, oldest first, the last 12 months (empty months included)
	Monthly []MonthCount `json:"monthly"`
}

// MonthCount is how many of the patients were registered in a month.
type MonthCount struct {
	Month string `json:"month"` // YYYY-MM
	Count int64  `json:"count"`
}

// Repository is the persistence boundary of the module.
type Repository interface {
	List(ctx context.Context, s Scope, q ListQuery) ([]models.Patient, int64, error)
	Stats(ctx context.Context, s Scope, q ListQuery, now time.Time) (*Stats, error)
	Get(ctx context.Context, s Scope, id string) (*models.Patient, error)
	Create(ctx context.Context, p *models.Patient) error
	Update(ctx context.Context, id string, fields map[string]any) error
	// Candidates returns visible patients sharing an identifier, the birth
	// date or the phone number with q (at most 50); see classify.
	Candidates(ctx context.Context, s Scope, q DuplicateQuery) ([]models.Patient, error)
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
	db := filtered(scoped(g.db.WithContext(ctx).Model(&models.Patient{}), s), q, time.Now())
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []models.Patient
	err := db.Order("created_at DESC").Limit(q.Limit).Offset(q.Offset).Find(&items).Error
	return items, total, err
}

// Stats summarises the patients matching q: total, care levels, gender and
// age bands (for the patients page's metrics).
func (g *GormRepository) Stats(ctx context.Context, s Scope, q ListQuery, now time.Time) (*Stats, error) {
	base := func() *gorm.DB { return filtered(scoped(g.db.WithContext(ctx).Model(&models.Patient{}), s), q, now) }
	st := &Stats{CareLevels: map[int]int64{}, Genders: map[int]int64{}, AgeBands: map[string]int64{}}
	if err := base().Count(&st.Total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		Key int
		N   int64
	}
	if err := base().Select("care_level AS key, COUNT(*) AS n").Group("care_level").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		st.CareLevels[r.Key] = r.N
	}
	rows = nil
	if err := base().Select("COALESCE(gender, 0) AS key, COUNT(*) AS n").Group("COALESCE(gender, 0)").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		st.Genders[r.Key] = r.N
	}
	var bands []struct {
		Band string
		N    int64
	}
	if err := base().Select(`CASE
			WHEN date_of_birth IS NULL THEN 'unknown'
			WHEN date_of_birth > ? THEN '0-17'
			WHEN date_of_birth > ? THEN '18-39'
			WHEN date_of_birth > ? THEN '40-64'
			WHEN date_of_birth > ? THEN '65-79'
			ELSE '80+' END AS band, COUNT(*) AS n`,
		now.AddDate(-18, 0, 0), now.AddDate(-40, 0, 0), now.AddDate(-65, 0, 0), now.AddDate(-80, 0, 0)).
		Group("band").Scan(&bands).Error; err != nil {
		return nil, err
	}
	for _, b := range bands {
		st.AgeBands[b.Band] = b.N
	}
	var avg *float64
	if err := base().Select("AVG(EXTRACT(YEAR FROM age(?, date_of_birth)))", now).Scan(&avg).Error; err != nil {
		return nil, err
	}
	if avg != nil {
		st.AverageAge = *avg
	}
	// Registrations per month over the last 12 months
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -11, 0)
	var months []struct {
		Month string
		N     int64
	}
	if err := base().Where("patients.created_at >= ?", first).
		Select("to_char(patients.created_at, 'YYYY-MM') AS month, COUNT(*) AS n").
		Group("month").Scan(&months).Error; err != nil {
		return nil, err
	}
	st.Monthly = monthSeries(first, months)
	return st, nil
}

// monthSeries lists the 12 months from first, with the counts found (0 if none).
func monthSeries(first time.Time, found []struct {
	Month string
	N     int64
}) []MonthCount {
	byMonth := map[string]int64{}
	for _, m := range found {
		byMonth[m.Month] = m.N
	}
	out := make([]MonthCount, 12)
	for i := range out {
		m := first.AddDate(0, i, 0).Format("2006-01")
		out[i] = MonthCount{Month: m, Count: byMonth[m]}
	}
	return out
}

// filtered applies the list filters (search, phone, clinic, age, care level, gender).
func filtered(db *gorm.DB, q ListQuery, now time.Time) *gorm.DB {
	if term := strings.TrimSpace(q.Search); term != "" {
		// Digits search the NIF and SNS number (from the start); words the name
		if id := normalizeID(term); len(id) >= 3 && textutil.Digits(id) == id {
			db = db.Where("patients.nif LIKE ? OR patients.sns_number LIKE ?", id+"%", id+"%")
		} else {
			db = db.Where("name ILIKE ?", "%"+escapeLike(term)+"%")
		}
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
	// Age from the date of birth: age_min 65 → born on or before 65 years ago
	if q.AgeMin != nil {
		db = db.Where("date_of_birth <= ?", now.AddDate(-*q.AgeMin, 0, 0))
	}
	if q.AgeMax != nil {
		db = db.Where("date_of_birth > ?", now.AddDate(-*q.AgeMax-1, 0, 0))
	}
	if len(q.CareLevels) > 0 {
		db = db.Where("care_level IN ?", q.CareLevels)
	}
	if q.Gender != nil {
		db = db.Where("gender = ?", *q.Gender)
	}
	if q.Physio != 0 {
		db = db.Where(`EXISTS (SELECT 1 FROM referrals r WHERE r.patient_id = patients.id AND r.therapist_id = ? AND r.deleted_at IS NULL)
			OR EXISTS (SELECT 1 FROM appointments a WHERE a.patient_id = patients.id AND a.therapist_id = ?)`, q.Physio, q.Physio)
	}
	if d := strings.TrimSpace(q.Diagnosis); d != "" {
		db = db.Where(`EXISTS (SELECT 1 FROM referrals r WHERE r.patient_id = patients.id AND r.deleted_at IS NULL
			AND (r.diagnosis ILIKE ? OR r.icd_code ILIKE ?))`, "%"+escapeLike(d)+"%", escapeLike(d)+"%")
	}
	return db
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
	err := g.db.WithContext(ctx).Omit("Facility", "State", "District", "LocalBody", "Ward",
		"AssignedTo", "CreatedBy", "LastEdited").Create(p).Error
	return translate(err)
}

func (g *GormRepository) Update(ctx context.Context, id string, fields map[string]any) error {
	res := g.db.WithContext(ctx).Model(&models.Patient{}).Where("id = ?", id).Updates(fields)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (g *GormRepository) Candidates(ctx context.Context, s Scope, q DuplicateQuery) ([]models.Patient, error) {
	fresh := g.db.Session(&gorm.Session{NewDB: true})
	cond := fresh.Where("1 = 0")
	if q.NIF != "" {
		cond = cond.Or("patients.nif = ?", q.NIF)
	}
	if q.SNS != "" {
		cond = cond.Or("patients.sns_number = ?", q.SNS)
	}
	if q.DateOfBirth != nil {
		cond = cond.Or("(patients.date_of_birth AT TIME ZONE 'UTC')::date = ?", q.DateOfBirth.UTC().Format(dateLayout))
	}
	if q.PhoneKey != "" {
		cond = cond.Or("RIGHT(regexp_replace(COALESCE(patients.phone_number, ''), '[^0-9]', '', 'g'), 9) = ?", q.PhoneKey)
	}
	var out []models.Patient
	err := scoped(g.db.WithContext(ctx).Model(&models.Patient{}), s).
		Where(cond).Order("created_at DESC").Limit(50).Find(&out).Error
	return out, err
}

// translate maps unique-index violations to ErrDuplicateIdentifier.
func translate(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return ErrDuplicateIdentifier
	}
	return err
}

// UniqueIndexes are created after AutoMigrate: one NIF / SNS number per
// clinic (partial, so patients without them and deleted ones don't count).
var UniqueIndexes = []string{
	`CREATE UNIQUE INDEX IF NOT EXISTS uniq_patients_clinic_nif ON patients (facility_id, nif) WHERE nif IS NOT NULL AND deleted_at IS NULL`,
	`CREATE UNIQUE INDEX IF NOT EXISTS uniq_patients_clinic_sns ON patients (facility_id, sns_number) WHERE sns_number IS NOT NULL AND deleted_at IS NULL`,
}

// escapeLike escapes LIKE wildcards so user input is matched literally.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
