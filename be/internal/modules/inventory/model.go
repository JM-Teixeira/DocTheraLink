// Package inventory tracks a clinic's consumables: what is in stock, who
// used, opened or restocked what, and what is running low.
//
// Two kinds of item, each with its own actions:
//
//	unit      - single-use items (gloves, needles, masks, tape). "Use" takes
//	            them out of stock. They can be used automatically per
//	            session on the time board (per division if needed).
//	container - multi-use items (gel bottles, oil, sanitiser, wipes, couch
//	            roll). "Open" moves a sealed one into use, recording who
//	            opened it and when; "Finish" records that an open one is
//	            empty. OpenLifeDays flags containers open for too long.
//
// Layout follows the other modules: model, repository, service, handler,
// routes, seed.
package inventory

import "time"

// Movement kinds.
const (
	KindRestock = "restock" // stock arrived: +amount sealed
	KindUse     = "use"     // unit items used: -amount
	KindOpen    = "open"    // containers: -amount sealed, +amount in use
	KindFinish  = "finish"  // containers: -amount in use (emptied)
	KindAdjust  = "adjust"  // physical count: sealed stock becomes amount
)

// Item kinds.
const (
	TrackUnit      = "unit"
	TrackContainer = "container"
)

// Item is one consumable a clinic keeps in stock.
type Item struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ClinicID string `json:"clinic_id" gorm:"type:uuid;not null;index"`
	Name     string `json:"name" gorm:"size:120;not null"`
	Category string `json:"category" gorm:"size:60"`
	Unit     string `json:"unit" gorm:"size:30;not null;default:units"`
	Tracking string `json:"tracking" gorm:"size:10;not null;default:unit"` // unit | container

	Quantity       int `json:"quantity" gorm:"not null;default:0"`        // sealed / available
	MinQuantity    int `json:"min_quantity" gorm:"not null;default:0"`    // alert at or below this
	TargetQuantity int `json:"target_quantity" gorm:"not null;default:0"` // reorder up to this (0 = 2 × minimum)

	// Containers
	InUse        int        `json:"in_use" gorm:"not null;default:0"`         // opened, not finished
	OldestOpenAt *time.Time `json:"oldest_open_at"`                           // when the oldest open one was opened
	OpenLifeDays int        `json:"open_life_days" gorm:"not null;default:0"` // discard after this many days open (0 = no limit)

	// Units: used automatically per attended session (Rules override per division)
	PerSession int           `json:"per_session" gorm:"not null;default:0"`
	Rules      []SessionRule `json:"session_rules" gorm:"foreignKey:ItemID"`

	Archived  bool      `json:"archived" gorm:"not null;default:false;index"`
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
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

// PerSessionIn is how much is used per session in a division: the
// division's rule if there is one, otherwise the default.
func (i Item) PerSessionIn(divisionID string) int {
	for _, r := range i.Rules {
		if r.DivisionID == divisionID {
			return r.Quantity
		}
	}
	return i.PerSession
}

// ReorderQuantity suggests how many to order to get back to the target.
func (i Item) ReorderQuantity() int {
	target := i.TargetQuantity
	if target == 0 {
		target = max(i.MinQuantity*2, 1)
	}
	return max(target-i.Quantity, 0)
}

// OpenExpired reports a container open longer than its open life.
func (i Item) OpenExpired(now time.Time) bool {
	return i.InUse > 0 && i.OpenLifeDays > 0 && i.OldestOpenAt != nil &&
		now.Sub(*i.OldestOpenAt) > time.Duration(i.OpenLifeDays)*24*time.Hour
}

// ItemView is an item as returned by the API.
type ItemView struct {
	Item
	Status      string `json:"status"`
	OpenExpired bool   `json:"open_expired"`
	Reorder     int    `json:"reorder_quantity"`
}

func view(i Item, now time.Time) ItemView {
	if i.Rules == nil {
		i.Rules = []SessionRule{}
	}
	return ItemView{Item: i, Status: i.Status(), OpenExpired: i.OpenExpired(now), Reorder: i.ReorderQuantity()}
}

