// Package inventory tracks a clinic's consumables (masks, gloves, gel,
// tape…): what is in stock, what was used or restocked, and what is running
// low.
//
// Layout follows the other modules:
//
//	model.go      - tables and request bodies
//	repository.go - persistence (Repository interface + GORM implementation)
//	service.go    - clinic scoping, permissions and stock rules
//	handler.go    - HTTP handlers
//	routes.go     - route registration
package inventory

import "time"

// Movement kinds.
const (
	KindRestock = "restock" // stock arrived: +quantity
	KindUse     = "use"     // used in treatment: -quantity
	KindAdjust  = "adjust"  // physical count: stock becomes quantity
	KindOpen    = "open"    // a pack/bottle opened for use: -quantity, audited separately
)

// Item is one consumable a clinic keeps in stock.
type Item struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ClinicID    string    `json:"clinic_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"size:120;not null"`
	Category    string    `json:"category" gorm:"size:60"`
	Unit        string    `json:"unit" gorm:"size:30;not null;default:units"`
	Quantity    int       `json:"quantity" gorm:"not null;default:0"`
	MinQuantity int       `json:"min_quantity" gorm:"not null;default:0"` // alert at or below this
	PerSession  int       `json:"per_session" gorm:"not null;default:0"`  // used automatically per attended session
	Archived    bool      `json:"archived" gorm:"not null;default:false;index"`
	CreatedAt   time.Time `json:"created_date"`
	UpdatedAt   time.Time `json:"modified_date"`
}

func (Item) TableName() string { return "stock_items" }

// Status is "ok", "low" (at or below the minimum) or "out" (none left).
func (i Item) Status() string {
	switch {
	case i.Quantity <= 0:
		return "out"
	case i.Quantity <= i.MinQuantity:
		return "low"
	}
	return "ok"
}

// ItemView is an item as returned by the API, with its stock status.
type ItemView struct {
	Item
	Status string `json:"status"`
}

func view(i Item) ItemView { return ItemView{Item: i, Status: i.Status()} }

// UserSummary is who recorded a movement.
type UserSummary struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (UserSummary) TableName() string { return "users" }

// Movement is one change of stock, kept as history.
type Movement struct {
	ID            string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ItemID        string       `json:"item_id" gorm:"type:uuid;not null;index"`
	Kind          string       `json:"kind" gorm:"size:10;not null"`
	Delta         int          `json:"delta" gorm:"not null"`
	QuantityAfter int          `json:"quantity_after" gorm:"not null"`
	Note          string       `json:"note" gorm:"size:300"`
	UserID        uint         `json:"user_id" gorm:"not null;index"`
	AppointmentID *string      `json:"appointment_id" gorm:"type:uuid;index"` // set for automatic session use
	Item          *ItemName    `json:"item,omitempty" gorm:"foreignKey:ItemID;->;-:migration"`
	User          *UserSummary `json:"user,omitempty" gorm:"foreignKey:UserID;->;-:migration"`
	CreatedAt     time.Time    `json:"created_date" gorm:"index"`
}

func (Movement) TableName() string { return "stock_movements" }

// ItemName is the item shown next to a movement in the audit log.
type ItemName struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid"`
	Name     string `json:"name"`
	Unit     string `json:"unit"`
	Category string `json:"category"`
}

func (ItemName) TableName() string { return "stock_items" }

// AuditFilter narrows the audit log of movements.
type AuditFilter struct {
	ItemID string     `form:"item" binding:"omitempty,uuid"`
	UserID uint       `form:"user"`
	Kind   string     `form:"kind" binding:"omitempty,oneof=restock use adjust open"`
	From   *time.Time `form:"from" time_format:"2006-01-02"`
	To     *time.Time `form:"to" time_format:"2006-01-02"`
	Auto   *bool      `form:"auto"` // true: only automatic session use
}

// Models lists the tables for AutoMigrate.
func Models() []any { return []any{&Item{}, &Movement{}} }

// ItemRequest creates or edits an item. Quantity only applies on create
// (the opening stock); later changes go through movements.
type ItemRequest struct {
	Name        string  `json:"name" binding:"required,max=120"`
	Category    string  `json:"category" binding:"max=60"`
	Unit        string  `json:"unit" binding:"max=30"`
	MinQuantity int     `json:"min_quantity" binding:"min=0,max=100000"`
	PerSession  int     `json:"per_session" binding:"min=0,max=100"`
	Quantity    int     `json:"quantity" binding:"min=0,max=100000"`
	ClinicID    *string `json:"clinic_id" binding:"omitempty,uuid"` // admins only
}

// MovementRequest records stock arriving, being used or counted.
type MovementRequest struct {
	Kind     string `json:"kind" binding:"required,oneof=restock use adjust open"`
	Quantity int    `json:"quantity" binding:"min=0,max=100000"`
	Note     string `json:"note" binding:"max=300"`
}

// Summary counts items per status for the page header.
type Summary struct {
	Items int `json:"items"`
	Low   int `json:"low"`
	Out   int `json:"out"`
}

// ListFilter narrows the item list.
type ListFilter struct {
	Search   string `form:"search"`
	Category string `form:"category"`
	Status   string `form:"status" binding:"omitempty,oneof=ok low out attention"` // attention = low or out
}
