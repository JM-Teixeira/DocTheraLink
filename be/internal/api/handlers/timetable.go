package handlers

import (
	"net/http"
	"time"

	"care-backend/internal/config"
	"care-backend/internal/database"
	"care-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TimetableHandler struct {
	config *config.Config
}

func NewTimetableHandler(cfg *config.Config) *TimetableHandler {
	return &TimetableHandler{
		config: cfg,
	}
}

// ListSections godoc
// @Summary List facility sections
// @Description Get all sections for a facility
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Facility ID"
// @Success 200 {array} models.FacilitySection
// @Router /facility/{id}/sections [get]
func (h *TimetableHandler) ListSections(c *gin.Context) {
	facilityID := c.Param("id")

	db := database.GetDB()
	var sections []models.TimetableSection
	if err := db.Where("facility_id = ? AND is_active = ?", facilityID, true).
		Order("name").
		Find(&sections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch sections",
		})
		return
	}

	c.JSON(http.StatusOK, sections)
}

// CreateSection godoc
// @Summary Create facility section
// @Description Create a new section for a facility
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Facility ID"
// @Param section body models.FacilitySectionRequest true "Section data"
// @Success 201 {object} models.FacilitySection
// @Router /facility/{id}/sections [post]
func (h *TimetableHandler) CreateSection(c *gin.Context) {
	facilityID := c.Param("id")

	var req models.TimetableSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	section := models.TimetableSection{
		FacilityID:  facilityID,
		Name:        req.Name,
		MaxCapacity: req.MaxCapacity,
		IsActive:    req.IsActive,
		Description: req.Description,
	}

	if section.MaxCapacity == 0 {
		section.MaxCapacity = 3
	}

	db := database.GetDB()
	if err := db.Create(&section).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to create section",
		})
		return
	}

	c.JSON(http.StatusCreated, section)
}

// GetSection godoc
// @Summary Get section
// @Description Get section by ID
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Section ID"
// @Success 200 {object} models.FacilitySection
// @Router /sections/{id} [get]
func (h *TimetableHandler) GetSection(c *gin.Context) {
	id := c.Param("id")

	var section models.TimetableSection
	db := database.GetDB()
	if err := db.Preload("Facility").
		Where("id = ?", id).
		First(&section).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Section not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, section)
}

// UpdateSection godoc
// @Summary Update section
// @Description Update section details
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Section ID"
// @Param section body models.FacilitySectionRequest true "Section data"
// @Success 200 {object} models.FacilitySection
// @Router /sections/{id} [put]
func (h *TimetableHandler) UpdateSection(c *gin.Context) {
	id := c.Param("id")

	var section models.TimetableSection
	db := database.GetDB()
	if err := db.Where("id = ?", id).First(&section).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Section not found",
		})
		return
	}

	var req models.TimetableSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	section.Name = req.Name
	section.MaxCapacity = req.MaxCapacity
	section.IsActive = req.IsActive
	section.Description = req.Description

	if err := db.Save(&section).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to update section",
		})
		return
	}

	c.JSON(http.StatusOK, section)
}

// DeleteSection godoc
// @Summary Delete section
// @Description Delete section (soft delete - sets is_active to false)
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Section ID"
// @Success 200 {object} models.SuccessResponse
// @Router /sections/{id} [delete]
func (h *TimetableHandler) DeleteSection(c *gin.Context) {
	id := c.Param("id")

	db := database.GetDB()
	result := db.Model(&models.TimetableSection{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete section",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Section not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Section deleted successfully",
	})
}

// GetTimetable godoc
// @Summary Get timetable
// @Description Get timetable for a facility
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Facility ID"
// @Param day_of_week query int false "Day of week (0-6)"
// @Param section_id query string false "Section ID"
// @Param user_id query int false "User ID"
// @Param date query string false "Date (YYYY-MM-DD)"
// @Success 200 {array} models.TimetableSlot
// @Router /timetable/{id} [get]
func (h *TimetableHandler) GetTimetable(c *gin.Context) {
	facilityID := c.Param("id")

	var params models.TimetableQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid query parameters",
		})
		return
	}

	db := database.GetDB()
	query := db.Model(&models.TimetableSlot{}).Where("facility_id = ?", facilityID)

	// Apply filters
	if params.DayOfWeek != nil {
		query = query.Where("day_of_week = ?", *params.DayOfWeek)
	}

	if params.SectionID != "" {
		query = query.Where("section_id = ?", params.SectionID)
	}

	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}

	if params.Date != "" {
		date, err := time.Parse("2006-01-02", params.Date)
		if err == nil {
			query = query.Where("(valid_from IS NULL OR valid_from <= ?) AND (valid_until IS NULL OR valid_until >= ?)", date, date)
		}
	}

	var slots []models.TimetableSlot
	if err := query.Preload("Section").
		Preload("User").
		Preload("Facility").
		Order("day_of_week, start_time").
		Find(&slots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch timetable",
		})
		return
	}

	c.JSON(http.StatusOK, slots)
}

