//go:build integration

package integration

import (
	"testing"

	"klyniva/internal/database"
	"klyniva/internal/models"
	"klyniva/internal/modules/patient"
)

// TestClimutuaSeed: the PT-PT demo clinic is complete and consistent — its
// sessions respect the patient and credit limits of every physio-hour, NIFs
// are valid, stock reflects the sessions, and seeding twice adds nothing.
func TestClimutuaSeed(t *testing.T) {
	startDB(t)
	db := database.GetDB()
	cid := database.ClimutuaClinicID

	count := func(what, table, where string, args ...any) int64 {
		t.Helper()
		var n int64
		if err := db.Table(table).Where(where, args...).Count(&n).Error; err != nil {
			t.Fatalf("%s: %v", what, err)
		}
		return n
	}
	if n := count("de.nuno.dias", "users", "home_facility_id = ? AND employment_status = 'active'", cid); n != 13 {
		t.Errorf("staff: %d", n)
	}
	if n := count("patients", "patients", "facility_id = ?", cid); n != database.ClimutuaPatients {
		t.Errorf("patients: %d", n)
	}
	if n := count("referrals", "referrals", "facility_id = ? AND status = 'in_progress'", cid); n != 84 {
		t.Errorf("referrals in treatment: %d", n)
	}
	if n := count("attended", "appointments", "clinic_id = ? AND status = 'attended'", cid); n < 400 {
		t.Errorf("attended sessions: %d", n)
	}
	if n := count("future", "appointments", "clinic_id = ? AND status = 'booked'", cid); n < 300 {
		t.Errorf("booked sessions: %d", n)
	}
	if n := count("former staff", "users", "home_facility_id = ? AND employment_status = 'left' AND NOT is_active", cid); n != 1 {
		t.Errorf("former staff kept as a record: %d", n)
	}
	if n := count("physio skills", "user_skills", "user_id IN (SELECT id FROM users WHERE username LIKE 'cm.%' AND user_type = 'PhysioTherapist')"); n < 12 {
		t.Errorf("physio skills: %d", n)
	}
	if n := count("holiday", "absences", "clinic_id = ? AND reason = 'vacation' AND cover_id IS NOT NULL", cid); n != 1 {
		t.Errorf("Inês's holiday: %d", n)
	}
	if n := count("handover", "handover_notes", "absence_id IN (SELECT id FROM absences WHERE clinic_id = ?)", cid); n == 0 {
		t.Error("the holiday should come with handover notes")
	}
	if n := count("floors", "divisions", "clinic_id = ? AND floor = 3", cid); n != 1 {
		t.Errorf("third floor divisions: %d", n)
	}
	if n := count("PT diagnosis", "referrals", "facility_id = ? AND diagnosis = ?", cid, "Sequelas de AVC isquémico"); n == 0 {
		t.Error("expected Portuguese diagnoses")
	}

	// No physio-hour over the division's patients or credits
	var over int64
	err := db.Raw(`
		SELECT COUNT(*) FROM (
		  SELECT a.therapist_id, a.date, a.hour, COUNT(*) AS n, SUM(a.credits) AS credits, MAX(d.capacity) AS cap, MAX(d.credit_limit) AS lim
		    FROM appointments a
		    JOIN shifts s ON s.therapist_id = a.therapist_id AND s.date = a.date AND a.hour >= s.start_hour AND a.hour < s.end_hour
		    JOIN divisions d ON d.id = s.division_id
		   WHERE a.clinic_id = ?
		   GROUP BY a.therapist_id, a.date, a.hour) x
		WHERE n > cap OR credits > lim`, cid).Scan(&over).Error
	if err != nil || over != 0 {
		t.Fatalf("physio-hours over the limits: %d (%v)", over, err)
	}
	var orphan int64
	db.Raw(`SELECT COUNT(*) FROM appointments a WHERE a.clinic_id = ? AND NOT EXISTS (
		SELECT 1 FROM shifts s WHERE s.therapist_id = a.therapist_id AND s.date = a.date AND a.hour >= s.start_hour AND a.hour < s.end_hour)`, cid).Scan(&orphan)
	if orphan != 0 {
		t.Errorf("sessions outside any shift: %d", orphan)
	}

	var patients []models.Patient
	db.Where("facility_id = ?", cid).Find(&patients)
	for _, p := range patients {
		if p.NIF == nil || !patient.ValidNIF(*p.NIF) || p.CareLevel < 1 || p.CareLevel > 5 {
			t.Fatalf("patient %s: NIF %v, care level %d", p.Name, p.NIF, p.CareLevel)
		}
	}

	// Gloves went down with the attended sessions and are audited per session
	var gloves struct{ Quantity int }
	db.Table("stock_items").Where("clinic_id = ? AND name = ?", cid, "Luvas de nitrilo (M)").Select("quantity").Scan(&gloves)
	auto := count("session use", "stock_movements", "appointment_id IS NOT NULL AND item_id IN (SELECT id FROM stock_items WHERE clinic_id = ?)", cid)
	if gloves.Quantity >= 1400 || gloves.Quantity <= 0 || auto == 0 {
		t.Errorf("gloves %d, automatic uses %d", gloves.Quantity, auto)
	}

	// Seeding again (every start) adds no duplicates
	before := count("appointments", "appointments", "clinic_id = ?", cid)
	if err := database.SeedData(true); err != nil {
		t.Fatal(err)
	}
	if after := count("appointments", "appointments", "clinic_id = ?", cid); after != before {
		t.Errorf("re-seeding changed appointments: %d → %d", before, after)
	}
	if n := count("patients", "patients", "facility_id = ?", cid); n != database.ClimutuaPatients {
		t.Errorf("re-seeding changed patients: %d", n)
	}
}
