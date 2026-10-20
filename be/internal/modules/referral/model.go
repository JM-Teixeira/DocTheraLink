// Package referral implements the core DocTheraLink workflow: a doctor refers
// a patient to a physiotherapist, the therapist accepts and treats the patient,
// and both sides exchange notes on a shared thread until the episode closes.
//
// The package is self-contained and layered:
//
//	model.go      - GORM models, enums and request DTOs
//	repository.go - persistence (Repository interface + GORM implementation)
//	service.go    - business rules: permissions and status transitions
//	handler.go    - HTTP handlers (Gin)
//	routes.go     - route registration
package referral

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Status is the lifecycle state of a referral.
type Status string

const (
	StatusPending    Status = "pending"     // created by the doctor, waiting for a therapist
	StatusAccepted   Status = "accepted"    // a therapist took the referral
	StatusInProgress Status = "in_progress" // treatment has started
	StatusCompleted  Status = "completed"   // treatment finished, summary written
	StatusDeclined   Status = "declined"    // therapist refused (reason required)
	StatusCancelled  Status = "cancelled"   // doctor withdrew the referral
)

// IsTerminal reports whether no further transitions are possible.
func (s Status) IsTerminal() bool {
	return s == StatusCompleted || s == StatusDeclined || s == StatusCancelled
}

// Priority tells the therapist how quickly the patient should be seen.
type Priority string

const (
	PriorityRoutine Priority = "routine"
	PriorityUrgent  Priority = "urgent"
)

// NoteKind classifies entries on the doctor ↔ therapist thread.
type NoteKind string

const (
	NoteProgress NoteKind = "progress" // therapist session/progress update
	NoteQuestion NoteKind = "question" // either side asks something
	NoteReply    NoteKind = "reply"    // answer to a question
	NoteSystem   NoteKind = "system"   // generated on status changes
)

// UserSummary is the minimal view of a user embedded in referral responses.
// It maps onto the existing users table without loading sensitive columns.
//
// Summaries are read-only views of tables owned by other modules. Relation
// fields pointing at them carry `-:migration`: without it GORM's AutoMigrate
// also "migrates" the referenced table from this struct and would, for
// example, turn patients.id from uuid into text.
type UserSummary struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserType  string `json:"user_type"`
}

func (UserSummary) TableName() string { return "users" }

// PatientSummary is the minimal view of a patient embedded in referral responses.
type PatientSummary struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid"`
	Name        string     `json:"name"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	PhoneNumber *string    `json:"phone_number"`
}

func (PatientSummary) TableName() string { return "patients" }

// Referral links a patient, the referring doctor and the treating physiotherapist.
type Referral struct {
	ID          string  `json:"id" gorm:"primaryKey;type:uuid"`
	PatientID   string  `json:"patient_id" gorm:"type:uuid;not null;index"`
	FacilityID  *string `json:"facility_id" gorm:"type:uuid;index"`
	DoctorID    uint    `json:"doctor_id" gorm:"not null;index"`
	TherapistID *uint   `json:"therapist_id" gorm:"index"`

	// Clinical content written by the doctor
	Diagnosis           string   `json:"diagnosis" gorm:"not null"`
	ICDCode             *string  `json:"icd_code"`
	Reason              string   `json:"reason" gorm:"type:text;not null"`
	Precautions         *string  `json:"precautions" gorm:"type:text"`
	Goals               *string  `json:"goals" gorm:"type:text"`
	RecommendedSessions *int     `json:"recommended_sessions"`
	Priority            Priority `json:"priority" gorm:"type:varchar(16);not null;default:routine"`

	// Workflow
	Status         Status     `json:"status" gorm:"type:varchar(16);not null;default:pending;index"`
	DeclineReason  *string    `json:"decline_reason" gorm:"type:text"`
	CancelReason   *string    `json:"cancel_reason" gorm:"type:text"`
	OutcomeSummary *string    `json:"outcome_summary" gorm:"type:text"`
	AcceptedAt     *time.Time `json:"accepted_at"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`

	// Relations (read-only views, never written through)
	Patient   *PatientSummary `json:"patient,omitempty" gorm:"foreignKey:PatientID;->;-:migration"`
	Doctor    *UserSummary    `json:"doctor,omitempty" gorm:"foreignKey:DoctorID;->;-:migration"`
	Therapist *UserSummary    `json:"therapist,omitempty" gorm:"foreignKey:TherapistID;->;-:migration"`

	CreatedAt time.Time      `json:"created_date"`
	UpdatedAt time.Time      `json:"modified_date"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Referral) TableName() string { return "referrals" }

// BeforeCreate assigns a UUID so IDs are known before the insert returns.
func (r *Referral) BeforeCreate(*gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}

// Note is one message on the shared doctor ↔ therapist thread of a referral.
type Note struct {
	ID         string       `json:"id" gorm:"primaryKey;type:uuid"`
	ReferralID string       `json:"referral_id" gorm:"type:uuid;not null;index"`
	AuthorID   uint         `json:"author_id" gorm:"not null"`
	Kind       NoteKind     `json:"kind" gorm:"type:varchar(16);not null"`
	Body       string       `json:"body" gorm:"type:text;not null"`
	Author     *UserSummary `json:"author,omitempty" gorm:"foreignKey:AuthorID;->;-:migration"`
	CreatedAt  time.Time    `json:"created_date"`
}

func (Note) TableName() string { return "referral_notes" }

// BeforeCreate assigns a UUID to new notes.
func (n *Note) BeforeCreate(*gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	return nil
}

// Models lists the tables owned by this module, for database migration.
func Models() []any {
	return []any{&Referral{}, &Note{}}
}

// CreateRequest is the body of POST /referrals/.
type CreateRequest struct {
	PatientID           string   `json:"patient_id" binding:"required,uuid"`
	FacilityID          *string  `json:"facility_id" binding:"omitempty,uuid"`
	TherapistID         *uint    `json:"therapist_id"`
	Diagnosis           string   `json:"diagnosis" binding:"required,max=255"`
	ICDCode             *string  `json:"icd_code" binding:"omitempty,max=32"`
	Reason              string   `json:"reason" binding:"required"`
	Precautions         *string  `json:"precautions"`
	Goals               *string  `json:"goals"`
	RecommendedSessions *int     `json:"recommended_sessions" binding:"omitempty,min=1,max=200"`
	Priority            Priority `json:"priority" binding:"omitempty,oneof=routine urgent"`
}

// ReasonRequest carries the free-text reason for decline/cancel/complete.
type ReasonRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// NoteRequest is the body of POST /referrals/:id/notes/.
type NoteRequest struct {
	Kind NoteKind `json:"kind" binding:"omitempty,oneof=progress question reply"`
	Body string   `json:"body" binding:"required"`
}

// ListFilter narrows GET /referrals/.
type ListFilter struct {
	Status      Status `form:"status"`
	PatientID   string `form:"patient"`
	DoctorID    uint   `form:"-"`
	TherapistID uint   `form:"-"`
	// InboxForTherapist also includes unassigned pending referrals.
	InboxForTherapist bool `form:"-"`
	Limit             int  `form:"limit"`
	Offset            int  `form:"offset"`
}
