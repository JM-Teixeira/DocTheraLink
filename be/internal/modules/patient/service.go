package patient

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/google/uuid"
)

// ErrForbidden means the actor's role may not perform the action.
var ErrForbidden = errors.New("not allowed to perform this action")

// ValidationError reports bad input with the offending field.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

// Actor is the authenticated user performing an action.
type Actor struct {
	ID   uint
	Role models.UserRole
}

const (
	defaultLimit = 20
	maxLimit     = 100
	dateLayout   = "2006-01-02"
)

var phoneRe = regexp.MustCompile(`^\+?[0-9 ()\-.]{6,20}$`)

// CreateRequest is the body of POST /patient/. Field names follow the
// existing frontend PatientModel (snake_case, CARE-compatible).
type CreateRequest struct {
	Name                  string  `json:"name" binding:"required,max=255"`
	Gender                *int    `json:"gender" binding:"omitempty,min=1,max=3"`
	DateOfBirth           *string `json:"date_of_birth"`
	YearOfBirth           *int    `json:"year_of_birth" binding:"omitempty,min=1900"`
	PhoneNumber           *string `json:"phone_number"`
	EmergencyPhoneNumber  *string `json:"emergency_phone_number"`
	BloodGroup            *string `json:"blood_group" binding:"omitempty,oneof=A+ A- B+ B- AB+ AB- O+ O-"`
	Address               *string `json:"address"`
	PermanentAddress      *string `json:"permanent_address"`
	Pincode               *int    `json:"pincode"`
	Nationality           *string `json:"nationality"`
	NIF                   *string `json:"nif"`
	SNSNumber             *string `json:"sns_number"`
	Allergies             *string `json:"allergies"`
	MedicalHistoryDetails *string `json:"medical_history_details"`
	OngoingMedication     *string `json:"ongoing_medication"`
	PresentHealth         *string `json:"present_health"`
	FacilityID            *string `json:"facility" binding:"omitempty,uuid"`
	// AllowDuplicate confirms the user checked the name-based possible
	// duplicates and this is a different person. Same NIF/SNS is never allowed.
	AllowDuplicate bool `json:"allow_duplicate"`
	// CareLevel 1–5 (default 1), see models.Patient.
	CareLevel *int `json:"care_level" binding:"omitempty,min=1,max=5"`
}

// Service holds the patient business rules.
type Service struct {
	repo    Repository
	clinics clinic.Resolver
	now     func() time.Time
}

// NewService builds a Service. now may be nil to use time.Now.
func NewService(repo Repository, clinics clinic.Resolver, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{repo: repo, clinics: clinics, now: now}
}

// scope resolves what the actor may see.
func (s *Service) scope(ctx context.Context, a Actor) (Scope, error) {
	if auth.IsAdmin(a.Role) {
		return Scope{All: true}, nil
	}
	home, err := s.clinics.HomeClinic(ctx, a.ID)
	if err != nil {
		return Scope{}, err
	}
	return Scope{UserID: a.ID, ClinicID: home}, nil
}

// List returns visible patients, newest first, with bounded pagination.
func (s *Service) List(ctx context.Context, a Actor, q ListQuery) ([]models.Patient, int64, error) {
	sc, err := s.scope(ctx, a)
	if err != nil {
		return nil, 0, err
	}
	if q.Limit <= 0 {
		q.Limit = defaultLimit
	}
	q.Limit = min(q.Limit, maxLimit)
	q.Offset = max(q.Offset, 0)
	return s.repo.List(ctx, sc, q)
}

// Get returns one visible patient.
func (s *Service) Get(ctx context.Context, a Actor, id string) (*models.Patient, error) {
	if uuid.Validate(id) != nil {
		return nil, ErrNotFound
	}
	sc, err := s.scope(ctx, a)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, sc, id)
}

