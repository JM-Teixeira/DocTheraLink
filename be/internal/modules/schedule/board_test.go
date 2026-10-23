package schedule

import (
	"testing"
	"time"
)

// Regression: the hours of a row follow the clock even when the shifts come
// out of order (CI failed on map-ordered fakes).
func TestBuildBoardHoursInOrder(t *testing.T) {
	day := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	divs := []Division{{ID: "d1", Capacity: 2}}
	shifts := []Shift{
		{ID: "late", Date: day, StartHour: 14, EndHour: 16, TherapistID: 1, DivisionID: "d1"},
		{ID: "early", Date: day, StartHour: 8, EndHour: 10, TherapistID: 1, DivisionID: "d1"},
	}
	b := buildBoard(day, divs, []UserSummary{{ID: 1}}, shifts, nil)
	var got []int
	for _, h := range b.Therapists[0].Hours {
		got = append(got, h.Hour)
	}
	want := []int{8, 9, 14, 15}
	if len(got) != len(want) {
		t.Fatalf("hours %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("hours %v, want %v", got, want)
		}
	}
	if b.Therapists[0].Shifts[0].ID != "early" {
		t.Errorf("shifts not in time order: %v", b.Therapists[0].Shifts)
	}
}
