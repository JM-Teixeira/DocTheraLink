package handlers

// Who may see and change users and clinics through the legacy handlers
// (pure rules, tested in access_test.go). Admins manage everything;
// coordinators the people of their own clinic; everyone their own profile.

import (
	"errors"
	"net/http"
	"strings"

	"klyniva/internal/auth"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// actor is the signed-in user with their clinic.
type actor struct {
	ID     uint
	Role   models.UserRole
	Clinic *string
}

var (
	errForbidden = errors.New("forbidden")
	errBadField  = errors.New("bad field")
)

// currentActor loads the caller; it answers 401 itself when there is none.
func currentActor(c *gin.Context, db *gorm.DB) (*actor, bool) {
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "User not authenticated"})
		return nil, false
	}
	var me models.User
	if err := db.Select("id", "user_type", "home_facility_id").First(&me, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "User not authenticated"})
		return nil, false
	}
	return &actor{ID: me.ID, Role: me.UserType, Clinic: me.HomeFacilityID}, true
}

func sameClinic(a, b *string) bool { return a != nil && b != nil && *a == *b }

// canSeeUser: admins, the person and the people of the same clinic.
func canSeeUser(a *actor, target *models.User) bool {
	return auth.IsAdmin(a.Role) || a.ID == target.ID || sameClinic(a.Clinic, target.HomeFacilityID)
}

// managesUser: admins, or the coordinator of the person's clinic.
func managesUser(a *actor, target *models.User) bool {
	return auth.IsAdmin(a.Role) || (a.Role == models.RoleCoordinator && sameClinic(a.Clinic, target.HomeFacilityID))
}

// Roles a coordinator may give: the clinic's own roles, never an admin one.
var coordinatorRoles = map[models.UserRole]bool{
	models.RoleDoctor: true, models.RolePhysioTherapist: true, models.RoleReceptionist: true,
	models.RoleNurse: true, models.RoleCoordinator: true, models.RoleStaff: true, models.RoleStaffReadOnly: true,
}

var allRoles = map[models.UserRole]bool{
	models.RolePharmacist: true, models.RoleVolunteer: true, models.RoleStaffReadOnly: true, models.RoleStaff: true,
	models.RoleReceptionist: true, models.RoleCoordinator: true, models.RoleNurse: true, models.RolePhysioTherapist: true,
	models.RoleDoctor: true, models.RoleWardAdmin: true, models.RoleLocalBodyAdmin: true, models.RoleDistrictLabAdmin: true,
	models.RoleDistrictReadOnlyAdmin: true, models.RoleDistrictAdmin: true, models.RoleStateLabAdmin: true,
	models.RoleStateReadOnlyAdmin: true, models.RoleStateAdmin: true,
}

// canAssignRole: admins any real role, coordinators the clinic roles.
func canAssignRole(a *actor, role models.UserRole) bool {
	if auth.IsAdmin(a.Role) {
		return allRoles[role]
	}
	return a.Role == models.RoleCoordinator && coordinatorRoles[role]
}

// Profile fields anyone may change on their own profile (JSON key → column).
var profileColumns = map[string]string{
	"first_name": "first_name", "last_name": "last_name", "email": "email",
	"phone_number": "phone_number", "alt_phone_number": "alt_phone_number",
	"gender": "gender", "date_of_birth": "date_of_birth", "video_connect_link": "video_connect_link",
	"qualification": "qualification", "doctor_experience_commenced_on": "doctor_experience_commenced_on",
	"doctor_medical_council_registration": "doctor_medical_council_registration",
	"weekly_working_hours":                "weekly_working_hours",
	"state":                               "state_id", "district": "district_id", "local_body": "local_body_id",
}

// Fields with their own endpoint and rules; sending them here is refused
// rather than silently ignored, so an attempt is visible.
var protectedKeys = map[string]bool{
	"password": true, "is_superuser": true, "is_active": true, "verified": true, "username": true,
	"home_facility": true, "employment_status": true, "employment_since": true, "employment_note": true,
	"date_format": true, "last_login": true, "id": true,
}

