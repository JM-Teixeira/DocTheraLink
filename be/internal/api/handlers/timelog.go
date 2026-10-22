package handlers

import (
	"net/http"
	"time"

	"klyniva/internal/config"
	"klyniva/internal/database"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TimeLogHandler struct {
	config *config.Config
}

func NewTimeLogHandler(cfg *config.Config) *TimeLogHandler {
	return &TimeLogHandler{
		config: cfg,
	}
}

// ClockIn godoc
// @Summary Clock in
// @Description Clock in for a shift
// @Tags timelog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.TimeLogRequest false "Clock in notes"
// @Success 201 {object} models.TimeLog
// @Router /timelog/clock-in [post]
func (h *TimeLogHandler) ClockIn(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req models.TimeLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Notes are optional, so we can proceed even if binding fails
		req.Notes = ""
	}

	db := database.GetDB()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Check if user already clocked in today without clocking out
	var existingLog models.TimeLog
	err := db.Where("user_id = ? AND date = ? AND clock_out IS NULL", userID, today).
		First(&existingLog).Error

	if err == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "already_clocked_in",
			Message: "You are already clocked in. Please clock out first.",
		})
		return
	}

	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Database error",
		})
		return
	}

	// Create new time log
	timeLog := models.TimeLog{
		UserID:  userID.(uint),
		ClockIn: now,
		Date:    today,
		Notes:   req.Notes,
	}

	if err := db.Create(&timeLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to clock in",
		})
		return
	}

	// Load user relationship
	db.Preload("User").First(&timeLog, timeLog.ID)

	c.JSON(http.StatusCreated, timeLog)
}

// ClockOut godoc
// @Summary Clock out
// @Description Clock out from a shift
// @Tags timelog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.TimeLogRequest false "Clock out notes"
// @Success 200 {object} models.TimeLog
// @Router /timelog/clock-out [post]
func (h *TimeLogHandler) ClockOut(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req models.TimeLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Notes = ""
	}

	db := database.GetDB()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Find active clock in for today
	var timeLog models.TimeLog
	err := db.Where("user_id = ? AND date = ? AND clock_out IS NULL", userID, today).
		First(&timeLog).Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "not_clocked_in",
			Message: "You are not clocked in. Please clock in first.",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Database error",
		})
		return
	}

	// Update clock out time
	timeLog.ClockOut = &now
	timeLog.Duration = timeLog.CalculateDuration()
	if req.Notes != "" {
		if timeLog.Notes != "" {
			timeLog.Notes += " | " + req.Notes
		} else {
			timeLog.Notes = req.Notes
		}
	}

	if err := db.Save(&timeLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to clock out",
		})
		return
	}

	// Load user relationship
	db.Preload("User").First(&timeLog, timeLog.ID)

	c.JSON(http.StatusOK, timeLog)
}

// GetMyLogs godoc
// @Summary Get my time logs
// @Description Get time logs for the authenticated user
// @Tags timelog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} models.PaginatedResponse
// @Router /timelog/my-logs [get]
func (h *TimeLogHandler) GetMyLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var params models.TimeLogQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid query parameters",
		})
		return
	}

	if params.Limit == 0 {
		params.Limit = 50
	}

	db := database.GetDB()
	query := db.Model(&models.TimeLog{}).Where("user_id = ?", userID)

	// Apply date filters
	if params.From != "" {
		fromDate, err := time.Parse("2006-01-02", params.From)
		if err == nil {
			query = query.Where("date >= ?", fromDate)
		}
	}

	if params.To != "" {
		toDate, err := time.Parse("2006-01-02", params.To)
		if err == nil {
			query = query.Where("date <= ?", toDate)
		}
	}

	var total int64
	query.Count(&total)

	var timeLogs []models.TimeLog
	if err := query.Order("date DESC, clock_in DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Preload("User").
		Find(&timeLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch time logs",
		})
		return
	}

	response := models.PaginatedResponse{
		Count:   int(total),
		Results: timeLogs,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserLogs godoc
// @Summary Get user time logs (Admin)
// @Description Get time logs for a specific user (admin only)
// @Tags timelog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} models.PaginatedResponse
// @Router /timelog/user/{id} [get]
func (h *TimeLogHandler) GetUserLogs(c *gin.Context) {
	userID := c.Param("id")

	// Own logs, or the logs of someone you manage
	a, ok := currentActor(c, database.GetDB())
	if !ok {
		return
	}
	var target models.User
	if err := database.GetDB().Select("id", "home_facility_id").First(&target, "id = ?", userID).Error; err != nil ||
		(target.ID != a.ID && !managesUser(a, &target)) {
		notFound(c, "User")
		return
	}

	var params models.TimeLogQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid query parameters",
		})
		return
	}

	if params.Limit == 0 {
		params.Limit = 50
	}

	db := database.GetDB()
	query := db.Model(&models.TimeLog{}).Where("user_id = ?", userID)

	// Apply date filters
	if params.From != "" {
		fromDate, err := time.Parse("2006-01-02", params.From)
		if err == nil {
			query = query.Where("date >= ?", fromDate)
		}
	}

	if params.To != "" {
		toDate, err := time.Parse("2006-01-02", params.To)
		if err == nil {
			query = query.Where("date <= ?", toDate)
		}
	}

	var total int64
	query.Count(&total)

	var timeLogs []models.TimeLog
	if err := query.Order("date DESC, clock_in DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Preload("User").
		Find(&timeLogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch time logs",
		})
		return
	}

	response := models.PaginatedResponse{
		Count:   int(total),
		Results: timeLogs,
	}

	c.JSON(http.StatusOK, response)
}

// GetCurrentStatus godoc
// @Summary Get current clock status
// @Description Check if user is currently clocked in
// @Tags timelog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{clocked_in=bool,time_log=models.TimeLog}
// @Router /timelog/status [get]
func (h *TimeLogHandler) GetCurrentStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	db := database.GetDB()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var timeLog models.TimeLog
	err := db.Where("user_id = ? AND date = ? AND clock_out IS NULL", userID, today).
		Preload("User").
		First(&timeLog).Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{
			"clocked_in": false,
			"time_log":   nil,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clocked_in": true,
		"time_log":   timeLog,
	})
}
