package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TimetableSection represents a section/room in a facility for timetable scheduling
type TimetableSection struct {
	ID          string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FacilityID  string `json:"facility_id" gorm:"not null;index"`
	Name        string `json:"name" gorm:"not null"`
	MaxCapacity int    `json:"max_capacity" gorm:"default:3"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Description string `json:"description" gorm:"type:text"`

	Facility Facility `json:"facility" gorm:"foreignKey:FacilityID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// TimetableSlot represents a staff assignment to a section at a specific time
type TimetableSlot struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FacilityID  string     `json:"facility_id" gorm:"not null;index"`
	SectionID   string     `json:"section_id" gorm:"not null;index"`
	UserID      uint       `json:"user_id" gorm:"not null;index"`
	DayOfWeek   int        `json:"day_of_week" gorm:"not null"` // 0=Sunday, 6=Saturday
	StartTime   string     `json:"start_time" gorm:"not null"`  // "09:00"
	EndTime     string     `json:"end_time" gorm:"not null"`    // "17:00"
	IsRecurring bool       `json:"is_recurring" gorm:"default:true"`
	ValidFrom   time.Time  `json:"valid_from"`
	ValidUntil  *time.Time `json:"valid_until"`
	Notes       string     `json:"notes" gorm:"type:text"`

	Facility Facility         `json:"facility" gorm:"foreignKey:FacilityID"`
	Section  TimetableSection `json:"section" gorm:"foreignKey:SectionID"`
	User     User             `json:"user" gorm:"foreignKey:UserID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// TimetableSectionRequest represents section creation/update request
type TimetableSectionRequest struct {
	Name        string `json:"name" binding:"required"`
	MaxCapacity int    `json:"max_capacity"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
}

// TimetableSlotRequest represents timetable slot creation/update request
type TimetableSlotRequest struct {
	SectionID   string     `json:"section_id" binding:"required"`
	UserID      uint       `json:"user_id" binding:"required"`
	DayOfWeek   int        `json:"day_of_week" binding:"required,min=0,max=6"`
	StartTime   string     `json:"start_time" binding:"required"`
	EndTime     string     `json:"end_time" binding:"required"`
	IsRecurring bool       `json:"is_recurring"`
	ValidFrom   *time.Time `json:"valid_from"`
	ValidUntil  *time.Time `json:"valid_until"`
	Notes       string     `json:"notes"`
}

// TimetableQueryParams for filtering timetable
type TimetableQueryParams struct {
	DayOfWeek *int   `form:"day_of_week"`
	SectionID string `form:"section_id"`
	UserID    *uint  `form:"user_id"`
	Date      string `form:"date"` // YYYY-MM-DD
}

// BeforeCreate hooks
func (ts *TimetableSection) BeforeCreate(tx *gorm.DB) error {
	if ts.ID == "" {
		ts.ID = uuid.New().String()
	}
	return nil
}

func (ts *TimetableSlot) BeforeCreate(tx *gorm.DB) error {
	if ts.ID == "" {
		ts.ID = uuid.New().String()
	}
	return nil
}
