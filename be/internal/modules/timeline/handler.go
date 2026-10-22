package timeline

import (
	"errors"
	"log"
	"net/http"

	"klyniva/internal/auth"
	"klyniva/internal/models"
	"klyniva/internal/modules/patient"

	"github.com/gin-gonic/gin"
)

// Handler exposes the Service over HTTP.
type Handler struct{ svc *Service }

// NewHandler builds a Handler around svc.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Get godoc
// @Summary Everything that happened to a patient (referrals, messages, files, sessions)
// @Tags patients
// @Security BearerAuth
// @Param id path string true "Patient ID"
// @Success 200 {object} Timeline
// @Router /patient/{id}/timeline/ [get]
func (h *Handler) Get(c *gin.Context) {
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "Invalid token"})
		return
	}
	tl, err := h.svc.For(c.Request.Context(), Actor{ID: claims.UserID, Role: claims.UserType}, c.Param("id"))
	switch {
	case err == nil:
		c.JSON(http.StatusOK, tl)
	case errors.Is(err, patient.ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, patient.ErrForbidden):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: err.Error()})
	default:
		log.Printf("timeline: internal error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
	}
}
