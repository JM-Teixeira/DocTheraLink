package models

import "time"

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Count    int         `json:"count"`
	Next     *string     `json:"next"`
	Previous *string     `json:"previous"`
	Results  interface{} `json:"results"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// QueryParams represents common query parameters for list endpoints
type QueryParams struct {
	Limit    int    `form:"limit"`
	Offset   int    `form:"offset"`
	Search   string `form:"search"`
	Ordering string `form:"ordering"`
}

// DateRange represents a date range filter
type DateRange struct {
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date" time_format:"2006-01-02"`
}

// BaseModel represents common fields for all models
type BaseModel struct {
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// SoftDeleteModel includes soft delete functionality
type SoftDeleteModel struct {
	BaseModel
	DeletedAt *time.Time `json:"-" gorm:"index"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	UserType UserRole `json:"user_type"`
	Exp      int64    `json:"exp"`
	Iat      int64    `json:"iat"`
}

// HealthCheck represents health check response
type HealthCheck struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Database  string    `json:"database"`
}

// FileUpload represents file upload information
type FileUpload struct {
	ID              string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FileName        string  `json:"file_name" gorm:"not null"`
	FileType        string  `json:"file_type" gorm:"not null"`
	FileSize        int64   `json:"file_size"`
	FilePath        string  `json:"file_path" gorm:"not null"`
	MimeType        string  `json:"mime_type"`
	AssociatingID   *string `json:"associating_id"`
	AssociationType *string `json:"associating_type"`
	UploadedByID    uint    `json:"-"`
	IsDeleted       bool    `json:"is_deleted" gorm:"default:false"`

	UploadedBy UserBareMinimum `json:"uploaded_by" gorm:"foreignKey:UploadedByID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// Notification represents system notifications
type Notification struct {
	ID               string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title            string `json:"title" gorm:"not null"`
	Message          string `json:"message" gorm:"not null"`
	NotificationType string `json:"notification_type" gorm:"not null"`
	IsRead           bool   `json:"is_read" gorm:"default:false"`
	UserID           uint   `json:"-"`

	User User `json:"-" gorm:"foreignKey:UserID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// Inventory represents facility inventory items
type InventoryItem struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	Name        string  `json:"name" gorm:"not null"`
	Description *string `json:"description"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit" gorm:"default:'nos'"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type InventoryLog struct {
	ID               string `json:"external_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FacilityID       string `json:"-"`
	ItemID           uint   `json:"-"`
	QuantityInStock  int    `json:"quantity_in_stock"`
	QuantityAdded    int    `json:"quantity_added"`
	QuantityConsumed int    `json:"quantity_consumed"`
	IsFlagged        bool   `json:"is_flagged" gorm:"default:false"`
	CreatedByID      uint   `json:"-"`

	Facility  Facility        `json:"-" gorm:"foreignKey:FacilityID"`
	Item      InventoryItem   `json:"item" gorm:"foreignKey:ItemID"`
	CreatedBy UserBareMinimum `json:"created_by" gorm:"foreignKey:CreatedByID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}
