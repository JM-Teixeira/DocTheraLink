package auth

import (
	"net/http"

	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// "View as": an admin sees the app as a clinic role would, to check what
// each role can see. The browser sends the role and clinic in headers;
// the request then runs with that role and clinic, and only reads: anything
// that changes data is refused, so looking around can never act as someone.
const (
	// WorkingClinicHeader: the clinic an admin works on (the time board of
	// that clinic); admins have no clinic of their own.
	WorkingClinicHeader = "X-Clinic"
	ViewAsRoleHeader    = "X-View-As-Role"
	ViewAsClinicHeader  = "X-View-As-Clinic"
	viewAsContextKey    = "view_as"
)

// ViewAsRoles are the roles an admin may view as.
var ViewAsRoles = map[models.UserRole]bool{
	models.RoleDoctor: true, models.RolePhysioTherapist: true, models.RoleCoordinator: true,
	models.RoleReceptionist: true, models.RoleNurse: true, models.RoleStaff: true, models.RoleStaffReadOnly: true,
}

// ViewAs describes an active "view as" on a request.
type ViewAs struct {
	Role     models.UserRole `json:"role"`
	Clinic   string          `json:"clinic"`
	RealRole models.UserRole `json:"real_role"`
}

// applyViewAs switches the claims to the viewed role when the headers ask
// for it. It answers and returns false when the request must stop.
func applyViewAs(ctx *gin.Context, claims *models.TokenClaims) bool {
	role := models.UserRole(ctx.GetHeader(ViewAsRoleHeader))
	if role == "" {
		// An admin working on one clinic (full rights, just that clinic)
		if id := ctx.GetHeader(WorkingClinicHeader); id != "" && IsAdmin(claims.UserType) && uuid.Validate(id) == nil {
			ctx.Request = ctx.Request.WithContext(clinic.WithViewAs(ctx.Request.Context(), claims.UserID, id))
		}
		return true
	}
	refuse := func(status int, code, msg string) bool {
		ctx.JSON(status, models.ErrorResponse{Error: code, Message: msg})
		ctx.Abort()
		return false
	}
	if !IsAdmin(claims.UserType) {
		return refuse(http.StatusForbidden, "forbidden", "only administrators can view as another role")
	}
	clinicID := ctx.GetHeader(ViewAsClinicHeader)
	if !ViewAsRoles[role] || uuid.Validate(clinicID) != nil {
		return refuse(http.StatusBadRequest, "bad_request", "unknown role or clinic to view as")
	}
	switch ctx.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
	default:
		return refuse(http.StatusForbidden, "view_as_read_only", "viewing as another role is read-only")
	}
	ctx.Set(viewAsContextKey, ViewAs{Role: role, Clinic: clinicID, RealRole: claims.UserType})
	claims.UserType = role
	ctx.Request = ctx.Request.WithContext(clinic.WithViewAs(ctx.Request.Context(), claims.UserID, clinicID))
	return true
}

// ViewingAs returns the active "view as" of the request, if any.
func ViewingAs(ctx *gin.Context) (ViewAs, bool) {
	v, ok := ctx.Get(viewAsContextKey)
	if !ok {
		return ViewAs{}, false
	}
	va, ok := v.(ViewAs)
	return va, ok
}
