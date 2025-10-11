package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FacilitySection represents configurable sections within a facility
// Examples: Electric Section, Pool Section, Massage Section, Gym Section
type FacilitySection struct {
	ID          string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string `json:"name" gorm:"not null"` // e.g., "Electric Section", "Pool Section"
	Description string `json:"description"`
	Color       string `json:"color" gorm:"default:'#3B82F6'"` // Hex color for UI display
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	
	// Foreign keys
	FacilityID string `json:"facility_id" gorm:"not null"`
	CreatedByID uint  `json:"-"`
	
	// Relationships
	Facility  Facility        `json:"facility" gorm:"foreignKey:FacilityID"`
	CreatedBy UserBareMinimum `json:"created_by" gorm:"foreignKey:CreatedByID"`
	
	CreatedAt time.Time      `json:"created_date"`
	UpdatedAt time.Time      `json:"modified_date"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TherapyPatient model temporarily disabled due to GORM constraint conflicts
// See notes_disabled.go for details

// TreatmentPlan temporarily disabled due to TherapyPatient dependency

// TherapySession temporarily disabled due to TherapyPatient dependency

// TherapistSchedule represents therapist availability and schedules
type TherapistSchedule struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TherapistID uint      `json:"therapist_id" gorm:"not null"`
	SectionID   string    `json:"section_id" gorm:"not null"`
	
	// Schedule details
	DayOfWeek   int       `json:"day_of_week" gorm:"not null"` // 0=Sunday, 1=Monday, etc.
	StartTime   time.Time `json:"start_time" gorm:"not null"`
	EndTime     time.Time `json:"end_time" gorm:"not null"`
	
	// Availability
	IsAvailable bool      `json:"is_available" gorm:"default:true"`
	MaxPatients int       `json:"max_patients" gorm:"default:1"`
	
	// Date range (for temporary schedules)
	ValidFrom   *time.Time `json:"valid_from"`
	ValidTo     *time.Time `json:"valid_to"`
	
	// Foreign keys
	CreatedByID uint      `json:"-"`
	
	// Relationships
	Therapist   User            `json:"therapist" gorm:"foreignKey:TherapistID"`
	Section     FacilitySection `json:"section" gorm:"foreignKey:SectionID"`
	CreatedBy   UserBareMinimum `json:"created_by" gorm:"foreignKey:CreatedByID"`
	
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// PatientNote represents notes about a patient (extends the base PatientNote)
// TherapyPatientNote model temporarily disabled due to GORM constraint conflicts
// See notes_disabled.go for details

// PatientStakeholder temporarily disabled due to TherapyPatient dependency

// Request/Response types for API

// CreateFacilitySectionRequest represents request to create a facility section
type CreateFacilitySectionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// CreateTherapyPatientRequest temporarily disabled

// CreateTreatmentPlanRequest temporarily disabled

// CreateTherapySessionRequest temporarily disabled

// UpdateSessionNotesRequest temporarily disabled

// CreatePatientNoteRequest temporarily disabled - see notes_disabled.go

// TimetableEntry represents a single entry in the timetable view
type TimetableEntry struct {
	ID            string    `json:"id"`
	PatientID     string    `json:"patient_id"`
	PatientName   string    `json:"patient_name"`
	TherapistID   uint      `json:"therapist_id"`
	TherapistName string    `json:"therapist_name"`
	SectionID     string    `json:"section_id"`
	SectionName   string    `json:"section_name"`
	SectionColor  string    `json:"section_color"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	Duration      int       `json:"duration"`
}

// TimetableResponse represents the timetable view response
type TimetableResponse struct {
	Date    time.Time        `json:"date"`
	Entries []TimetableEntry `json:"entries"`
}

// BeforeCreate hooks to set UUIDs
func (fs *FacilitySection) BeforeCreate(tx *gorm.DB) error {
	if fs.ID == "" {
		fs.ID = uuid.New().String()
	}
	return nil
}

// TherapyPatient BeforeCreate hook temporarily disabled

// TreatmentPlan BeforeCreate hook temporarily disabled

// TherapySession BeforeCreate hook temporarily disabled

func (tsc *TherapistSchedule) BeforeCreate(tx *gorm.DB) error {
	if tsc.ID == "" {
		tsc.ID = uuid.New().String()
	}
	return nil
}

// TherapyPatientNote BeforeCreate hook temporarily disabled

// PatientStakeholder BeforeCreate hook temporarily disabled