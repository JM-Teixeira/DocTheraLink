package patient

import (
	"context"
	"fmt"
	"time"

	"klyniva/internal/models"
)

// Why a patient looks like a duplicate.
const (
	MatchNIF       = "nif"        // same tax number: certainly the same person
	MatchSNS       = "sns_number" // same SNS number: certainly the same person
	MatchNameBirth = "name_dob"   // similar name and same date of birth
	MatchNamePhone = "name_phone" // similar name and same phone number
)

// DuplicateQuery describes the patient being registered or edited.
type DuplicateQuery struct {
	Name        string
	DateOfBirth *time.Time
	PhoneKey    string // see phoneKey
	NIF         string
	SNS         string
	ExcludeID   string // the patient being edited
}

// Duplicate is a summary of an existing patient that may be the same person.
// Only visible patients are returned, so no data leaks across clinics.
type Duplicate struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	PhoneNumber *string    `json:"phone_number"`
	Reason      string     `json:"reason"`
	// Blocking matches (same NIF/SNS) cannot be overridden; the others can
	// be confirmed with allow_duplicate.
	Blocking bool `json:"blocking"`
}

// DuplicateError is returned when registering (or editing) would duplicate
// an existing patient. The handler answers 409 with the matches.
type DuplicateError struct{ Matches []Duplicate }

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("possible duplicate of %d existing patient(s)", len(e.Matches))
}

// Blocking reports whether any match cannot be overridden.
func (e *DuplicateError) Blocking() bool {
	for _, m := range e.Matches {
		if m.Blocking {
			return true
		}
	}
	return false
}

// classify decides whether candidate p matches q, and why. Strong
// identifiers win over name-based matches.
func classify(q DuplicateQuery, p models.Patient) (Duplicate, bool) {
	d := Duplicate{ID: p.ID, Name: p.Name, DateOfBirth: p.DateOfBirth, PhoneNumber: p.PhoneNumber}
	switch {
	case q.NIF != "" && p.NIF != nil && *p.NIF == q.NIF:
		d.Reason, d.Blocking = MatchNIF, true
	case q.SNS != "" && p.SNSNumber != nil && *p.SNSNumber == q.SNS:
		d.Reason, d.Blocking = MatchSNS, true
	case !similarNames(q.Name, p.Name):
		return d, false
	case q.DateOfBirth != nil && p.DateOfBirth != nil && sameDay(*q.DateOfBirth, *p.DateOfBirth):
		d.Reason = MatchNameBirth
	case q.PhoneKey != "" && phoneKey(p.PhoneNumber) == q.PhoneKey:
		d.Reason = MatchNamePhone
	default:
		return d, false
	}
	return d, true
}

func sameDay(a, b time.Time) bool {
	return a.UTC().Format(dateLayout) == b.UTC().Format(dateLayout)
}

// findDuplicates returns the visible patients matching q, blocking ones first.
func (s *Service) findDuplicates(ctx context.Context, a Actor, q DuplicateQuery) ([]Duplicate, error) {
	if q.NIF == "" && q.SNS == "" && q.DateOfBirth == nil && q.PhoneKey == "" {
		return nil, nil
	}
	sc, err := s.scope(ctx, a)
	if err != nil {
		return nil, err
	}
	candidates, err := s.repo.Candidates(ctx, sc, q)
	if err != nil {
		return nil, err
	}
	var blocking, others []Duplicate
	for _, p := range candidates {
		if p.ID == q.ExcludeID {
			continue
		}
		if d, ok := classify(q, p); ok && d.Blocking {
			blocking = append(blocking, d)
		} else if ok {
			others = append(others, d)
		}
	}
	return append(blocking, others...), nil
}
