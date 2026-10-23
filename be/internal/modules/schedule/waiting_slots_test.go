package schedule

import (
	"testing"
	"time"
)

func TestFirstSlot(t *testing.T) {
	monday := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC) // a Monday
	room := Division{ID: "d1", Name: "Ginásio", Capacity: 1, CreditLimit: 5}
	data := &boardData{
		divisions:  []Division{room},
		therapists: []UserSummary{{ID: 1, FirstName: "Rui", LastName: "Costa"}, {ID: 2, FirstName: "Ana"}},
		shifts: []Shift{
			{Date: monday, StartHour: 9, EndHour: 12, TherapistID: 1, DivisionID: "d1"},
			{Date: monday.AddDate(0, 0, 1), StartHour: 14, EndHour: 16, TherapistID: 2, DivisionID: "d1"},
		},
		appts: []Appointment{{ID: "a", Date: monday, Hour: 9, TherapistID: 1, PatientID: "other", Credits: 1}},
	}
	entry := WaitingEntry{PatientID: "p1", Status: WaitingOpen, Period: "any"}

	got := firstSlot(data, entry, monday, 0)
	if got == nil || got.Date != "2026-10-26" || got.Hour != 10 || got.TherapistID != 1 || got.Therapist != "Rui Costa" {
		t.Fatalf("first free hour after the booked one: %+v", got)
	}
	// Afternoons only: Tuesday with Ana
	entry.Period = "afternoon"
	if got = firstSlot(data, entry, monday, 0); got == nil || got.Date != "2026-10-27" || got.Hour != 14 || got.TherapistID != 2 {
		t.Errorf("afternoon: %+v", got)
	}
	// Only Rui and only afternoons: nothing in two weeks
	rui := uint(1)
	entry.PreferredTherapistID = &rui
	if got = firstSlot(data, entry, monday, 0); got != nil {
		t.Errorf("no afternoon with Rui: %+v", got)
	}
	// Only Wednesdays: nothing
	entry = WaitingEntry{PatientID: "p1", Weekdays: []int{3}, Period: "any"}
	if got = firstSlot(data, entry, monday, 0); got != nil {
		t.Errorf("wednesdays: %+v", got)
	}
	// Later today only: from 11 on
	entry = WaitingEntry{PatientID: "p1", Period: "morning"}
	if got = firstSlot(data, entry, monday, 11); got == nil || got.Hour != 11 {
		t.Errorf("from 11: %+v", got)
	}
	// A care level that does not fit the room's credits
	entry.Patient = &PatientSummary{CareLevel: 5}
	data.appts = append(data.appts, Appointment{ID: "b", Date: monday, Hour: 11, TherapistID: 9, PatientID: "x"})
	room.CreditLimit = 4
	data.divisions[0] = room
	if got = firstSlot(data, entry, monday, 11); got != nil {
		t.Errorf("level 5 over the credit limit: %+v", got)
	}
}
