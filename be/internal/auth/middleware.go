package auth

import (
	"net/http"
	"strings"

	"care-backend/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
	CurrentUserKey          = "current_user"
)

// AuthMiddleware creates a gin middleware for authorization
func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization header is not provided",
			})
			ctx.Abort()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
			ctx.Abort()
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Unsupported authorization type",
			})
			ctx.Abort()
			return
		}

		accessToken := fields[1]
		payload, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: err.Error(),
			})
			ctx.Abort()
			return
		}

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}

// OptionalAuthMiddleware creates a gin middleware for optional authorization
func OptionalAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.Next()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.Next()
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			ctx.Next()
			return
		}

		accessToken := fields[1]
		payload, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			ctx.Next()
			return
		}

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}

// RequireRole creates a middleware that requires specific user roles
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload, exists := ctx.Get(AuthorizationPayloadKey)
		if !exists {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization required",
			})
			ctx.Abort()
			return
		}

		claims, ok := payload.(*models.TokenClaims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization payload",
			})
			ctx.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if claims.UserType == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			ctx.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "Insufficient permissions",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// GetCurrentUserClaims extracts the current user claims from the context
func GetCurrentUserClaims(ctx *gin.Context) (*models.TokenClaims, error) {
	payload, exists := ctx.Get(AuthorizationPayloadKey)
	if !exists {
		return nil, ErrInvalidToken
	}

	claims, ok := payload.(*models.TokenClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// IsAdmin checks if the current user is an admin
func IsAdmin(userType models.UserRole) bool {
	adminRoles := []models.UserRole{
		models.RoleStateAdmin,
		models.RoleDistrictAdmin,
		models.RoleLocalBodyAdmin,
	}

	for _, role := range adminRoles {
		if userType == role {
			return true
		}
	}
	return false
}

// IsReadOnlyUser checks if the current user has read-only access
func IsReadOnlyUser(userType models.UserRole) bool {
	readOnlyRoles := []models.UserRole{
		models.RoleStaffReadOnly,
		models.RoleDistrictReadOnlyAdmin,
		models.RoleStateReadOnlyAdmin,
	}

	for _, role := range readOnlyRoles {
		if userType == role {
			return true
		}
	}
	return false
}

// CanAccessFacility checks if user can access a specific facility
func CanAccessFacility(userType models.UserRole, userFacilityID *string, targetFacilityID string) bool {
	// Admins can access all facilities
	if IsAdmin(userType) {
		return true
	}

	// Users can access their home facility
	if userFacilityID != nil && *userFacilityID == targetFacilityID {
		return true
	}

	return false
}

// CanModifyResource checks if user can modify a resource
func CanModifyResource(userType models.UserRole) bool {
	// Read-only users cannot modify resources
	if IsReadOnlyUser(userType) {
		return false
	}

	return true
}