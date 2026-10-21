// Package notification tells people about things that concern them: in-app
// notifications (bell in the UI) and, when SMTP is configured, an e-mail.
//
// It implements referral.Notifier, so the referral module publishes events
// without knowing who is notified or how.
//
//	model.go      - Notification table and API DTOs
//	repository.go - persistence + recipient lookups
//	service.go    - who gets notified, e-mail content
//	handler.go    - HTTP API for the bell
//	routes.go     - route registration
package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification is one in-app notification for one user. Display text is
// built by the frontend from Event + the snapshot fields (translated).
type Notification struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid"`
	UserID      uint       `json:"-" gorm:"not null;index:idx_user_notifications_user_read,priority:1"`
	Event       string     `json:"event" gorm:"type:varchar(32);not null"`
	ReferralID  *string    `json:"referral_id" gorm:"type:uuid;index"`
	ActorID     *uint      `json:"actor_id"`
	ActorName   string     `json:"actor_name" gorm:"type:varchar(255)"`
	PatientName string     `json:"patient_name" gorm:"type:varchar(255)"`
	Excerpt     string     `json:"excerpt" gorm:"type:varchar(200)"`
	ReadAt      *time.Time `json:"read_at" gorm:"index:idx_user_notifications_user_read,priority:2"`
	CreatedAt   time.Time  `json:"created_date"`
}

func (Notification) TableName() string { return "user_notifications" }

// BeforeCreate assigns a UUID.
func (n *Notification) BeforeCreate(*gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	return nil
}

// Models lists the tables owned by this module, for database migration.
func Models() []any { return []any{&Notification{}} }

// ListQuery holds GET /notifications/ parameters.
type ListQuery struct {
	Unread bool `form:"unread"`
	Limit  int  `form:"limit"`
	Offset int  `form:"offset"`
}

// Page is the response of GET /notifications/.
type Page struct {
	Count   int64          `json:"count"`
	Unread  int64          `json:"unread"`
	Results []Notification `json:"results"`
}

// Person is a recipient: in-app by ID, e-mail by address.
type Person struct {
	ID        uint
	Email     string
	FirstName string
	LastName  string
}

// FullName is "First Last".
func (p Person) FullName() string { return p.FirstName + " " + p.LastName }
