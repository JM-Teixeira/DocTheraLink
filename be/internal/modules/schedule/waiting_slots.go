package schedule

import (
	"context"
	"slices"
	"strings"
	"time"
)

// Free slots for the waiting list: for each waiting patient, the first hour
// in the next two weeks that suits them (their days, morning / afternoon,
// the chosen physio or any) and where the booking rules allow them (on
// shift, not absent, room and credits, patient free). The board highlights
// these entries and the coordinator's home counts them.

const slotHorizonDays = 14

// Slot is a free hour for a waiting patient.
type Slot struct {
	Date        string `json:"date"` // YYYY-MM-DD
	Hour        int    `json:"hour"`
	TherapistID uint   `json:"therapist_id"`
	Therapist   string `json:"therapist"`
}

// periodHas reports whether an hour falls in the preferred time of day.
func periodHas(period string, hour int) bool {
	switch period {
	case "morning":
		return hour < 13
	case "afternoon":
		return hour >= 13
	}
	return true
}

// firstSlot finds the earliest suitable hour for one entry, from the start
// of hour `fromHour` today on.
func firstSlot(data *boardData, e WaitingEntry, today time.Time, fromHour int) *Slot {
	level := 1
	if e.Patient != nil && e.Patient.CareLevel > 0 {
		level = e.Patient.CareLevel
	}
	plan := planner{data: data, level: level, exclude: map[string]bool{}}
	names := map[uint]string{}
	for _, th := range data.therapists {
		names[th.ID] = strings.TrimSpace(th.FirstName + " " + th.LastName)
	}
	for i := 0; i <= slotHorizonDays; i++ {
		d := today.AddDate(0, 0, i)
		if len(e.Weekdays) > 0 && !slices.Contains(e.Weekdays, int(d.Weekday())) {
			continue
		}
		for hour := 0; hour < 24; hour++ {
			if (i == 0 && hour < fromHour) || !periodHas(e.Period, hour) {
				continue
			}
			for _, sh := range data.shifts {
				if !sameDate(sh.Date, d) || !sh.Covers(hour) {
					continue
				}
				if e.PreferredTherapistID != nil && sh.TherapistID != *e.PreferredTherapistID {
					continue
				}
				if reason, _ := plan.problem(d, hour, sh.TherapistID, e.PatientID); reason == "" {
					return &Slot{Date: d.Format(dateLayout), Hour: hour, TherapistID: sh.TherapistID, Therapist: names[sh.TherapistID]}
				}
			}
		}
	}
	return nil
}

// withSlots fills Slot on open entries (one board load for all of them).
func (s *Service) withSlots(ctx context.Context, cid string, entries []WaitingEntry) error {
	if len(entries) == 0 {
		return nil
	}
	today := s.today()
	data, err := s.load(ctx, cid, today, today.AddDate(0, 0, slotHorizonDays))
	if err != nil {
		return err
	}
	fromHour := s.now().Hour() + 1 // a slot later today, not one already started
	for i := range entries {
		if entries[i].Status == WaitingOpen {
			entries[i].Slot = firstSlot(data, entries[i], today, fromHour)
		}
	}
	return nil
}
