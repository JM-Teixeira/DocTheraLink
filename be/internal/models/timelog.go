package models

import (
	"time"

	"gorm.io/gorm"
)

// TimeLog represents a user's clock in/out record
type TimeLog struct {
	ID       uint       `json:"id" gorm:"primaryKey"`
	UserID   uint       `json:"user_id" gorm:"not null;index"`
	User     User       `json:"user" gorm:"foreignKey:UserID"`
	ClockIn  time.Time  `json:"clock_in" gorm:"not null"`
	ClockOut *time.Time `json:"clock_out"`
	Date     time.Time  `json:"date" gorm:"type:date;not null;index"`
	Duration int        `json:"duration" gorm:"default:0"` // in minutes
	Notes    string     `json:"notes" gorm:"type:text"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// TimeLogRequest represents clock in/out request
type TimeLogRequest struct {
	Notes string `json:"notes"`
}

// TimeLogQueryParams for filtering time logs
type TimeLogQueryParams struct {
	UserID uint   `form:"user_id"`
	From   string `form:"from"` // YYYY-MM-DD
	To     string `form:"to"`   // YYYY-MM-DD
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// CalculateDuration calculates the duration in minutes between clock in and clock out
func (tl *TimeLog) CalculateDuration() int {
	if tl.ClockOut == nil {
		return 0
	}
	duration := tl.ClockOut.Sub(tl.ClockIn)
	return int(duration.Minutes())
}

// BeforeSave hook to calculate duration before saving.
// GORM only calls hooks with the exact (*gorm.DB) error signature.
func (tl *TimeLog) BeforeSave(*gorm.DB) error {
	if tl.ClockOut != nil {
		tl.Duration = tl.CalculateDuration()
	}
	return nil
}
