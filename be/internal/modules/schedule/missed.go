package schedule

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
)

// Missed sessions: when a patient doesn't come, the session is marked
// missed and the miss is justified (ill, warned in advance) or not. Patients
// with MissFlagThreshold or more unjustified misses in the last
// MissWindowDays are flagged to the coordinator.
const (
	MissFlagThreshold = 3
	MissWindowDays    = 90
)

// MissCount is how often a patient missed sessions in the window.
type MissCount struct {
	PatientID   string          `json:"patient_id"`
	Patient     *PatientSummary `json:"patient,omitempty" gorm:"-"`
	Unjustified int             `json:"unjustified"`
	Justified   int             `json:"justified"`
	Last        time.Time       `json:"last"` // date of the latest miss
}

// MissReport is the coordinator's list of patients who miss too often.
type MissReport struct {
	Threshold  int         `json:"threshold"`
	WindowDays int         `json:"window_days"`
	Patients   []MissCount `json:"patients"`
}

// missWindow is the first day counted.
func (s *Service) missWindow() time.Time { return s.today().AddDate(0, 0, -MissWindowDays) }

// MissedFlags lists the clinic's patients with too many unjustified misses,
// worst first (coordinators and admins).
func (s *Service) MissedFlags(ctx context.Context, a Actor) (*MissReport, error) {
	if !canManage(a) {
		return nil, ErrForbidden
	}
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	counts, err := s.repo.MissedCounts(ctx, cid, s.missWindow(), "")
	if err != nil {
		return nil, err
	}
	out := []MissCount{}
	for _, c := range counts {
		if c.Unjustified >= MissFlagThreshold {
			out = append(out, c)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Unjustified != out[j].Unjustified {
			return out[i].Unjustified > out[j].Unjustified
		}
		return out[i].Last.After(out[j].Last)
	})
	return &MissReport{Threshold: MissFlagThreshold, WindowDays: MissWindowDays, Patients: out}, nil
}

// PatientMisses is one patient's misses in the window (for the patient pop-up).
func (s *Service) PatientMisses(ctx context.Context, a Actor, patientID string) (*MissCount, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if uuid.Validate(patientID) != nil {
		return nil, ErrNotFound
	}
	if _, ok, err := s.repo.PatientCareLevel(ctx, cid, patientID); err != nil || !ok {
		return nil, firstErr(err, ErrNotFound)
	}
	counts, err := s.repo.MissedCounts(ctx, cid, s.missWindow(), patientID)
	if err != nil {
		return nil, err
	}
	if len(counts) == 0 {
		return &MissCount{PatientID: patientID}, nil
	}
	return &counts[0], nil
}

// missFields is what to store with a status change: the justification only
// for missed sessions (cleared otherwise).
func missFields(req StatusRequest) map[string]any {
	if req.Status != StatusMissed {
		return map[string]any{"status": req.Status, "miss_justified": nil, "miss_reason": ""}
	}
	justified := req.Justified != nil && *req.Justified
	return map[string]any{"status": req.Status, "miss_justified": justified, "miss_reason": req.Reason}
}
