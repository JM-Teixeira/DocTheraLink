package patient

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory Repository. It records the scope it was called
// with; visibility itself is enforced in SQL and covered by integration tests.
type fakeRepo struct {
	consents  []Consent
	patients  map[string]models.Patient
	lastScope Scope
	lastQuery ListQuery
	updates   map[string]any
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{patients: map[string]models.Patient{}}
}

func (f *fakeRepo) List(_ context.Context, s Scope, q ListQuery) ([]models.Patient, int64, error) {
	f.lastScope, f.lastQuery = s, q
	var out []models.Patient
	for _, p := range f.patients {
		out = append(out, p)
	}
	return out, int64(len(out)), nil
}

func (f *fakeRepo) Get(_ context.Context, s Scope, id string) (*models.Patient, error) {
	f.lastScope = s
	p, ok := f.patients[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &p, nil
}

// Consents returns the patient's records newest first (they are appended in time order).
func (f *fakeRepo) Consents(_ context.Context, patientID string) ([]Consent, error) {
	var out []Consent
	for i := len(f.consents) - 1; i >= 0; i-- {
		if f.consents[i].PatientID == patientID {
			out = append(out, f.consents[i])
		}
	}
	return out, nil
}

func (f *fakeRepo) AddConsent(_ context.Context, c *Consent) error {
	_ = c.BeforeCreate(nil)
	f.consents = append(f.consents, *c)
	return nil
}

func (f *fakeRepo) Create(_ context.Context, p *models.Patient) error {
	_ = p.BeforeCreate(nil)
	f.patients[p.ID] = *p
	return nil
}

func (f *fakeRepo) Update(_ context.Context, id string, fields map[string]any) error {
	if _, ok := f.patients[id]; !ok {
		return ErrNotFound
	}
	f.updates = fields
	return nil
}

func (f *fakeRepo) Stats(_ context.Context, s Scope, q ListQuery, _ time.Time) (*Stats, error) {
	f.lastScope, f.lastQuery = s, q
	return &Stats{Total: int64(len(f.patients))}, nil
}

// Candidates mirrors the SQL pre-filter of GormRepository.Candidates.
func (f *fakeRepo) Candidates(_ context.Context, _ Scope, q DuplicateQuery) ([]models.Patient, error) {
	var out []models.Patient
	for _, p := range f.patients {
		if (q.NIF != "" && p.NIF != nil && *p.NIF == q.NIF) ||
			(q.SNS != "" && p.SNSNumber != nil && *p.SNSNumber == q.SNS) ||
			(q.DateOfBirth != nil && p.DateOfBirth != nil && sameDay(*q.DateOfBirth, *p.DateOfBirth)) ||
			(q.PhoneKey != "" && phoneKey(p.PhoneNumber) == q.PhoneKey) {
			out = append(out, p)
		}
	}
	return out, nil
}

var (
	nurse        = Actor{ID: 5, Role: models.RoleNurse}
	receptionist = Actor{ID: 7, Role: models.RoleReceptionist}
	readOnly     = Actor{ID: 6, Role: models.RoleStaffReadOnly}
	admin        = Actor{ID: 1, Role: models.RoleStateAdmin}
	now          = time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

	home        = uuid.NewString() // clinic of nurse and receptionist
	otherClinic = uuid.NewString()
)

func newSvc() (*Service, *fakeRepo) {
	repo := newFakeRepo()
	clinics := clinic.Static{nurse.ID: home, receptionist.ID: home}
	return NewService(repo, clinics, func() time.Time { return now }), repo
}

func strp(s string) *string { return &s }

func TestCreateClinicRules(t *testing.T) {
	ctx := context.Background()

	t.Run("receptionist registers patients in their clinic", func(t *testing.T) {
		s, _ := newSvc()
		p, err := s.Create(ctx, receptionist, CreateRequest{Name: "Ana"})
		if err != nil || p.FacilityID == nil || *p.FacilityID != home {
			t.Fatalf("err=%v facility=%v", err, p)
		}
	})

	t.Run("non-admin cannot register into another clinic", func(t *testing.T) {
		s, _ := newSvc()
		_, err := s.Create(ctx, nurse, CreateRequest{Name: "Ana", FacilityID: &otherClinic})
		var v *ValidationError
		if !errors.As(err, &v) || v.Field != "facility" {
			t.Fatalf("want facility validation error, got %v", err)
		}
	})

	t.Run("admin may register into any clinic", func(t *testing.T) {
		s, _ := newSvc()
		p, err := s.Create(ctx, admin, CreateRequest{Name: "Ana", FacilityID: &otherClinic})
		if err != nil || *p.FacilityID != otherClinic {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestCreateSetsServerFields(t *testing.T) {
	s, _ := newSvc()

	p, err := s.Create(context.Background(), nurse, CreateRequest{
		Name: "  Ana Pereira ", DateOfBirth: strp("1980-05-02"), PhoneNumber: strp("+351 912 345 678"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Ana Pereira" || p.CreatedByID == nil || *p.CreatedByID != nurse.ID {
		t.Errorf("server fields: %+v", p)
	}
	if p.FacilityID == nil || *p.FacilityID != home {
		t.Error("facility should default to the user's home facility")
	}
	if p.YearOfBirth == nil || *p.YearOfBirth != 1980 {
		t.Error("year_of_birth should be derived from date_of_birth")
	}
}

func TestCreateValidation(t *testing.T) {
	cases := []struct {
		name  string
		actor Actor
		req   CreateRequest
		want  string // expected ValidationError field, or "" for ErrForbidden
	}{
		{"read-only role", readOnly, CreateRequest{Name: "Ana"}, ""},
		{"short name", nurse, CreateRequest{Name: "A"}, "name"},
		{"future birth date", nurse, CreateRequest{Name: "Ana", DateOfBirth: strp("2030-01-01")}, "date_of_birth"},
		{"bad date", nurse, CreateRequest{Name: "Ana", DateOfBirth: strp("02/05/1980")}, "date_of_birth"},
		{"bad phone", nurse, CreateRequest{Name: "Ana", PhoneNumber: strp("call me")}, "phone_number"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newSvc()
			_, err := s.Create(context.Background(), tc.actor, tc.req)
			if tc.want == "" {
				if !errors.Is(err, ErrForbidden) {
					t.Fatalf("want forbidden, got %v", err)
				}
				return
			}
			var v *ValidationError
			if !errors.As(err, &v) || v.Field != tc.want {
				t.Fatalf("want validation error on %s, got %v", tc.want, err)
			}
		})
	}
}

func TestScopeAndPagination(t *testing.T) {
	s, repo := newSvc()
	ctx := context.Background()

	_, _, _ = s.List(ctx, admin, ListQuery{Limit: 1000})
	if !repo.lastScope.All || repo.lastQuery.Limit != maxLimit {
		t.Errorf("admin scope/limit: %+v %+v", repo.lastScope, repo.lastQuery)
	}

	_, _, _ = s.List(ctx, nurse, ListQuery{Offset: -3})
	if repo.lastScope.All || repo.lastScope.UserID != nurse.ID || repo.lastQuery.Offset != 0 ||
		repo.lastScope.ClinicID == nil || *repo.lastScope.ClinicID != home {
		t.Errorf("nurse scope: %+v %+v", repo.lastScope, repo.lastQuery)
	}

	if _, err := s.Get(ctx, nurse, "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Errorf("invalid id should be not found, got %v", err)
	}
}

func TestUpdateWhitelistAndCoercion(t *testing.T) {
	s, _ := newSvc()
	ctx := context.Background()
	p, _ := s.Create(ctx, nurse, CreateRequest{Name: "Ana"})

	s2, repo := s, s.repo.(*fakeRepo)
	_, err := s2.Update(ctx, nurse, p.ID, map[string]any{
		"name":          "Ana Maria",
		"gender":        float64(2),
		"is_active":     false,
		"id":            "attacker-chosen", // ignored
		"created_by_id": float64(999),      // ignored
		"allergies":     nil,               // clears the field
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.updates["name"] != "Ana Maria" || repo.updates["gender"] != int64(2) || repo.updates["is_active"] != false {
		t.Errorf("coerced fields: %+v", repo.updates)
	}
	if _, ok := repo.updates["id"]; ok {
		t.Error("id must not be updatable")
	}
	if _, ok := repo.updates["allergies"]; !ok {
		t.Error("null should clear allergies")
	}
	if repo.updates["last_edited_id"] != nurse.ID {
		t.Error("last_edited_id not set")
	}

	bad := []map[string]any{
		{"gender": "female"},
		{"gender": 1.5},
		{"facility": "nope"},
		{"name": nil},
		{"name": " x "},
		{"phone_number": "abc"},
	}
	for _, body := range bad {
		var v *ValidationError
		if _, err := s.Update(ctx, nurse, p.ID, body); !errors.As(err, &v) {
			t.Errorf("%v: want validation error, got %v", body, err)
		}
	}

	if _, err := s.Update(ctx, readOnly, p.ID, map[string]any{"name": "Bob"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("read-only update: want forbidden, got %v", err)
	}
}

func TestEscapeLike(t *testing.T) {
	if got := escapeLike(`50%_off\`); got != `50\%\_off\\` {
		t.Errorf("escapeLike: %q", got)
	}
}

func TestRegistrationExtras(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()

	p, err := s.Create(ctx, receptionist, CreateRequest{
		Name: "Ana", Email: strp(" ana@example.pt "), PostalCode: strp("4000123"), City: strp(" Porto "),
		EmergencyContactName: strp("Rui"), EmergencyContactRelation: strp("spouse"),
		Insurer: strp("ADSE"), InsuranceNumber: strp(""),
	})
	if err != nil {
		t.Fatal(err)
	}
	if *p.PostalCode != "4000-123" || *p.City != "Porto" || *p.Email != "ana@example.pt" ||
		*p.EmergencyContactRelation != "spouse" || p.InsuranceNumber != nil {
		t.Errorf("extras not stored as expected: %+v", p)
	}

	for field, req := range map[string]CreateRequest{
		"email":                      {Name: "Ana", Email: strp("ana@")},
		"postal_code":                {Name: "Ana", PostalCode: strp("40001")},
		"emergency_contact_relation": {Name: "Ana", EmergencyContactRelation: strp("boss")},
	} {
		var v *ValidationError
		if _, err := s.Create(ctx, receptionist, req); !errors.As(err, &v) || v.Field != field {
			t.Errorf("%s: want validation error, got %v", field, err)
		}
	}

	// Edits: the postal code is normalised, empty clears, bad values fail
	if _, err := s.Update(ctx, receptionist, p.ID, map[string]any{
		"postal_code": "1000 001", "email": "", "emergency_contact_relation": "parent", "city": "Lisboa",
	}); err != nil {
		t.Fatal(err)
	}
	if repo.updates["postal_code"] != "1000-001" || repo.updates["email"] != nil ||
		repo.updates["emergency_contact_relation"] != "parent" || repo.updates["city"] != "Lisboa" {
		t.Errorf("updates: %+v", repo.updates)
	}
	for _, body := range []map[string]any{{"postal_code": "abc"}, {"email": "x"}, {"emergency_contact_relation": "boss"}} {
		var v *ValidationError
		if _, err := s.Update(ctx, receptionist, p.ID, body); !errors.As(err, &v) {
			t.Errorf("%v: want validation error, got %v", body, err)
		}
	}
}

func TestClinicalLists(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	since := 2015
	p, err := s.Create(ctx, nurse, CreateRequest{Name: "Ana", ClinicalLists: ClinicalLists{
		Allergies:   []models.Allergy{{Substance: " Penicilina ", Severity: "severe"}, {Substance: "  "}},
		Medications: []models.Medication{{Name: "Ibuprofeno", Dose: "400 mg"}},
		Conditions:  []models.Condition{{Name: "Diabetes tipo 2", Since: &since}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.AllergyList) != 1 || p.AllergyList[0].Substance != "Penicilina" || len(p.MedicationList) != 1 ||
		*p.ConditionList[0].Since != 2015 {
		t.Errorf("lists: %+v %+v %+v", p.AllergyList, p.MedicationList, p.ConditionList)
	}

	future := now.Year() + 1
	long := strings.Repeat("x", maxItemText+1)
	many := make([]models.Medication, maxListItems+1)
	for i := range many {
		many[i].Name = "m"
	}
	for name, l := range map[string]ClinicalLists{
		"bad severity": {Allergies: []models.Allergy{{Substance: "Pó", Severity: "deadly"}}},
		"future year":  {Conditions: []models.Condition{{Name: "Asma", Since: &future}}},
		"long text":    {Medications: []models.Medication{{Name: long}}},
		"too many":     {Medications: many},
	} {
		var v *ValidationError
		if _, err := s.Create(ctx, nurse, CreateRequest{Name: "Ana", ClinicalLists: l}); !errors.As(err, &v) {
			t.Errorf("%s: want validation error, got %v", name, err)
		}
	}

	// PATCH stores the cleaned list as JSON text; null clears; junk fails
	if _, err := s.Update(ctx, nurse, p.ID, map[string]any{
		"allergy_list":    []any{map[string]any{"substance": " Látex ", "reaction": "urticária"}, map[string]any{"substance": ""}},
		"medication_list": nil,
	}); err != nil {
		t.Fatal(err)
	}
	if repo.updates["allergy_list"] != `[{"substance":"Látex","reaction":"urticária"}]` {
		t.Errorf("stored allergies: %v", repo.updates["allergy_list"])
	}
	if v, ok := repo.updates["medication_list"]; !ok || v != nil {
		t.Errorf("null should clear medications: %v", v)
	}
	for _, body := range []map[string]any{
		{"allergy_list": "penicilina"},
		{"condition_list": []any{map[string]any{"name": "Asma", "since": "ontem"}}},
		{"allergy_list": []any{map[string]any{"substance": "Pó", "severity": "x"}}},
	} {
		var v *ValidationError
		if _, err := s.Update(ctx, nurse, p.ID, body); !errors.As(err, &v) {
			t.Errorf("%v: want validation error, got %v", body, err)
		}
	}
}

func TestMetricsFiltersAndExport(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	coord := Actor{ID: 9, Role: models.RoleCoordinator}
	physio := Actor{ID: 10, Role: models.RolePhysioTherapist}

	// Diagnosis is clinical: not for the front desk
	if _, _, err := s.List(ctx, receptionist, ListQuery{Diagnosis: "lombalgia"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("receptionist diagnosis filter: %v", err)
	}
	if _, err := s.Stats(ctx, nurse, ListQuery{Diagnosis: "x"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("nurse diagnosis stats: %v", err)
	}
	if _, _, err := s.List(ctx, receptionist, ListQuery{Physio: 3}); err != nil {
		t.Errorf("physio filter is fine for everyone: %v", err)
	}
	s2 := NewService(repo, clinic.Static{physio.ID: home, coord.ID: home}, func() time.Time { return now })
	if _, _, err := s2.List(ctx, physio, ListQuery{Diagnosis: "lombalgia"}); err != nil {
		t.Errorf("physio diagnosis filter: %v", err)
	}

	// Export: coordinators and admins only, up to maxExport rows
	if _, err := s2.Export(ctx, physio, ListQuery{}); !errors.Is(err, ErrForbidden) {
		t.Errorf("physio export: %v", err)
	}
	if _, err := s2.Export(ctx, coord, ListQuery{Limit: 3}); err != nil || repo.lastQuery.Limit != maxExport {
		t.Errorf("coordinator export: %v limit %d", err, repo.lastQuery.Limit)
	}
}

func TestWriteCSV(t *testing.T) {
	dob := time.Date(1960, 12, 31, 0, 0, 0, 0, time.UTC)
	g := 2
	var b strings.Builder
	err := WriteCSV(&b, []models.Patient{{
		Name: "=HYPERLINK(\"x\")", DateOfBirth: &dob, Gender: &g, CareLevel: 3,
		City: strp("Porto"), CreatedAt: now,
	}}, "pt", now)
	if err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.HasPrefix(out, "\uFEFFNome;Data de nascimento;") {
		t.Errorf("header: %q", out[:40])
	}
	if !strings.Contains(out, `"'=HYPERLINK(""x"")";1960-12-31;65;Feminino;3;;;;Porto;2026-10-20`) {
		t.Errorf("row: %q", out)
	}
	b.Reset()
	_ = WriteCSV(&b, nil, "en", now)
	if !strings.Contains(b.String(), "Name;Date of birth") {
		t.Errorf("english header: %q", b.String())
	}
}

func TestMonthSeries(t *testing.T) {
	first := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	got := monthSeries(first, []struct {
		Month string
		N     int64
	}{{"2025-11", 2}, {"2026-10", 5}})
	if len(got) != 12 || got[0] != (MonthCount{"2025-11", 2}) || got[1].Count != 0 || got[11] != (MonthCount{"2026-10", 5}) {
		t.Errorf("series: %+v", got)
	}
}

func TestConsents(t *testing.T) {
	ctx := context.Background()
	s, _ := newSvc()
	p, _ := s.Create(ctx, receptionist, CreateRequest{Name: "Ana"})
	yes, no := true, false

	v, err := s.Consents(ctx, receptionist, p.ID)
	if err != nil || len(v.Current) != 0 || v.History == nil {
		t.Fatalf("empty: %v %+v", err, v)
	}
	if _, err := s.RecordConsent(ctx, receptionist, p.ID, ConsentRequest{Kind: "data_processing", Granted: &yes, Method: "signed_paper"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordConsent(ctx, receptionist, p.ID, ConsentRequest{Kind: "contact_sms", Granted: &yes, Method: "verbal"}); err != nil {
		t.Fatal(err)
	}
	v, err = s.RecordConsent(ctx, receptionist, p.ID, ConsentRequest{Kind: "contact_sms", Granted: &no, Method: "verbal", Note: " pediu "})
	if err != nil {
		t.Fatal(err)
	}
	if len(v.History) != 3 || v.Current["contact_sms"].Granted || !v.Current["data_processing"].Granted ||
		v.Current["contact_sms"].Note != "pediu" || v.Current["contact_sms"].RecordedByID != receptionist.ID {
		t.Errorf("view: %+v", v)
	}

	for field, req := range map[string]ConsentRequest{
		"kind":   {Kind: "marketing", Granted: &yes, Method: "verbal"},
		"method": {Kind: "photos", Granted: &yes, Method: "telepathy"},
	} {
		var ve *ValidationError
		if _, err := s.RecordConsent(ctx, receptionist, p.ID, req); !errors.As(err, &ve) || ve.Field != field {
			t.Errorf("%s: %v", field, err)
		}
	}
	if _, err := s.RecordConsent(ctx, readOnly, p.ID, ConsentRequest{Kind: "photos", Granted: &yes, Method: "verbal"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("read-only: %v", err)
	}
	if _, err := s.Consents(ctx, receptionist, uuid.NewString()); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown patient: %v", err)
	}
}