// userUpdate turns a PATCH body into the columns to save. Profile fields
// need self or manager; user_type needs a manager allowed to give that role.
// Unknown keys (read-only display fields the forms send back) are dropped.
func userUpdate(a *actor, target *models.User, body map[string]any) (map[string]any, error) {
	self, manager := a.ID == target.ID, managesUser(a, target)
	if !self && !manager {
		return nil, errForbidden
	}
	out := map[string]any{}
	for k, v := range body {
		switch {
		case protectedKeys[k]:
			return nil, errForbidden
		case k == "user_type":
			role, _ := v.(string)
			if models.UserRole(role) == target.UserType {
				continue
			}
			if !manager || self || !canAssignRole(a, models.UserRole(role)) {
				return nil, errForbidden
			}
			out["user_type"] = role
		case profileColumns[k] != "":
			if k == "email" {
				if s, _ := v.(string); !strings.Contains(s, "@") {
					return nil, errBadField
				}
			}
			out[profileColumns[k]] = v
		}
	}
	return out, nil
}

// newUserRequest is what the "add person" form sends.
type newUserRequest struct {
	Username      string          `json:"username" binding:"required,min=3,max=150"`
	Password      string          `json:"password" binding:"required,min=8,max=128"`
	Email         string          `json:"email" binding:"required,email"`
	FirstName     string          `json:"first_name" binding:"required,max=150"`
	LastName      string          `json:"last_name" binding:"required,max=150"`
	UserType      models.UserRole `json:"user_type" binding:"required"`
	HomeFacility  *string         `json:"home_facility"`
	PhoneNumber   *string         `json:"phone_number"`
	Gender        *string         `json:"gender"`
	Qualification *string         `json:"qualification"`
	StateID       *uint           `json:"state"`
	DistrictID    *uint           `json:"district"`
	LocalBodyID   *uint           `json:"local_body"`
}

// newUserClinic decides the clinic of a new user: coordinators always add
// to their own clinic; admins choose.
func newUserClinic(a *actor, req *newUserRequest) (*string, error) {
	if !canAssignRole(a, req.UserType) {
		return nil, errForbidden
	}
	if auth.IsAdmin(a.Role) {
		if req.HomeFacility != nil && *req.HomeFacility == "" {
			return nil, nil
		}
		return req.HomeFacility, nil
	}
	if a.Clinic == nil {
		return nil, errForbidden
	}
	return a.Clinic, nil
}

// canManageClinic: admins any clinic, coordinators their own (details only).
func canManageClinic(a *actor, clinicID string) bool {
	return auth.IsAdmin(a.Role) || (a.Role == models.RoleCoordinator && a.Clinic != nil && *a.Clinic == clinicID)
}

// canSeeClinic: admins any clinic, everyone else their own.
func canSeeClinic(a *actor, clinicID string) bool {
	return auth.IsAdmin(a.Role) || (a.Clinic != nil && *a.Clinic == clinicID)
}

// Clinic fields a PATCH may change (JSON key → column).
// (features is an array column: change it with PUT)
var clinicColumns = map[string]string{
	"name": "name", "address": "address", "pincode": "pincode", "middleware_address": "middleware_address",
	"latitude": "latitude", "longitude": "longitude", "state": "state_id", "district": "district_id",
	"facility_type": "facility_type",
}

func clinicUpdate(body map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range body {
		if col := clinicColumns[k]; col != "" {
			out[col] = v
		}
	}
	return out
}

// pageLimit caps list sizes (default 20, at most 100).
func pageLimit(n int) int {
	if n <= 0 || n > 100 {
		return 20
	}
	return n
}

// forbid answers 403 with a short reason.
func forbid(c *gin.Context) {
	c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "You do not have permission to perform this action"})
}

// notFound answers 404 (also used when the caller may not see the record).
func notFound(c *gin.Context, what string) {
	c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: what + " not found"})
}
