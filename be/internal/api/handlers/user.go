package handlers

import (
	"errors"
	"net/http"
	"strings"

	"klyniva/internal/auth"
	"klyniva/internal/config"
	"klyniva/internal/database"
	"klyniva/internal/models"

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
	params.Limit = pageLimit(params.Limit)

	db := database.GetDB()
	query := db.Model(&models.User{}).
		Preload("LocalBody").
		Preload("District").
		Preload("State").
		Preload("HomeFacility")

	// Apply filters
	if params.Search != "" {
		// Full name too, so "Inês Costa" finds Inês Costa
		like := "%" + params.Search + "%"
		query = query.Where("first_name || ' ' || last_name ILIKE ? OR username ILIKE ? OR email ILIKE ?", like, like, like)
	}

	if userType := c.Query("user_type"); userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if clinic := c.Query("home_facility"); clinic != "" {
		query = query.Where("home_facility_id = ?", clinic)
	}
	if status := c.Query("employment_status"); status != "" {
		query = query.Where("employment_status = ?", status)
	}
	// Staff see the people of their own clinic; admins everyone
	// (without a clinic: only themselves — never everyone)
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "User not authenticated"})
		return
	}
	if !auth.IsAdmin(claims.UserType) {
		// Their own clinic (currentActor also applies an admin's "view as")
		a, ok := currentActor(c, db)
		if !ok {
			return
		}
		if a.Clinic != nil {
			query = query.Where("home_facility_id = ?", *a.Clinic)
		} else {
			query = query.Where("id = ?", claims.UserID)
		}
	}
	query = query.Order("first_name, last_name")

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
	userList := []models.UserBareMinimum{} // [] not null when nobody matches
	for _, user := range users {
		clinicName := ""
		if user.HomeFacility != nil {
			clinicName = user.HomeFacility.Name
		}
		userList = append(userList, models.UserBareMinimum{
			ID:                user.ID,
			Username:          user.Username,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			Email:             user.Email,
			UserType:          user.UserType,
			LastLogin:         user.LastLogin,
			ProfilePictureURL: user.ProfilePictureURL,
			EmploymentStatus:  user.EmploymentStatus,
			EmploymentSince:   user.EmploymentSince,
			ClinicName:        clinicName,
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

	// Viewing as a role: the app shows that role and clinic, and says so
	if va, ok := auth.ViewingAs(c); ok {
		user.UserType = va.Role
		clinicID := va.Clinic
		user.HomeFacilityID = &clinicID
		user.HomeFacility = nil
		var f models.Facility
		if db.First(&f, "id = ?", clinicID).Error == nil {
			user.HomeFacility = &f
		}
		c.JSON(http.StatusOK, struct {
			models.User
			ViewAs auth.ViewAs `json:"view_as"`
		}{user, va})
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
	a, ok := currentActor(c, db)
	if !ok {
		return
	}
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
	// People of other clinics look like they don't exist
	if !canSeeUser(a, &user) {
		notFound(c, "User")
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
// Only admins (any clinic, any role) and coordinators (their own clinic,
// clinic roles) add people; the new account can never be a superuser.
func (h *UserHandler) CreateUser(c *gin.Context) {
	db := database.GetDB()
	a, ok := currentActor(c, db)
	if !ok {
		return
	}
	var req newUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}
	clinic, err := newUserClinic(a, &req)
	if err != nil {
		forbid(c)
		return
	}
	username, err := newUsername(db, clinic, req.FirstName, req.LastName)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	user := models.User{
		Username: username, Email: strings.TrimSpace(req.Email),
		FirstName: strings.TrimSpace(req.FirstName), LastName: strings.TrimSpace(req.LastName),
		UserType: req.UserType, HomeFacilityID: clinic, PhoneNumber: req.PhoneNumber,
		Qualification: req.Qualification, StateID: req.StateID, DistrictID: req.DistrictID,
		LocalBodyID: req.LocalBodyID, IsActive: true, EmploymentStatus: "active", Password: req.Password,
		MustChangePassword: true, // they choose their own at the first sign-in
	}
	if req.Gender != nil && *req.Gender != "" {
		g := models.GenderType(*req.Gender)
		user.Gender = &g
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
// PUT follows the same field rules as PATCH (only the fields sent change).
func (h *UserHandler) UpdateUser(c *gin.Context) {
	h.PartialUpdateUser(c)
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
// The person edits their own profile; their clinic's coordinator (or an
// admin) also their role. Password, status, clinic and flags have their
// own endpoints and are refused here.
func (h *UserHandler) PartialUpdateUser(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	db := database.GetDB()
	a, ok := currentActor(c, db)
	if !ok {
		return
	}
	if err := db.Where("username = ?", username).First(&user).Error; err != nil || !canSeeUser(a, &user) {
		notFound(c, "User")
		return
	}

	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}
	updateData, err := userUpdate(a, &user, body)
	if errors.Is(err, errBadField) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "Invalid email"})
		return
	}
	if err != nil {
		forbid(c)
		return
	}

	// Update user
	if len(updateData) == 0 {
		c.JSON(http.StatusOK, user)
		return
	}
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
// Admins only, never themselves. People who leave should be marked "left"
// instead (keeps their record and history).
func (h *UserHandler) DeleteUser(c *gin.Context) {
	username := c.Param("username")

	db := database.GetDB()
	a, ok := currentActor(c, db)
	if !ok {
		return
	}
	if !auth.IsAdmin(a.Role) {
		forbid(c)
		return
	}
	result := db.Where("username = ? AND id <> ?", username, a.ID).Delete(&models.User{})
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
		"pf_p256dh":   "",
		"pf_auth":     "",
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
		"pf_p256dh":   "",
		"pf_auth":     "",
	})
}
