package handlers

import (
	"net/http"

	"care-backend/internal/auth"
	"care-backend/internal/config"
	"care-backend/internal/database"
	"care-backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	config *config.Config
}

func NewUserHandler(cfg *config.Config) *UserHandler {
	return &UserHandler{
		config: cfg,
	}
}

// ListUsers godoc
// @Summary List users
// @Description Get paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of results per page"
// @Param offset query int false "Number of results to skip"
// @Param search query string false "Search term"
// @Param user_type query string false "Filter by user type"
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/ [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var params models.QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid query parameters",
		})
		return
	}

	// Set defaults
	if params.Limit == 0 {
		params.Limit = 20
	}

	db := database.GetDB()
	query := db.Model(&models.User{}).
		Preload("LocalBody").
		Preload("District").
		Preload("State").
		Preload("HomeFacility")

	// Apply filters
	if params.Search != "" {
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR username ILIKE ? OR email ILIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	if userType := c.Query("user_type"); userType != "" {
		query = query.Where("user_type = ?", userType)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Apply pagination
	var users []models.User
	if err := query.Limit(params.Limit).Offset(params.Offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch users",
		})
		return
	}

	// Convert to bare minimum for list view
	var userList []models.UserBareMinimum
	for _, user := range users {
		userList = append(userList, models.UserBareMinimum{
			ID:                user.ID,
			Username:          user.Username,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			Email:             user.Email,
			UserType:          user.UserType,
			LastLogin:         user.LastLogin,
			ProfilePictureURL: user.ProfilePictureURL,
		})
	}

	response := models.PaginatedResponse{
		Count:   int(total),
		Results: userList,
	}

	c.JSON(http.StatusOK, response)
}

// GetCurrentUser godoc
// @Summary Get current user
// @Description Get current authenticated user details
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} models.ErrorResponse
// @Router /users/getcurrentuser/ [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid token",
		})
		return
	}

	var user models.User
	db := database.GetDB()
	if err := db.Preload("LocalBody").
		Preload("District").
		Preload("State").
		Preload("HomeFacility").
		Preload("Skills.Skill").
		Where("id = ?", claims.UserID).
		First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUser godoc
// @Summary Get user by username
// @Description Get user details by username
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Success 200 {object} models.User
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{username}/ [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	db := database.GetDB()
	if err := db.Preload("LocalBody").
		Preload("District").
		Preload("State").
		Preload("HomeFacility").
		Preload("Skills.Skill").
		Where("username = ?", username).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser godoc
// @Summary Create new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.User true "User data"
// @Success 201 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Router /users/ [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), h.config.Security.BcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to hash password",
		})
		return
	}
	user.Password = string(hashedPassword)

	// Create user
	db := database.GetDB()
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to create user",
		})
		return
	}

	// Load relationships
	db.Preload("LocalBody").
		Preload("District").
		Preload("State").
		Preload("HomeFacility").
		First(&user, user.ID)

	c.JSON(http.StatusCreated, user)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user details
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Param user body models.User true "User data"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Router /users/{username}/ [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	db := database.GetDB()
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	var updateData models.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Don't allow password updates through this endpoint
	updateData.Password = user.Password

	// Update user
	if err := db.Model(&user).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to update user",
		})
		return
	}

	// Load relationships
	db.Preload("LocalBody").
		Preload("District").
		Preload("State").
		Preload("HomeFacility").
		First(&user, user.ID)

	c.JSON(http.StatusOK, user)
}

// PartialUpdateUser godoc
// @Summary Partially update user
// @Description Partially update user details
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Param user body object true "Partial user data"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Router /users/{username}/ [patch]
func (h *UserHandler) PartialUpdateUser(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	db := database.GetDB()
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Remove password from update data
	delete(updateData, "password")

	// Update user
	if err := db.Model(&user).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to update user",
		})
		return
	}

	// Load relationships
	db.Preload("LocalBody").
		Preload("District").
		Preload("State").
		Preload("HomeFacility").
		First(&user, user.ID)

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Success 200 {object} models.SuccessResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{username}/ [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	username := c.Param("username")

	db := database.GetDB()
	result := db.Where("username = ?", username).Delete(&models.User{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete user",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "User deleted successfully",
	})
}

