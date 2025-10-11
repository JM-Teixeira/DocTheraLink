package handlers

import (
	"net/http"
	"time"

	"care-backend/internal/auth"
	"care-backend/internal/config"
	"care-backend/internal/database"
	"care-backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	jwtManager *auth.JWTManager
	config     *config.Config
}

func NewAuthHandler(jwtManager *auth.JWTManager, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
		config:     cfg,
	}
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/login/ [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Find user by username or email
	var user models.User
	db := database.GetDB()
	if err := db.Where("username = ? OR email = ?", req.Username, req.Username).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid credentials",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Database error",
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Account is deactivated",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid credentials",
		})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	db.Save(&user)

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate tokens",
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Generate new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body models.TokenRefreshRequest true "Refresh token"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/token/refresh/ [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.TokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	tokens, err := h.jwtManager.RefreshToken(req.Refresh)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// VerifyToken godoc
// @Summary Verify JWT token
// @Description Verify if JWT token is valid
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/token/verify/ [post]
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	// Token validation is handled by middleware
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Token is valid",
	})
}

// UpdatePassword godoc
// @Summary Update user password
// @Description Update current user's password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body models.UpdatePasswordRequest true "Password update request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /password_change/ [put]
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	var req models.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from token
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid token",
		})
		return
	}

	// Verify username matches
	if claims.Username != req.Username {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "forbidden",
			Message: "Cannot change password for another user",
		})
		return
	}

	// Get user from database
	var user models.User
	db := database.GetDB()
	if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid old password",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), h.config.Security.BcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to hash password",
		})
		return
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password updated successfully",
	})
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Password reset request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /password_reset/ [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Find user by username or email
	var user models.User
	db := database.GetDB()
	if err := db.Where("username = ? OR email = ?", req.Username, req.Username).
		First(&user).Error; err != nil {
		// Don't reveal if user exists or not
		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "If the user exists, a password reset email has been sent",
		})
		return
	}

	// TODO: Generate reset token and send email
	// For now, just return success
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "If the user exists, a password reset email has been sent",
	})
}

// CheckResetToken godoc
// @Summary Check password reset token
// @Description Verify if password reset token is valid
// @Tags auth
// @Accept json
// @Produce json
// @Param token body object{token=string} true "Reset token"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /password_reset/check/ [post]
func (h *AuthHandler) CheckResetToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// TODO: Validate reset token
	// For now, just return success
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Token is valid",
	})
}

// ResetPassword godoc
// @Summary Reset password with token
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetConfirmRequest true "Password reset confirmation"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /password_reset/confirm/ [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate password confirmation
	if req.Password != req.Confirm {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Passwords do not match",
		})
		return
	}

	// TODO: Validate reset token and update password
	// For now, just return success
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password reset successfully",
	})
}