// Package schedule is the clinic's time board: which physiotherapist works
// in which division (gym, treatment rooms, pool…) and when, and which
// patients are assigned to each physiotherapist per hour.
//
//	Division    - an area of the clinic. Per physiotherapist and hour it
//	              allows at most Capacity patients and CreditLimit credits.
//	Shift       - a physiotherapist in a division on a date, from-to hour
//	Appointment - a patient with a physiotherapist on a date at an hour.
//	              Credits = the patient's care level (1–5) when booked, so
//	              two level-5 patients never fit an 8-credit hour.
//	Series      - appointments booked together from a weekly plan
//	              (e.g. Mon/Wed/Fri at 10:00, 12 sessions).
//
// Layout follows the other modules (model, repository, service, series,
// board, stats, handler, routes).
package schedule

import (
	"context"
	"time"
)

// Appointment statuses.
const (
	StatusBooked   = "booked"
	StatusAttended = "attended"
	StatusMissed   = "missed"
)

// DefaultCreditLimit is the credits one physiotherapist handles per hour.
const DefaultCreditLimit = 8

// Colors a division can use on the board (Tailwind palette names).
var Colors = []string{"teal", "indigo", "amber", "sky", "rose", "violet", "emerald", "orange"}

// Division is an area of the clinic where physiotherapists work.
type Division struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ClinicID    string    `json:"clinic_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"size:60;not null"`
	Color       string    `json:"color" gorm:"size:20;not null;default:teal"`
	Capacity    int       `json:"capacity" gorm:"not null;default:3"`     // patients per physio per hour
	CreditLimit int       `json:"credit_limit" gorm:"not null;default:8"` // care-level credits per physio per hour
	Floor       int       `json:"floor" gorm:"not null;default:1"`        // building floor (1, 2, 3…)
	SortOrder   int       `json:"sort_order" gorm:"not null;default:0"`
	Active      bool      `json:"active" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_date"`
	UpdatedAt   time.Time `json:"modified_date"`
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
	ID        string `json:"id" gorm:"primaryKey;type:uuid"`
	Name      string `json:"name"`
	CareLevel int    `json:"care_level"`
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
	Credits     int             `json:"credits" gorm:"not null;default:1"` // patient's care level when booked
	ReferralID  *string         `json:"referral_id" gorm:"type:uuid"`
	SeriesID    *string         `json:"series_id" gorm:"type:uuid;index"`
	Status      string          `json:"status" gorm:"size:12;not null;default:booked"`
	// Missed sessions only: was the miss justified (ill, warned), and why
	MissJustified *bool     `json:"miss_justified"`
	MissReason    string    `json:"miss_reason" gorm:"size:200"`
	Notes         string    `json:"notes" gorm:"size:300"`
	CreatedByID   uint      `json:"created_by_id"`
	CreatedAt     time.Time `json:"created_date"`
	UpdatedAt     time.Time `json:"modified_date"`
}

// Models lists the tables for AutoMigrate.
func Models() []any {
	return []any{&Division{}, &Shift{}, &Appointment{}, &Absence{}, &HandoverNote{}, &WaitingEntry{}}
}

// SessionHook reacts to sessions being held, e.g. the inventory using the
// consumables every session needs. Warnings are shown to the user.
type SessionHook interface {
	// divisionID is where the session took place ("" if unknown).
	SessionAttended(ctx context.Context, clinicID, appointmentID, divisionID string, actorID uint) (warnings []string, err error)
	SessionUndone(ctx context.Context, clinicID, appointmentID string, actorID uint) error
}

type noHook struct{}

func (noHook) SessionAttended(context.Context, string, string, string, uint) ([]string, error) {
	return nil, nil
}
func (noHook) SessionUndone(context.Context, string, string, uint) error { return nil }

// DivisionRequest creates or edits a division.
type DivisionRequest struct {
	Name        string `json:"name" binding:"required,max=60"`
	Color       string `json:"color" binding:"omitempty,oneof=teal indigo amber sky rose violet emerald orange"`
	Capacity    int    `json:"capacity" binding:"required,min=1,max=20"`
	CreditLimit int    `json:"credit_limit" binding:"omitempty,min=1,max=40"`
	Floor       int    `json:"floor" binding:"min=-3,max=50"`
	SortOrder   int    `json:"sort_order"`
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

// RepeatWeekRequest repeats the week starting WeekStart (a Monday) for the
// following Weeks weeks.
type RepeatWeekRequest struct {
	WeekStart string `json:"week_start" binding:"required"`
	Weeks     int    `json:"weeks" binding:"required,min=1,max=26"`
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

// SeriesRequest books a weekly treatment plan: on the given weekdays
// (0 = Sunday … 6 = Saturday) at Hour, from StartDate, until Sessions are
// booked. Dates that cannot be booked are skipped and reported.
type SeriesRequest struct {
	PatientID   string  `json:"patient_id" binding:"required,uuid"`
	TherapistID uint    `json:"therapist_id" binding:"required"`
	Weekdays    []int   `json:"weekdays" binding:"required,min=1,max=7,dive,min=0,max=6"`
	Hour        int     `json:"hour" binding:"min=0,max=23"`
	StartDate   string  `json:"start_date" binding:"required"`
	Sessions    int     `json:"sessions" binding:"required,min=1,max=100"`
	ReferralID  *string `json:"referral_id" binding:"omitempty,uuid"`
	Notes       string  `json:"notes" binding:"max=300"`
	DryRun      bool    `json:"dry_run"` // preview only
}

// SeriesResult lists what was (or would be) booked and what was skipped.
type SeriesResult struct {
	SeriesID string        `json:"series_id,omitempty"`
	Booked   []string      `json:"booked"`
	Skipped  []SkippedDate `json:"skipped"`
}

// SkippedDate is a plan date that could not be booked, and why.
type SkippedDate struct {
	Date   string `json:"date"`
	Reason string `json:"reason"` // not_on_shift | full | over_credits | patient_busy
}

// StatusRequest marks an appointment attended or missed (or booked again).
type StatusRequest struct {
	Status string `json:"status" binding:"required,oneof=booked attended missed"`
	// Missed only: justified (default no) and an optional reason
	Justified *bool  `json:"justified"`
	Reason    string `json:"reason" binding:"max=200"`
}

// StatusResult is the updated appointment plus warnings (e.g. out of gloves).
type StatusResult struct {
	Appointment
	Warnings []string `json:"warnings"`
}
