package schedule

import "time"

// Default opening hours shown on the board; shifts outside widen it.
const (
	defaultOpen  = 8
	defaultClose = 20
)

// Board is one day of the clinic: divisions, and per physiotherapist their
// shifts and how many patients they have each hour.
type Board struct {
	Date        string           `json:"date"`
	OpenHour    int              `json:"open_hour"`
	CloseHour   int              `json:"close_hour"` // exclusive
	Divisions   []Division       `json:"divisions"`
	Therapists  []BoardTherapist `json:"therapists"`
	Permissions Permissions      `json:"permissions"`
}

// Permissions tell the UI which controls to show (the API enforces them).
type Permissions struct {
	ManageShifts bool `json:"manage_shifts"`
	BookPatients bool `json:"book_patients"`
}

// BoardTherapist is one row of the board.
type BoardTherapist struct {
	UserSummary
	Shifts   []Shift    `json:"shifts"`
	Hours    []HourLoad `json:"hours"` // only hours on shift
	Patients int        `json:"patients"`
	Capacity int        `json:"capacity"` // sum over the hours on shift
}

// HourLoad is one physiotherapist-hour: where, how many patients, how many fit.
type HourLoad struct {
	Hour         int           `json:"hour"`
	DivisionID   string        `json:"division_id"`
	ShiftID      string        `json:"shift_id"`
	Capacity     int           `json:"capacity"`
	Appointments []Appointment `json:"appointments"`
}

// buildBoard assembles the board from the day's records (pure, tested).
func buildBoard(date time.Time, divs []Division, therapists []UserSummary, shifts []Shift, appts []Appointment) *Board {
	capacity := map[string]int{}
	for _, d := range divs {
		capacity[d.ID] = d.Capacity
	}
	b := &Board{Date: date.Format(dateLayout), OpenHour: defaultOpen, CloseHour: defaultClose, Divisions: divs}
	for _, sh := range shifts {
		b.OpenHour = min(b.OpenHour, sh.StartHour)
		b.CloseHour = max(b.CloseHour, sh.EndHour)
	}
	if b.Divisions == nil {
		b.Divisions = []Division{}
	}

	for _, t := range therapists {
		row := BoardTherapist{UserSummary: t, Shifts: []Shift{}, Hours: []HourLoad{}}
		for _, sh := range shifts {
			if sh.TherapistID != t.ID {
				continue
			}
			row.Shifts = append(row.Shifts, sh)
			for h := sh.StartHour; h < sh.EndHour; h++ {
				load := HourLoad{Hour: h, DivisionID: sh.DivisionID, ShiftID: sh.ID, Capacity: capacity[sh.DivisionID], Appointments: []Appointment{}}
				for _, ap := range appts {
					if ap.TherapistID == t.ID && ap.Hour == h {
						load.Appointments = append(load.Appointments, ap)
					}
				}
				row.Patients += len(load.Appointments)
				row.Capacity += load.Capacity
				row.Hours = append(row.Hours, load)
			}
		}
		b.Therapists = append(b.Therapists, row)
	}
	if b.Therapists == nil {
		b.Therapists = []BoardTherapist{}
	}
	return b
}
