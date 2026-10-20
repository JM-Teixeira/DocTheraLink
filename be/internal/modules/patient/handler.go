package patient

import (
	"errors"
	"log"
	"net/http"

	"care-backend/internal/auth"
	"care-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Handler exposes the Service over HTTP.
type Handler struct{ svc *Service }

// NewHandler builds a Handler around svc.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func actorFrom(c *gin.Context) (Actor, bool) {
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "Invalid token"})
		return Actor{}, false
	}
	return Actor{ID: claims.UserID, Role: claims.UserType}, true
}

// writeError maps service errors onto HTTP status codes.
func writeError(c *gin.Context, err error) {
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "validation_error", Message: verr.Error(),
			Details: map[string]string{verr.Field: verr.Msg},
		})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: err.Error()})
	default:
		log.Printf("patient: internal error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
	}
}

// List godoc
// @Summary List patients visible to the current user
// @Tags patients
// @Security BearerAuth
// @Param search query string false "Name contains"
// @Param phone_number query string false "Exact phone number"
// @Param facility query string false "Facility ID"
// @Success 200 {object} models.PaginatedResponse
// @Router /patient/ [get]
func (h *Handler) List(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), actor, q)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Count: int(total), Results: items})
}

// Create godoc
// @Summary Register a patient
// @Tags patients
// @Security BearerAuth
// @Param patient body CreateRequest true "Patient"
// @Success 201 {object} models.Patient
// @Router /patient/ [post]
func (h *Handler) Create(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	p, err := h.svc.Create(c.Request.Context(), actor, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

// Get godoc
// @Summary Get a patient
// @Tags patients
// @Security BearerAuth
// @Param id path string true "Patient ID"
// @Success 200 {object} models.Patient
// @Router /patient/{id}/ [get]
func (h *Handler) Get(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	p, err := h.svc.Get(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

// Update godoc
// @Summary Update a patient (PUT and PATCH both apply the fields sent)
// @Tags patients
// @Security BearerAuth
// @Param id path string true "Patient ID"
// @Success 200 {object} models.Patient
// @Router /patient/{id}/ [patch]
func (h *Handler) Update(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	p, err := h.svc.Update(c.Request.Context(), actor, c.Param("id"), body)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}
