package schedule

import (
	"context"
	"time"
)

// MaxRangeDays bounds the week/month views.
const MaxRangeDays = 42

// RangeView summarises several days for the week and month views.
type RangeView struct {
	From        string        `json:"from"`
	To          string        `json:"to"`
	Divisions   []Division    `json:"divisions"`
	Therapists  []UserSummary `json:"therapists"`
	Days        []DaySummary  `json:"days"`
	Permissions Permissions   `json:"permissions"`
}

// DaySummary is one day: totals and the load of each physiotherapist.
type DaySummary struct {
	Date       string         `json:"date"`
	Totals     DayTotals      `json:"totals"`
	Therapists []TherapistDay `json:"therapists"` // only those working that day
}

// DayTotals counts the day's sessions and how full it is.
type DayTotals struct {
	Patients       int `json:"patients"` // appointments booked that day
	Attended       int `json:"attended"`
	Missed         int `json:"missed"`
	Capacity       int `json:"capacity"`
	Credits        int `json:"credits"`
	CreditCapacity int `json:"credit_capacity"`
}

// TherapistDay is a physiotherapist's day in the week view.
type TherapistDay struct {
	ID             uint    `json:"id"`
	Shifts         []Shift `json:"shifts"`
	Patients       int     `json:"patients"`
	Capacity       int     `json:"capacity"`
	Credits        int     `json:"credits"`
	CreditCapacity int     `json:"credit_capacity"`
}

// MonthStats is one month of the year view.
type MonthStats struct {
	Month      int `json:"month"` // 1–12
	Booked     int `json:"booked"`
	Attended   int `json:"attended"`
	Missed     int `json:"missed"`
	Patients   int `json:"patients"`   // distinct patients with sessions
	Referred   int `json:"referred"`   // referrals created
	Discharged int `json:"discharged"` // referrals completed (patients cleared)
}

// YearStats is the year view.
type YearStats struct {
	Year   int          `json:"year"`
	Months []MonthStats `json:"months"`
}

// Range builds the week / month view for from..to (inclusive).
func (s *Service) Range(ctx context.Context, a Actor, fromS, toS string) (*RangeView, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	from, err := parseDate("from", fromS)
	if err != nil {
		return nil, err
	}
	to, err := parseDate("to", toS)
	if err != nil {
		return nil, err
	}
	if to.Before(from) || to.Sub(from) > (MaxRangeDays-1)*24*time.Hour {
		return nil, &ValidationError{"to", "must be after from and at most 42 days later"}
	}
	data, err := s.load(ctx, cid, from, to)
	if err != nil {
		return nil, err
	}
	v := &RangeView{From: fromS, To: toS, Divisions: data.divisions, Therapists: data.therapists,
		Days: []DaySummary{}, Permissions: permissions(a)}
	if v.Therapists == nil {
		v.Therapists = []UserSummary{}
	}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		v.Days = append(v.Days, summarise(buildBoard(d, data.divisions, data.therapists, data.shifts, data.appts)))
	}
	return v, nil
}

// summarise turns a day's board into totals (pure, tested).
func summarise(b *Board) DaySummary {
	day := DaySummary{Date: b.Date, Therapists: []TherapistDay{}}
	for _, row := range b.Therapists {
		if len(row.Shifts) == 0 {
			continue
		}
		day.Therapists = append(day.Therapists, TherapistDay{ID: row.ID, Shifts: row.Shifts, Patients: row.Patients,
			Capacity: row.Capacity, Credits: row.Credits, CreditCapacity: row.CreditCapacity})
		t := &day.Totals
		t.Patients += row.Patients
		t.Capacity += row.Capacity
		t.Credits += row.Credits
		t.CreditCapacity += row.CreditCapacity
		for _, h := range row.Hours {
			for _, ap := range h.Appointments {
				switch ap.Status {
				case StatusAttended:
					t.Attended++
				case StatusMissed:
					t.Missed++
				}
			}
		}
	}
	return day
}

// Year returns month-by-month sessions and referrals (patients referred and
// discharged) of the clinic.
func (s *Service) Year(ctx context.Context, a Actor, year int) (*YearStats, error) {
	cid, err := s.clinicOf(ctx, a)
	if err != nil {
		return nil, err
	}
	if year < 2000 || year > 2100 {
		return nil, &ValidationError{"year", "is out of range"}
	}
	months, err := s.repo.MonthlyStats(ctx, cid, year)
	if err != nil {
		return nil, err
	}
	return &YearStats{Year: year, Months: months}, nil
}
