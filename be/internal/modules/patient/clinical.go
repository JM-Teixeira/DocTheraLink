package patient

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"klyniva/internal/models"
)

// Limits for the structured clinical lists.
const (
	maxListItems = 30
	maxItemText  = 120
)

var severities = map[string]bool{"": true, "mild": true, "moderate": true, "severe": true}

// ClinicalLists are the structured allergies, medications and conditions.
type ClinicalLists struct {
	Allergies   []models.Allergy    `json:"allergy_list"`
	Medications []models.Medication `json:"medication_list"`
	Conditions  []models.Condition  `json:"condition_list"`
}

// clean trims every text, drops empty rows and checks sizes and values.
// The year of a condition may not be before 1900 or after this year.
func (l *ClinicalLists) clean(now time.Time) error {
	var err error
	if l.Allergies, err = cleanList("allergy_list", l.Allergies, func(a *models.Allergy) (bool, error) {
		a.Substance, a.Reaction, a.Severity = trim(a.Substance), trim(a.Reaction), trim(a.Severity)
		if !severities[a.Severity] {
			return false, fmt.Errorf("severity must be mild, moderate or severe")
		}
		return a.Substance != "", textsOK(a.Substance, a.Reaction)
	}); err != nil {
		return err
	}
	if l.Medications, err = cleanList("medication_list", l.Medications, func(m *models.Medication) (bool, error) {
		m.Name, m.Dose, m.Frequency = trim(m.Name), trim(m.Dose), trim(m.Frequency)
		return m.Name != "", textsOK(m.Name, m.Dose, m.Frequency)
	}); err != nil {
		return err
	}
	l.Conditions, err = cleanList("condition_list", l.Conditions, func(c *models.Condition) (bool, error) {
		c.Name = trim(c.Name)
		if c.Since != nil && (*c.Since < 1900 || *c.Since > now.Year()) {
			return false, fmt.Errorf("year must be between 1900 and %d", now.Year())
		}
		return c.Name != "", textsOK(c.Name)
	})
	return err
}

// cleanList applies fix to each item, keeping those it says are filled in.
// A nil input stays nil (field not sent).
func cleanList[T any](field string, items []T, fix func(*T) (bool, error)) ([]T, error) {
	if items == nil {
		return nil, nil
	}
	out := []T{}
	for i := range items {
		keep, err := fix(&items[i])
		if err != nil {
			return nil, &ValidationError{field, err.Error()}
		}
		if keep {
			out = append(out, items[i])
		}
	}
	if len(out) > maxListItems {
		return nil, &ValidationError{field, fmt.Sprintf("at most %d items", maxListItems)}
	}
	return out, nil
}

func trim(s string) string { return strings.TrimSpace(s) }

func textsOK(texts ...string) error {
	for _, s := range texts {
		if utf8.RuneCountInString(s) > maxItemText {
			return fmt.Errorf("texts are limited to %d characters", maxItemText)
		}
	}
	return nil
}

// coerceList validates one list sent in a PATCH body and returns it as the
// JSON text stored in the column (null clears it).
func (s *Service) coerceList(key string, raw any) (any, error) {
	if raw == nil {
		return nil, nil
	}
	b, err := json.Marshal(map[string]any{key: raw})
	if err != nil {
		return nil, &ValidationError{key, "has an invalid value"}
	}
	var l ClinicalLists
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, &ValidationError{key, "has an invalid value"}
	}
	if err := l.clean(s.now()); err != nil {
		return nil, err
	}
	var v any
	switch key {
	case "allergy_list":
		v = l.Allergies
	case "medication_list":
		v = l.Medications
	default:
		v = l.Conditions
	}
	out, _ := json.Marshal(v)
	return string(out), nil
}
