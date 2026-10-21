package schedule

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SeedDemo gives the demo clinic its divisions (once) and a typical week of
// shifts for its physiotherapists on every weekday of the next two weeks
// that has none yet, so the board always has something to show.
func SeedDemo(db *gorm.DB, clinicID string, today time.Time) error {
	var n int64
	if err := db.Model(&Division{}).Where("clinic_id = ?", clinicID).Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		for i, d := range []Division{
			{Name: "Gym", Color: "teal", Capacity: 4},
			{Name: "Treatment room 1", Color: "indigo", Capacity: 1},
			{Name: "Treatment room 2", Color: "amber", Capacity: 1},
			{Name: "Hydrotherapy pool", Color: "sky", Capacity: 3},
		} {
			d.ClinicID, d.SortOrder, d.Active = clinicID, i, true
			if err := db.Create(&d).Error; err != nil {
				return fmt.Errorf("seed division %s: %w", d.Name, err)
			}
		}
	}
	var divs []Division
	if err := db.Where("clinic_id = ? AND active", clinicID).Order("sort_order").Find(&divs).Error; err != nil || len(divs) < 4 {
		return err
	}
	var physios []UserSummary
	if err := db.Where("home_facility_id = ? AND user_type = ? AND is_active", clinicID, "PhysioTherapist").
		Order("id").Limit(2).Find(&physios).Error; err != nil || len(physios) < 2 {
		return err
	}
	// Morning and afternoon pattern per physiotherapist
	pattern := [][]struct{ start, end, div int }{
		{{8, 13, 0}, {14, 18, 1}},
		{{9, 13, 3}, {14, 19, 0}},
	}
	start := today.Truncate(24 * time.Hour)
	for d := 0; d < 14; d++ {
		date := start.AddDate(0, 0, d)
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}
		var existing int64
		if err := db.Model(&Shift{}).Where("clinic_id = ? AND date = ?", clinicID, day(date)).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			continue
		}
		for i, p := range physios {
			for _, s := range pattern[i] {
				sh := Shift{ClinicID: clinicID, Date: date, StartHour: s.start, EndHour: s.end,
					TherapistID: p.ID, DivisionID: divs[s.div].ID}
				if err := db.Create(&sh).Error; err != nil {
					return fmt.Errorf("seed shift: %w", err)
				}
			}
		}
	}
	return nil
}
