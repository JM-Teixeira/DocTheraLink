package handlers

import (
	"errors"
	"strings"

	"klyniva/internal/models"
	"klyniva/internal/usernames"

	"gorm.io/gorm"
)

// Clinic codes prefix the usernames of their people (cm.joao.teixeira), so a
// code is 2-4 lower-case letters, unique, and fixed once set.
var (
	errCodeFormat = errors.New("the clinic code must be 2 to 4 lower-case letters")
	errCodeTaken  = errors.New("another clinic already uses this code")
	errCodeFixed  = errors.New("the clinic code cannot change once set (usernames use it)")
)

// normalizeCode trims and lower-cases what was typed.
func normalizeCode(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// checkClinicCode validates a code for clinic selfID ("" for a new clinic).
// current is the clinic's code today ("" if none).
func checkClinicCode(db *gorm.DB, code, selfID, current string) error {
	if current != "" && code != current {
		return errCodeFixed
	}
	if !usernames.ValidCode(code) {
		return errCodeFormat
	}
	var n int64
	q := db.Model(&models.Facility{}).Where("code = ?", code)
	if selfID != "" {
		q = q.Where("id <> ?", selfID)
	}
	if err := q.Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return errCodeTaken
	}
	return nil
}

// codeError reports whether err is one of the validation errors above.
func codeError(err error) bool {
	return errors.Is(err, errCodeFormat) || errors.Is(err, errCodeTaken) || errors.Is(err, errCodeFixed)
}
