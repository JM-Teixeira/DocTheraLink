// Package schedule is the clinic's time board: which physiotherapist works
// in which division (gym, treatment rooms, pool…) and when, and which
// patients are assigned to each physiotherapist per hour.
//
//	Division    - an area of the clinic; Capacity = patients one
//	              physiotherapist can take per hour there
//	Shift       - a physiotherapist in a division on a date, from-to hour
//	Appointment - a patient with a physiotherapist on a date at an hour
//
// Layout follows the other modules (model, repository, service, board,
// handler, routes).
package schedule

import "time"

// Appointment statuses.
const (
	StatusBooked   = "booked"
	StatusAttended = "attended"
	StatusMissed   = "missed"
)

// Colors a division can use on the board (Tailwind palette names).
var Colors = []string{"teal", "indigo", "amber", "sky", "rose", "violet", "emerald", "orange"}

// Division is an area of the clinic where physiotherapists work.
type Division struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ClinicID  string    `json:"clinic_id" gorm:"type:uuid;not null;index"`
	Name      string    `json:"name" gorm:"size:60;not null"`
	Color     string    `json:"color" gorm:"size:20;not null;default:teal"`
	Capacity  int       `json:"capacity" gorm:"not null;default:3"` // patients per physio per hour
	SortOrder int       `json:"sort_order" gorm:"not null;default:0"`
	Active    bool      `json:"active" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// UserSummary is a physiotherapist as shown on the board.
type UserSummary struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (UserSummary) TableName() string { return "users" }

// PatientSummary is a patient as shown on the board.
type PatientSummary struct {
	ID   string `json:"id" gorm:"primaryKey;type:uuid"`
	Name string `json:"name"`
}

func (PatientSummary) TableName() string { return "patients" }

// Shift places a physiotherapist in a division on a date, StartHour to EndHour.
type Shift struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ClinicID    string    `json:"clinic_id" gorm:"type:uuid;not null;index:idx_shift_day"`
	Date        time.Time `json:"date" gorm:"type:date;not null;index:idx_shift_day"`
	StartHour   int       `json:"start_hour" gorm:"not null"`
	EndHour     int       `json:"end_hour" gorm:"not null"` // exclusive
	TherapistID uint      `json:"therapist_id" gorm:"not null;index"`
	DivisionID  string    `json:"division_id" gorm:"type:uuid;not null"`
	Notes       string    `json:"notes" gorm:"size:200"`
	CreatedByID uint      `json:"created_by_id"`
	CreatedAt   time.Time `json:"created_date"`
	UpdatedAt   time.Time `json:"modified_date"`
}

// Covers reports whether the shift includes the hour.
func (s Shift) Covers(hour int) bool { return hour >= s.StartHour && hour < s.EndHour }

// Appointment assigns a patient to a physiotherapist for one hour.
type Appointment struct {
	ID          string          `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ClinicID    string          `json:"clinic_id" gorm:"type:uuid;not null;index:idx_appt_day"`
	Date        time.Time       `json:"date" gorm:"type:date;not null;index:idx_appt_day"`
	Hour        int             `json:"hour" gorm:"not null"`
	TherapistID uint            `json:"therapist_id" gorm:"not null;index"`
	PatientID   string          `json:"patient_id" gorm:"type:uuid;not null;index"`
	Patient     *PatientSummary `json:"patient,omitempty" gorm:"foreignKey:PatientID;->;-:migration"`
	ReferralID  *string         `json:"referral_id" gorm:"type:uuid"`
	Status      string          `json:"status" gorm:"size:12;not null;default:booked"`
	Notes       string          `json:"notes" gorm:"size:300"`
	CreatedByID uint            `json:"created_by_id"`
	CreatedAt   time.Time       `json:"created_date"`
	UpdatedAt   time.Time       `json:"modified_date"`
}

// Models lists the tables for AutoMigrate.
func Models() []any { return []any{&Division{}, &Shift{}, &Appointment{}} }

// DivisionRequest creates or edits a division.
type DivisionRequest struct {
	Name      string `json:"name" binding:"required,max=60"`
	Color     string `json:"color" binding:"omitempty,oneof=teal indigo amber sky rose violet emerald orange"`
	Capacity  int    `json:"capacity" binding:"required,min=1,max=20"`
	SortOrder int    `json:"sort_order"`
}

// ShiftRequest creates or edits a shift.
type ShiftRequest struct {
	Date        string `json:"date" binding:"required"` // YYYY-MM-DD
	StartHour   int    `json:"start_hour" binding:"min=0,max=23"`
	EndHour     int    `json:"end_hour" binding:"min=1,max=24"`
	TherapistID uint   `json:"therapist_id" binding:"required"`
	DivisionID  string `json:"division_id" binding:"required,uuid"`
	Notes       string `json:"notes" binding:"max=200"`
}

// CopyRequest copies every shift of one day onto another.
type CopyRequest struct {
	FromDate string `json:"from_date" binding:"required"`
	ToDate   string `json:"to_date" binding:"required"`
}

// AppointmentRequest assigns a patient to a physiotherapist's hour.
type AppointmentRequest struct {
	Date        string  `json:"date" binding:"required"`
	Hour        int     `json:"hour" binding:"min=0,max=23"`
	TherapistID uint    `json:"therapist_id" binding:"required"`
	PatientID   string  `json:"patient_id" binding:"required,uuid"`
	ReferralID  *string `json:"referral_id" binding:"omitempty,uuid"`
	Notes       string  `json:"notes" binding:"max=300"`
}

// StatusRequest marks an appointment attended or missed (or booked again).
type StatusRequest struct {
	Status string `json:"status" binding:"required,oneof=booked attended missed"`
}
