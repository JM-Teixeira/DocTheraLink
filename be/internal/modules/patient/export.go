package patient

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/models"
)

// maxExport bounds one CSV export.
const maxExport = 5000

// Export returns the patients matching q for a CSV file. Only coordinators
// and admins export (it is personal data); the scope is the same as the list.
func (s *Service) Export(ctx context.Context, a Actor, q ListQuery) ([]models.Patient, error) {
	if a.Role != models.RoleCoordinator && !auth.IsAdmin(a.Role) {
		return nil, ErrForbidden
	}
	if err := checkFilters(a, q); err != nil {
		return nil, err
	}
	sc, err := s.scope(ctx, a)
	if err != nil {
		return nil, err
	}
	q.Limit, q.Offset = maxExport, 0
	items, _, err := s.repo.List(ctx, sc, q)
	return items, err
}

// exportHeaders per language; the columns are contact and demographic data
// only (no identifiers or clinical fields).
var exportHeaders = map[string][]string{
	"pt": {"Nome", "Data de nascimento", "Idade", "Sexo", "Nível de cuidados", "Telefone", "E-mail", "Código postal", "Localidade", "Registado em"},
	"en": {"Name", "Date of birth", "Age", "Gender", "Care level", "Phone", "E-mail", "Postal code", "City", "Registered on"},
}

var genderNames = map[string][]string{
	"pt": {"", "Masculino", "Feminino", "Outro"},
	"en": {"", "Male", "Female", "Other"},
}

// WriteCSV writes the patients as CSV for spreadsheets: UTF-8 with BOM and
// ";" (what Excel expects in Portugal). Cells that a spreadsheet would run
// as a formula are prefixed with a quote.
func WriteCSV(w io.Writer, patients []models.Patient, lang string, now time.Time) error {
	if _, ok := exportHeaders[lang]; !ok {
		lang = "pt"
	}
	if _, err := io.WriteString(w, "\uFEFF"); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	if err := cw.Write(exportHeaders[lang]); err != nil {
		return err
	}
	for _, p := range patients {
		dob, age := "", ""
		if p.DateOfBirth != nil {
			dob = p.DateOfBirth.Format(dateLayout)
			age = strconv.Itoa(yearsBetween(*p.DateOfBirth, now))
		}
		gender := ""
		if p.Gender != nil && *p.Gender >= 1 && *p.Gender <= 3 {
			gender = genderNames[lang][*p.Gender]
		}
		row := []string{p.Name, dob, age, gender, strconv.Itoa(p.CareLevel), deref(p.PhoneNumber),
			deref(p.Email), deref(p.PostalCode), deref(p.City), p.CreatedAt.Format(dateLayout)}
		for i := range row {
			row[i] = safeCell(row[i])
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// safeCell stops spreadsheets from treating text as a formula (CSV injection).
func safeCell(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

func yearsBetween(from, to time.Time) int {
	y := to.Year() - from.Year()
	if to.YearDay() < from.YearDay() {
		y--
	}
	return y
}
