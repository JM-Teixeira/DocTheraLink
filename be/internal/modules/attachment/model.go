// Package attachment stores files attached to referrals (imaging, reports,
// PDFs). Access follows the referral: whoever can see the referral can
// download its files; only participants can upload.
//
//	model.go      - Attachment table
//	repository.go - persistence
//	service.go    - access rules, validation, storage
//	handler.go    - HTTP (multipart upload, streaming download)
//	routes.go     - route registration
package attachment

import (
	"time"

	"klyniva/internal/modules/referral"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Attachment is the metadata of one stored file.
type Attachment struct {
	ID          string `json:"id" gorm:"primaryKey;type:uuid"`
	ReferralID  string `json:"referral_id" gorm:"type:uuid;not null;index"`
	UploaderID  uint   `json:"uploader_id" gorm:"not null"`
	FileName    string `json:"file_name" gorm:"type:varchar(255);not null"`
	ContentType string `json:"content_type" gorm:"type:varchar(100);not null"`
	Size        int64  `json:"size" gorm:"not null"`
	StorageKey  string `json:"-" gorm:"type:varchar(255);not null"`

	// Read-only view of the users table (see the -:migration note in referral/model.go)
	Uploader *referral.UserSummary `json:"uploader,omitempty" gorm:"foreignKey:UploaderID;->;-:migration"`

	CreatedAt time.Time      `json:"created_date"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Attachment) TableName() string { return "referral_attachments" }

// BeforeCreate assigns a UUID.
func (a *Attachment) BeforeCreate(*gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}

// Models lists the tables owned by this module, for database migration.
func Models() []any { return []any{&Attachment{}} }
