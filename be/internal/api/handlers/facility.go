package handlers

import (
	"net/http"
	"strings"

	"klyniva/internal/config"
	"klyniva/internal/database"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FacilityHandler struct {
	config *config.Config
}

func NewFacilityHandler(cfg *config.Config) *FacilityHandler {
	return &FacilityHandler{
		config: cfg,
	}
}

// ListFacilities godoc
// @Summary List facilities
// @Description Get paginated list of facilities
// @Tags facilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of results per page"
// @Param offset query int false "Number of results to skip"
// @Param search query string false "Search term"
// @Success 200 {object} models.PaginatedResponse
// @Router /facility/ [get]
func (h *FacilityHandler) ListFacilities(c *gin.Context) {
	var params models.QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid query parameters",
		})
		return
	}

	if params.Limit == 0 {
		params.Limit = 20
	}

	db := database.GetDB()
	query := db.Model(&models.Facility{}).
		Preload("State").
		Preload("District")

	if params.Search != "" {
		query = query.Where("name ILIKE ?", "%"+params.Search+"%")
	}

	var total int64
	query.Count(&total)

	var facilities []models.Facility
	if err := query.Limit(params.Limit).Offset(params.Offset).Find(&facilities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch facilities",
		})
		return
	}

	response := models.PaginatedResponse{
		Count:   int(total),
		Results: facilities,
	}

	c.JSON(http.StatusOK, response)
}

// CreateFacility godoc
// @Summary Create facility
// @Description Create a new facility
// @Tags facilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param facility body models.FacilityRequest true "Facility data"
// @Success 201 {object} models.Facility
// @Router /facility/ [post]
func (h *FacilityHandler) CreateFacility(c *gin.Context) {
	var req models.FacilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log the actual error for debugging
		println("ERROR: Failed to bind JSON:", err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Log what we received
	println("DEBUG: Received facility request:")
	println("  Name:", req.Name)
	println("  FacilityType:", req.FacilityType)
	println("  Address:", req.Address)

	facility := models.Facility{
		Name:              req.Name,
		FacilityType:      req.FacilityType,
		Address:           req.Address,
		MiddlewareAddress: req.MiddlewareAddress,
		Pincode:           req.Pincode,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		StateID:           req.StateID,
		DistrictID:        req.DistrictID,
		KaspEmpanelled:    req.KaspEmpanelled,
		Features:          req.Features,
	}

	db := database.GetDB()
	if err := db.Create(&facility).Error; err != nil {
		// Log the actual database error
		println("ERROR: Failed to create facility in database:", err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to create facility: " + err.Error(),
		})
		return
	}

	println("SUCCESS: Facility created with ID:", facility.ID)

	// Load relationships
	db.Preload("State").
		Preload("District").
		First(&facility, "id = ?", facility.ID)

	c.JSON(http.StatusCreated, facility)
}

// GetFacility godoc
// @Summary Get facility
// @Description Get facility by ID
// @Tags facilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Facility ID"
// @Success 200 {object} models.Facility
// @Router /facility/{id}/ [get]
func (h *FacilityHandler) GetFacility(c *gin.Context) {
	id := c.Param("id")

	var facility models.Facility
	db := database.GetDB()
	if err := db.Preload("State").
		Preload("District").
		Where("id = ?", id).
		First(&facility).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Facility not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, facility)
}

// UpdateFacility godoc
// @Summary Update facility
// @Description Update facility details
// @Tags facilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Facility ID"
// @Param facility body models.FacilityRequest true "Facility data"
// @Success 200 {object} models.Facility
// @Router /facility/{id}/ [put]
func (h *FacilityHandler) UpdateFacility(c *gin.Context) {
	id := c.Param("id")

	var facility models.Facility
	db := database.GetDB()
	if err := db.Where("id = ?", id).First(&facility).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Facility not found",
		})
		return
	}

	var req models.FacilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
		return
	}

	// Update facility
	facility.Name = req.Name
	facility.FacilityType = req.FacilityType
	facility.Address = req.Address
	facility.MiddlewareAddress = req.MiddlewareAddress
	facility.Pincode = req.Pincode
	facility.Latitude = req.Latitude
	facility.Longitude = req.Longitude
	facility.StateID = req.StateID
	facility.DistrictID = req.DistrictID
	facility.KaspEmpanelled = req.KaspEmpanelled
	facility.Features = req.Features

	if err := db.Save(&facility).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to update facility",
		})
		return
	}

	// Load relationships
	db.Preload("State").
		Preload("District").
		First(&facility, "id = ?", facility.ID)

	c.JSON(http.StatusOK, facility)
}

