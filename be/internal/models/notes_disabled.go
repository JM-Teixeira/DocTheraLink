package models

// DISABLED MODELS - These models have GORM relationship conflicts
// They are temporarily disabled until the constraint issues are resolved

/*
// PatientNote - DISABLED due to GORM foreign key constraint conflicts
// Issue: GORM automatically creates conflicting constraints to both 'patients' and 'therapy_patients' tables
type PatientNoteDisabled struct {
	tableName struct{} `gorm:"table:patient_notes"`
	ID             string           `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Note           string           `json:"note" gorm:"not null"`
	Thread         string           `json:"thread" gorm:"not null"`
	PatientID      string           `json:"-" gorm:"type:uuid"`
	ConsultationID *string          `json:"consultation" gorm:"type:uuid"`
	CreatedByID    uint             `json:"-"`
	ReplyToID      *string          `json:"reply_to" gorm:"type:uuid"`
	
	Patient      Patient          `json:"-" gorm:"foreignKey:PatientID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Consultation *Consultation    `json:"-" gorm:"foreignKey:ConsultationID;references:ID"`
	CreatedBy    UserBareMinimum  `json:"created_by" gorm:"foreignKey:CreatedByID;references:ID"`
	
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// TherapyPatient - DISABLED due to relationship conflicts with PatientNote
type TherapyPatientDisabled struct {
	tableName struct{} `gorm:"table:therapy_patients"`
	ID                string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UniqueID          string    `json:"unique_id" gorm:"uniqueIndex;not null"`
	Name              string    `json:"name" gorm:"not null"`
	// ... other fields would go here
}
*/

// TODO: Fix GORM relationship conflicts and re-enable these models
// The main issue is that GORM automatically detects relationships between models
// with similar field names and creates conflicting foreign key constraints.
//
// Solutions to try:
// 1. Use explicit table names and constraint names
// 2. Disable automatic foreign key creation
// 3. Create separate migration phases
// 4. Use manual SQL for complex relationships