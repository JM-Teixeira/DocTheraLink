package models

import (
	"regexp"
	"strings"
)

// Portugal phone number constants
const (
	DefaultCountryCode = "+351"
	PhonePlaceholder   = "+351 912 345 678"
)

// ValidatePortuguesePhone validates Portuguese phone numbers
// Portuguese mobile: +351 9XX XXX XXX
// Portuguese landline: +351 2XX XXX XXX
func ValidatePortuguesePhone(phone string) bool {
	if phone == "" {
		return true // Allow empty
	}

	// Remove spaces for validation
	cleanPhone := strings.ReplaceAll(phone, " ", "")

	// Portuguese phone format: +351 followed by 9 digits
	// Mobile: starts with 9
	// Landline: starts with 2
	re := regexp.MustCompile(`^\+351[29]\d{8}$`)
	return re.MatchString(cleanPhone)
}

// FormatPortuguesePhone formats a phone number to Portuguese standard
func FormatPortuguesePhone(phone string) string {
	if phone == "" {
		return ""
	}

	// Remove all spaces
	cleanPhone := strings.ReplaceAll(phone, " ", "")

	// If doesn't start with +351, add it
	if !strings.HasPrefix(cleanPhone, "+351") {
		// If starts with 9 or 2, assume it's Portuguese
		if len(cleanPhone) == 9 && (cleanPhone[0] == '9' || cleanPhone[0] == '2') {
			cleanPhone = "+351" + cleanPhone
		}
	}

	// Format: +351 9XX XXX XXX
	if len(cleanPhone) == 13 && strings.HasPrefix(cleanPhone, "+351") {
		return cleanPhone[:4] + " " + cleanPhone[4:7] + " " + cleanPhone[7:10] + " " + cleanPhone[10:]
	}

	return cleanPhone
}