// CreateSlot godoc
// @Summary Create timetable slot
// @Description Create a new timetable slot
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Facility ID"
// @Param slot body models.TimetableSlotRequest true "Slot data"
// @Success 201 {object} models.TimetableSlot
// @Router /timetable/{id}/slot [post]
func (h *TimetableHandler) CreateSlot(c *gin.Context) {
	facilityID := c.Param("id")

	var req models.TimetableSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Check for capacity conflicts
	db := database.GetDB()
	if err := h.checkCapacity(db, facilityID, req.SectionID, req.DayOfWeek, req.StartTime, req.EndTime, ""); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "capacity_exceeded",
			Message: err.Error(),
		})
		return
	}

	slot := models.TimetableSlot{
		FacilityID:  facilityID,
		SectionID:   req.SectionID,
		UserID:      req.UserID,
		DayOfWeek:   req.DayOfWeek,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsRecurring: req.IsRecurring,
		Notes:       req.Notes,
	}

	if req.ValidFrom != nil {
		slot.ValidFrom = *req.ValidFrom
	} else {
		slot.ValidFrom = time.Now()
	}

	if req.ValidUntil != nil {
		slot.ValidUntil = req.ValidUntil
	}

	if err := db.Create(&slot).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to create slot",
		})
		return
	}

	// Load relationships
	db.Preload("Section").
		Preload("User").
		Preload("Facility").
		First(&slot, "id = ?", slot.ID)

	c.JSON(http.StatusCreated, slot)
}

// UpdateSlot godoc
// @Summary Update timetable slot
// @Description Update timetable slot details
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Slot ID"
// @Param slot body models.TimetableSlotRequest true "Slot data"
// @Success 200 {object} models.TimetableSlot
// @Router /timetable/slot/{id} [put]
func (h *TimetableHandler) UpdateSlot(c *gin.Context) {
	id := c.Param("id")

	var slot models.TimetableSlot
	db := database.GetDB()
	if err := db.Where("id = ?", id).First(&slot).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Slot not found",
		})
		return
	}

	var req models.TimetableSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Check for capacity conflicts (excluding current slot)
	if err := h.checkCapacity(db, slot.FacilityID, req.SectionID, req.DayOfWeek, req.StartTime, req.EndTime, id); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "capacity_exceeded",
			Message: err.Error(),
		})
		return
	}

	slot.SectionID = req.SectionID
	slot.UserID = req.UserID
	slot.DayOfWeek = req.DayOfWeek
	slot.StartTime = req.StartTime
	slot.EndTime = req.EndTime
	slot.IsRecurring = req.IsRecurring
	slot.Notes = req.Notes

	if req.ValidFrom != nil {
		slot.ValidFrom = *req.ValidFrom
	}

	slot.ValidUntil = req.ValidUntil

	if err := db.Save(&slot).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to update slot",
		})
		return
	}

	// Load relationships
	db.Preload("Section").
		Preload("User").
		Preload("Facility").
		First(&slot, "id = ?", slot.ID)

	c.JSON(http.StatusOK, slot)
}

// DeleteSlot godoc
// @Summary Delete timetable slot
// @Description Delete timetable slot
// @Tags timetable
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Slot ID"
// @Success 200 {object} models.SuccessResponse
// @Router /timetable/slot/{id} [delete]
func (h *TimetableHandler) DeleteSlot(c *gin.Context) {
	id := c.Param("id")

	db := database.GetDB()
	result := db.Where("id = ?", id).Delete(&models.TimetableSlot{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete slot",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Slot not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Slot deleted successfully",
	})
}

// checkCapacity checks if adding/updating a slot would exceed section capacity
func (h *TimetableHandler) checkCapacity(db *gorm.DB, facilityID, sectionID string, dayOfWeek int, startTime, endTime, excludeSlotID string) error {
	// Get section max capacity
	var section models.TimetableSection
	if err := db.Where("id = ?", sectionID).First(&section).Error; err != nil {
		return err
	}

	// Count overlapping slots
	query := db.Model(&models.TimetableSlot{}).
		Where("facility_id = ? AND section_id = ? AND day_of_week = ?", facilityID, sectionID, dayOfWeek).
		Where("(start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?)",
			endTime, startTime, endTime, endTime, startTime, endTime)

	if excludeSlotID != "" {
		query = query.Where("id != ?", excludeSlotID)
	}

	var count int64
	query.Count(&count)

	if int(count) >= section.MaxCapacity {
		return gorm.ErrInvalidData
	}

	return nil
}