// PartialUpdateFacility godoc
// @Summary Partially update facility
// @Description Partially update facility details
// @Tags facilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Facility ID"
// @Param facility body object true "Partial facility data"
// @Success 200 {object} models.Facility
// @Router /facility/{id}/ [patch]
func (h *FacilityHandler) PartialUpdateFacility(c *gin.Context) {
	id := c.Param("id")

	var facility models.Facility
	db := database.GetDB()
	if err := db.Where("id = ?", id).First(&facility).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Facility not found",
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

	if err := db.Model(&facility).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "bad_request",
			Message: "Failed to update facility",
		})
		return
	}

	// Load relationships
	db.Preload("State").
		Preload("District").
		First(&facility, "id = ?", facility.ID)

	c.JSON(http.StatusOK, facility)
}

// DeleteFacility godoc
// @Summary Delete facility
// @Description Delete facility
// @Tags facilities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Facility ID"
// @Success 200 {object} models.SuccessResponse
// @Router /facility/{id}/ [delete]
func (h *FacilityHandler) DeleteFacility(c *gin.Context) {
	id := c.Param("id")

	db := database.GetDB()
	result := db.Where("id = ?", id).Delete(&models.Facility{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete facility",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Facility not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Facility deleted successfully",
	})
}

// ListAllFacilities - Public endpoint for all facilities
func (h *FacilityHandler) ListAllFacilities(c *gin.Context) {
	h.ListFacilities(c)
}

// GetAnyFacility - Public endpoint for any facility
func (h *FacilityHandler) GetAnyFacility(c *gin.Context) {
	h.GetFacility(c)
}

// ListStates godoc
// @Summary List states
// @Description Get all states
// @Tags geographic
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.PaginatedResponse
// @Router /state/ [get]
func (h *FacilityHandler) ListStates(c *gin.Context) {
	var states []models.State
	db := database.GetDB()
	if err := db.Find(&states).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch states",
		})
		return
	}

	response := models.PaginatedResponse{
		Count:   len(states),
		Results: states,
	}

	c.JSON(http.StatusOK, response)
}

// GetState godoc
// @Summary Get state
// @Description Get state by ID
// @Tags geographic
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "State ID"
// @Success 200 {object} models.State
// @Router /state/{id}/ [get]
func (h *FacilityHandler) GetState(c *gin.Context) {
	id := c.Param("id")

	var state models.State
	db := database.GetDB()
	if err := db.Where("id = ?", id).First(&state).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "State not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, state)
}

// Placeholder implementations for other endpoints
// GetFacilityUsers lists the active users working in a clinic (their home
// facility), optionally filtered by ?username=, with limit/offset paging.
func (h *FacilityHandler) GetFacilityUsers(c *gin.Context) {
	var params models.QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: "Invalid query parameters"})
		return
	}
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 20
	}
	query := database.GetDB().Model(&models.User{}).Where("home_facility_id = ? AND is_active", c.Param("id"))
	if u := strings.TrimSpace(c.Query("username")); u != "" {
		query = query.Where("username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", "%"+u+"%", "%"+u+"%", "%"+u+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Failed to fetch users"})
		return
	}
	var users []models.User
	if err := query.Preload("HomeFacility").Order("first_name, last_name").
		Limit(params.Limit).Offset(params.Offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Count: int(total), Results: users})
}

func (h *FacilityHandler) GetDistrictsByState(c *gin.Context) {
	stateID := c.Param("id")

	var districts []models.District
	db := database.GetDB()

	if err := db.Where("state_id = ?", stateID).Preload("State").Find(&districts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch districts",
		})
		return
	}

	c.JSON(http.StatusOK, districts)
}

func (h *FacilityHandler) GetDistrictByName(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: "not_implemented", Message: "Endpoint not implemented yet"})
}

func (h *FacilityHandler) GetDistrict(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: "not_implemented", Message: "Endpoint not implemented yet"})
}

func (h *FacilityHandler) GetAllLocalBodyByDistrict(c *gin.Context) {
	districtID := c.Param("id")

	var localBodies []models.LocalBody
	db := database.GetDB()

	if err := db.Where("district_id = ?", districtID).Preload("District").Find(&localBodies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch cities",
		})
		return
	}

	c.JSON(http.StatusOK, localBodies)
}

func (h *FacilityHandler) GetLocalBodyByDistrict(c *gin.Context) {
	districtID := c.Param("id")

	var localBodies []models.LocalBody
	db := database.GetDB()

	if err := db.Where("district_id = ?", districtID).Preload("District").Find(&localBodies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch cities",
		})
		return
	}

	c.JSON(http.StatusOK, localBodies)
}

func (h *FacilityHandler) GetAllLocalBody(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: "not_implemented", Message: "Endpoint not implemented yet"})
}

func (h *FacilityHandler) GetLocalBody(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: "not_implemented", Message: "Endpoint not implemented yet"})
}

func (h *FacilityHandler) GetWards(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: "not_implemented", Message: "Endpoint not implemented yet"})
}

func (h *FacilityHandler) GetWard(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: "not_implemented", Message: "Endpoint not implemented yet"})
}

func (h *FacilityHandler) DownloadFacility(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: "not_implemented", Message: "Endpoint not implemented yet"})
}
