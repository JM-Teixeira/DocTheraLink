package schedule

import "time"

// Default opening hours shown on the board; shifts outside widen it.
const (
	defaultOpen  = 8
	defaultClose = 20
)

// Board is one day of the clinic: divisions, and per physiotherapist their
// shifts and how many patients (and care credits) they have each hour.
type Board struct {
	Date        string           `json:"date"`
	OpenHour    int              `json:"open_hour"`
	CloseHour   int              `json:"close_hour"` // exclusive
	Divisions   []Division       `json:"divisions"`
	Therapists  []BoardTherapist `json:"therapists"`
	Permissions Permissions      `json:"permissions"`
	Absences    []Absence        `json:"absences"` // overlapping the date
}

// Permissions tell the UI which controls to show (the API enforces them).
type Permissions struct {
	ManageShifts bool `json:"manage_shifts"`
	BookPatients bool `json:"book_patients"`
}

// BoardTherapist is one row of the board.
type BoardTherapist struct {
	UserSummary
	Shifts         []Shift    `json:"shifts"`
	Hours          []HourLoad `json:"hours"` // only hours on shift
	Patients       int        `json:"patients"`
	Capacity       int        `json:"capacity"` // patients, summed over the hours on shift
	Credits        int        `json:"credits"`
	CreditCapacity int        `json:"credit_capacity"`
}

// HourLoad is one physiotherapist-hour: where, who, and how much room is left.
type HourLoad struct {
	Hour         int           `json:"hour"`
	DivisionID   string        `json:"division_id"`
	ShiftID      string        `json:"shift_id"`
	Capacity     int           `json:"capacity"`
	Credits      int           `json:"credits"`
	CreditLimit  int           `json:"credit_limit"`
	Appointments []Appointment `json:"appointments"`
}

// buildBoard assembles the board of one date from records that may span
// more days (pure, tested).
func buildBoard(date time.Time, divs []Division, therapists []UserSummary, shifts []Shift, appts []Appointment) *Board {
	byID := map[string]Division{}
	for _, d := range divs {
		byID[d.ID] = d
	}
	b := &Board{Date: date.Format(dateLayout), OpenHour: defaultOpen, CloseHour: defaultClose, Divisions: divs}
	if b.Divisions == nil {
		b.Divisions = []Division{}
	}
	var dayShifts []Shift
	for _, sh := range shifts {
		if sameDate(sh.Date, date) {
			dayShifts = append(dayShifts, sh)
			b.OpenHour = min(b.OpenHour, sh.StartHour)
			b.CloseHour = max(b.CloseHour, sh.EndHour)
		}
	}

	b.Therapists = []BoardTherapist{}
	for _, t := range therapists {
		row := BoardTherapist{UserSummary: t, Shifts: []Shift{}, Hours: []HourLoad{}}
		for _, sh := range dayShifts {
			if sh.TherapistID != t.ID {
				continue
			}
			row.Shifts = append(row.Shifts, sh)
			div := byID[sh.DivisionID]
			for h := sh.StartHour; h < sh.EndHour; h++ {
				load := HourLoad{Hour: h, DivisionID: sh.DivisionID, ShiftID: sh.ID, Capacity: div.Capacity,
					CreditLimit: creditLimit(&div), Appointments: []Appointment{}}
				for _, ap := range appts {
					if ap.TherapistID == t.ID && ap.Hour == h && sameDate(ap.Date, date) {
						load.Appointments = append(load.Appointments, ap)
						load.Credits += ap.Credits
					}
				}
				row.Patients += len(load.Appointments)
				row.Capacity += load.Capacity
				row.Credits += load.Credits
				row.CreditCapacity += load.CreditLimit
				row.Hours = append(row.Hours, load)
			}
		}
		b.Therapists = append(b.Therapists, row)
	}
	return b
}

func sameDate(a, b time.Time) bool { return a.Format(dateLayout) == b.Format(dateLayout) }