// SessionRule sets the per-session quantity of a unit item in one division
// (e.g. 2 pairs of gloves in the pool, 0 in the gym).
type SessionRule struct {
	ID         string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ItemID     string `json:"item_id" gorm:"type:uuid;not null;uniqueIndex:uniq_rule_item_division"`
	DivisionID string `json:"division_id" gorm:"type:uuid;not null;uniqueIndex:uniq_rule_item_division"`
	Quantity   int    `json:"quantity" gorm:"not null"`
}

func (SessionRule) TableName() string { return "stock_session_rules" }

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
	Amount        int          `json:"amount" gorm:"not null;default:0"` // how many the action concerned (positive)
	Delta         int          `json:"delta" gorm:"not null"`            // change of the sealed/available stock
	QuantityAfter int          `json:"quantity_after" gorm:"not null"`
	InUseAfter    int          `json:"in_use_after" gorm:"not null;default:0"`
	Note          string       `json:"note" gorm:"size:300"`
	UserID        uint         `json:"user_id" gorm:"not null;index"`
	AppointmentID *string      `json:"appointment_id" gorm:"type:uuid;index"` // set for automatic session use
	User          *UserSummary `json:"user,omitempty" gorm:"foreignKey:UserID;->;-:migration"`
	Item          *ItemName    `json:"item,omitempty" gorm:"foreignKey:ItemID;->;-:migration"`
	CreatedAt     time.Time    `json:"created_date" gorm:"index"`
}

func (Movement) TableName() string { return "stock_movements" }

// ItemName is the item shown next to a movement in the audit log.
type ItemName struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid"`
	Name     string `json:"name"`
	Unit     string `json:"unit"`
	Category string `json:"category"`
	Tracking string `json:"tracking"`
}

func (ItemName) TableName() string { return "stock_items" }

// Models lists the tables for AutoMigrate.
func Models() []any { return []any{&Item{}, &Movement{}, &SessionRule{}} }

// ItemRequest creates or edits an item. Quantity only applies on create
// (the opening stock); later changes go through movements.
type ItemRequest struct {
	Name           string  `json:"name" binding:"required,max=120"`
	Category       string  `json:"category" binding:"max=60"`
	Unit           string  `json:"unit" binding:"max=30"`
	Tracking       string  `json:"tracking" binding:"omitempty,oneof=unit container"`
	MinQuantity    int     `json:"min_quantity" binding:"min=0,max=100000"`
	TargetQuantity int     `json:"target_quantity" binding:"min=0,max=100000"`
	PerSession     int     `json:"per_session" binding:"min=0,max=100"`
	OpenLifeDays   int     `json:"open_life_days" binding:"min=0,max=3650"`
	Quantity       int     `json:"quantity" binding:"min=0,max=100000"`
	ClinicID       *string `json:"clinic_id" binding:"omitempty,uuid"` // admins only
}

// MovementRequest records stock arriving, being used, opened, finished or counted.
type MovementRequest struct {
	Kind     string `json:"kind" binding:"required,oneof=restock use adjust open finish"`
	Quantity int    `json:"quantity" binding:"min=0,max=100000"`
	Note     string `json:"note" binding:"max=300"`
}

// SessionRulesRequest replaces an item's per-division session quantities.
type SessionRulesRequest struct {
	Rules []RuleInput `json:"rules" binding:"dive"`
}

// RuleInput is one division's per-session quantity.
type RuleInput struct {
	DivisionID string `json:"division_id" binding:"required,uuid"`
	Quantity   int    `json:"quantity" binding:"min=0,max=100"`
}

// Summary counts items per status for the page header.
type Summary struct {
	Items       int `json:"items"`
	Low         int `json:"low"`
	Out         int `json:"out"`
	OpenExpired int `json:"open_expired"`
}

// ListFilter narrows the item list.
type ListFilter struct {
	Search   string `form:"search"`
	Category string `form:"category"`
	Status   string `form:"status" binding:"omitempty,oneof=ok low out attention"` // attention = low or out
}

// AuditFilter narrows the audit log of movements.
type AuditFilter struct {
	ItemID string     `form:"item" binding:"omitempty,uuid"`
	UserID uint       `form:"user"`
	Kind   string     `form:"kind" binding:"omitempty,oneof=restock use adjust open finish"`
	From   *time.Time `form:"from" time_format:"2006-01-02"`
	To     *time.Time `form:"to" time_format:"2006-01-02"`
	Auto   *bool      `form:"auto"` // true: only automatic session use
}