// CheckUsername godoc
// @Summary Check username availability
// @Description Check if username is available
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Success 200 {object} models.SuccessResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /users/{username}/check_availability/ [get]
func (h *UserHandler) CheckUsername(c *gin.Context) {
	username := c.Param("username")

	var count int64
	db := database.GetDB()
	db.Model(&models.User{}).Where("username = ?", username).Count(&count)

	if count > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   "conflict",
			Message: "Username is already taken",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Username is available",
	})
}

// ListUserSkills godoc
// @Summary List user skills
// @Description Get user's skills
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Success 200 {object} models.PaginatedResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{username}/skill/ [get]
func (h *UserHandler) ListUserSkills(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	db := database.GetDB()
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	var skills []models.UserSkill
	if err := db.Preload("Skill").Where("user_id = ?", user.ID).Find(&skills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch skills",
		})
		return
	}

	response := models.PaginatedResponse{
		Count:   len(skills),
		Results: skills,
	}

	c.JSON(http.StatusOK, response)
}

// AddUserSkill godoc
// @Summary Add user skill
// @Description Add skill to user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Param skill body object{skill=string} true "Skill ID"
// @Success 201 {object} models.UserSkill
// @Failure 400 {object} models.ErrorResponse
// @Router /users/{username}/skill/ [post]
func (h *UserHandler) AddUserSkill(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	db := database.GetDB()
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	var req struct {
		Skill string `json:"skill" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	userSkill := models.UserSkill{
		UserID:  user.ID,
		SkillID: req.Skill,
	}

	if err := db.Create(&userSkill).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to add skill",
		})
		return
	}

	// Load skill details
	db.Preload("Skill").First(&userSkill, "id = ?", userSkill.ID)

	c.JSON(http.StatusCreated, userSkill)
}

// DeleteUserSkill godoc
// @Summary Delete user skill
// @Description Remove skill from user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Param id path string true "Skill ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{username}/skill/{id}/ [delete]
func (h *UserHandler) DeleteUserSkill(c *gin.Context) {
	username := c.Param("username")
	skillID := c.Param("id")

	var user models.User
	db := database.GetDB()
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	result := db.Where("user_id = ? AND id = ?", user.ID, skillID).Delete(&models.UserSkill{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete skill",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Skill not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Skill removed successfully",
	})
}

// ListAllSkills godoc
// @Summary List all skills
// @Description Get all available skills
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.PaginatedResponse
// @Router /skill/ [get]
func (h *UserHandler) ListAllSkills(c *gin.Context) {
	var skills []models.Skill
	db := database.GetDB()
	if err := db.Find(&skills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch skills",
		})
		return
	}

	response := models.PaginatedResponse{
		Count:   len(skills),
		Results: skills,
	}

	c.JSON(http.StatusOK, response)
}

// Placeholder implementations for other user-related endpoints
func (h *UserHandler) ListUserFacilities(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *UserHandler) AddUserFacility(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *UserHandler) DeleteUserFacility(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *UserHandler) ClearHomeFacility(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *UserHandler) UpdateProfilePicture(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *UserHandler) DeleteProfilePicture(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *UserHandler) GetPNConfig(c *gin.Context) {
	// Return empty push notification config for now
	// This prevents 501 errors during facility creation
	c.JSON(http.StatusOK, gin.H{
		"pf_endpoint": "",
		"pf_p256dh": "",
		"pf_auth": "",
	})
}

func (h *UserHandler) UpdatePNConfig(c *gin.Context) {
	// Accept push notification config updates but don't store them yet
	// This prevents 501 errors during facility creation
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"pf_endpoint": "",
		"pf_p256dh": "",
		"pf_auth": "",
	})
}