// Create registers a patient. Read-only roles cannot write.
func (s *Service) Create(ctx context.Context, a Actor, req CreateRequest) (*models.Patient, error) {
	if !auth.CanModifyResource(a.Role) {
		return nil, ErrForbidden
	}
	name := strings.TrimSpace(req.Name)
	if len(name) < 2 {
		return nil, &ValidationError{"name", "must have at least 2 characters"}
	}
	if err := validatePhone("phone_number", req.PhoneNumber); err != nil {
		return nil, err
	}
	if err := validatePhone("emergency_phone_number", req.EmergencyPhoneNumber); err != nil {
		return nil, err
	}
	dob, err := s.parseBirthDate(req.DateOfBirth)
	if err != nil {
		return nil, err
	}
	nif, err := identifier("nif", req.NIF)
	if err != nil {
		return nil, err
	}
	sns, err := identifier("sns_number", req.SNSNumber)
	if err != nil {
		return nil, err
	}
	dupes, err := s.findDuplicates(ctx, a, DuplicateQuery{
		Name: name, DateOfBirth: dob, PhoneKey: phoneKey(req.PhoneNumber),
		NIF: deref(nif), SNS: deref(sns),
	})
	if err != nil {
		return nil, err
	}
	if de := (&DuplicateError{Matches: dupes}); len(dupes) > 0 && (de.Blocking() || !req.AllowDuplicate) {
		return nil, de
	}

	// Patients belong to the registering user's clinic so colleagues see them.
	// Non-admins cannot register patients into another clinic.
	home, err := s.clinics.HomeClinic(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	facility := req.FacilityID
	if facility == nil {
		facility = home
	} else if !auth.IsAdmin(a.Role) && !clinic.SameClinic(facility, home) {
		return nil, &ValidationError{"facility", "must be your own clinic"}
	}

	p := &models.Patient{
		Name:                  name,
		Gender:                req.Gender,
		DateOfBirth:           dob,
		YearOfBirth:           req.YearOfBirth,
		PhoneNumber:           req.PhoneNumber,
		EmergencyPhoneNumber:  req.EmergencyPhoneNumber,
		BloodGroup:            (*models.BloodGroup)(req.BloodGroup),
		Address:               req.Address,
		PermanentAddress:      req.PermanentAddress,
		Pincode:               req.Pincode,
		Nationality:           req.Nationality,
		NIF:                   nif,
		CareLevel:             careLevelOr(req.CareLevel),
		SNSNumber:             sns,
		Allergies:             req.Allergies,
		MedicalHistoryDetails: req.MedicalHistoryDetails,
		OngoingMedication:     req.OngoingMedication,
		PresentHealth:         req.PresentHealth,
		FacilityID:            facility,
		IsActive:              true,
		CreatedByID:           &a.ID,
		LastEditedID:          &a.ID,
	}
	if dob != nil && p.YearOfBirth == nil {
		y := dob.Year()
		p.YearOfBirth = &y
	}
	if err := s.repo.Create(ctx, p); err != nil {
		if errors.Is(err, ErrDuplicateIdentifier) {
			return nil, &DuplicateError{}
		}
		return nil, err
	}
	return s.Get(ctx, a, p.ID)
}

// fieldKind describes how an editable JSON field is validated.
type fieldKind int

const (
	kindString fieldKind = iota
	kindInt
	kindBool
	kindDate
	kindUUID
	kindPhone
	kindIdentifier
)

// editable maps JSON keys accepted by PUT/PATCH to their column and type.
// Anything not listed (ids, audit fields, relations) is rejected.
var editable = map[string]struct {
	column string
	kind   fieldKind
}{
	"name":                    {"name", kindString},
	"gender":                  {"gender", kindInt},
	"date_of_birth":           {"date_of_birth", kindDate},
	"year_of_birth":           {"year_of_birth", kindInt},
	"phone_number":            {"phone_number", kindPhone},
	"emergency_phone_number":  {"emergency_phone_number", kindPhone},
	"blood_group":             {"blood_group", kindString},
	"address":                 {"address", kindString},
	"permanent_address":       {"permanent_address", kindString},
	"pincode":                 {"pincode", kindInt},
	"nationality":             {"nationality", kindString},
	"nif":                     {"nif", kindIdentifier},
	"sns_number":              {"sns_number", kindIdentifier},
	"care_level":              {"care_level", kindInt},
	"allergies":               {"allergies", kindString},
	"medical_history_details": {"medical_history_details", kindString},
	"ongoing_medication":      {"ongoing_medication", kindString},
	"present_health":          {"present_health", kindString},
	"facility":                {"facility_id", kindUUID},
	"assigned_to":             {"assigned_to_id", kindInt},
	"is_active":               {"is_active", kindBool},
}

// Update applies a partial update from decoded JSON. Unknown keys are ignored
// so clients can send back a full object; server-controlled fields never change.
func (s *Service) Update(ctx context.Context, a Actor, id string, body map[string]any) (*models.Patient, error) {
	if !auth.CanModifyResource(a.Role) {
		return nil, ErrForbidden
	}
	if _, err := s.Get(ctx, a, id); err != nil {
		return nil, err
	}

	fields := map[string]any{}
	for key, raw := range body {
		spec, ok := editable[key]
		if !ok {
			continue
		}
		v, err := s.coerce(key, spec.kind, raw)
		if err != nil {
			return nil, err
		}
		fields[spec.column] = v
	}
	if v, ok := fields["care_level"]; ok {
		if lvl, isInt := v.(int64); !isInt || lvl < 1 || lvl > 5 {
			return nil, &ValidationError{"care_level", "must be between 1 and 5"}
		}
	}
	if name, ok := fields["name"].(string); ok && len(strings.TrimSpace(name)) < 2 {
		return nil, &ValidationError{"name", "must have at least 2 characters"}
	}
	if len(fields) == 0 {
		return s.Get(ctx, a, id)
	}
	// A new NIF / SNS number must not belong to another patient
	q := DuplicateQuery{ExcludeID: id}
	q.NIF, _ = fields["nif"].(string)
	q.SNS, _ = fields["sns_number"].(string)
	if q.NIF != "" || q.SNS != "" {
		dupes, err := s.findDuplicates(ctx, a, q)
		if err != nil {
			return nil, err
		}
		if len(dupes) > 0 {
			return nil, &DuplicateError{Matches: dupes}
		}
	}
	fields["last_edited_id"] = a.ID
	if err := s.repo.Update(ctx, id, fields); err != nil {
		if errors.Is(err, ErrDuplicateIdentifier) {
			return nil, &DuplicateError{}
		}
		return nil, err
	}
	return s.Get(ctx, a, id)
}

// coerce validates one JSON value against its kind. JSON null clears the field.
func (s *Service) coerce(key string, kind fieldKind, raw any) (any, error) {
	if raw == nil {
		if key == "name" || key == "is_active" {
			return nil, &ValidationError{key, "cannot be null"}
		}
		return nil, nil
	}
	bad := &ValidationError{key, "has an invalid value"}
	switch kind {
	case kindString:
		str, ok := raw.(string)
		if !ok {
			return nil, bad
		}
		return strings.TrimSpace(str), nil
	case kindInt:
		f, ok := raw.(float64) // encoding/json decodes numbers as float64
		if !ok || f != float64(int64(f)) {
			return nil, bad
		}
		return int64(f), nil
	case kindBool:
		b, ok := raw.(bool)
		if !ok {
			return nil, bad
		}
		return b, nil
	case kindDate:
		str, ok := raw.(string)
		if !ok {
			return nil, bad
		}
		return s.parseBirthDate(&str)
	case kindUUID:
		str, ok := raw.(string)
		if !ok || uuid.Validate(str) != nil {
			return nil, bad
		}
		return str, nil
	case kindIdentifier:
		str, ok := raw.(string)
		if !ok {
			return nil, bad
		}
		v, err := identifier(key, &str)
		if err != nil || v == nil {
			return nil, err // empty string clears the field
		}
		return *v, nil
	case kindPhone:
		str, ok := raw.(string)
		if !ok {
			return nil, bad
		}
		if err := validatePhone(key, &str); err != nil {
			return nil, err
		}
		return str, nil
	}
	return nil, bad
}

// parseBirthDate accepts YYYY-MM-DD (or a full RFC 3339 timestamp) not in the future.
func (s *Service) parseBirthDate(v *string) (*time.Time, error) {
	if v == nil || *v == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, *v)
	if err != nil {
		if t, err = time.Parse(time.RFC3339, *v); err != nil {
			return nil, &ValidationError{"date_of_birth", fmt.Sprintf("must be a date in %s format", dateLayout)}
		}
	}
	if t.After(s.now()) {
		return nil, &ValidationError{"date_of_birth", "cannot be in the future"}
	}
	return &t, nil
}

func careLevelOr(v *int) int {
	if v == nil {
		return 1
	}
	return *v
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func validatePhone(field string, v *string) error {
	if v == nil || *v == "" {
		return nil
	}
	if !phoneRe.MatchString(*v) {
		return &ValidationError{field, "is not a valid phone number"}
	}
	return nil
}
