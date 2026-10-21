package inventory

import (
	"errors"
	"log"
	"net/http"

	"klyniva/internal/auth"
	"klyniva/internal/models"

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

func writeError(c *gin.Context, err error) {
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "validation_error", Message: verr.Error(), Details: map[string]string{verr.Field: verr.Msg},
		})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrNoClinic):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: err.Error()})
	case errors.Is(err, ErrInsufficient):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "insufficient_stock", Message: err.Error()})
	default:
		log.Printf("inventory: internal error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
	}
}

// bind decodes a JSON body or answers 400.
func bind(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return false
	}
	return true
}

// List handles GET /inventory/items/?search=&category=&status=
func (h *Handler) List(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var f ListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	items, err := h.svc.List(c.Request.Context(), a, f)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// Summary handles GET /inventory/summary/
func (h *Handler) Summary(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	sum, err := h.svc.Summary(c.Request.Context(), a)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, sum)
}

// Create handles POST /inventory/items/
func (h *Handler) Create(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req ItemRequest
	if !bind(c, &req) {
		return
	}
	it, err := h.svc.Create(c.Request.Context(), a, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, it)
}

// Update handles PUT /inventory/items/:id/
func (h *Handler) Update(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req ItemRequest
	if !bind(c, &req) {
		return
	}
	it, err := h.svc.Update(c.Request.Context(), a, c.Param("id"), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

// Archive handles DELETE /inventory/items/:id/
func (h *Handler) Archive(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	if err := h.svc.Archive(c.Request.Context(), a, c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Move handles POST /inventory/items/:id/movements/
func (h *Handler) Move(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req MovementRequest
	if !bind(c, &req) {
		return
	}
	it, err := h.svc.Move(c.Request.Context(), a, c.Param("id"), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

// Audit handles GET /inventory/audit/?item=&user=&kind=&from=&to=&auto=
func (h *Handler) Audit(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var f AuditFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	ms, err := h.svc.Audit(c.Request.Context(), a, f)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, ms)
}

// History handles GET /inventory/items/:id/movements/
func (h *Handler) History(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	ms, err := h.svc.History(c.Request.Context(), a, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, ms)
}
