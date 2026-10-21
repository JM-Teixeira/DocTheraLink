package notification

import (
	"errors"
	"log"
	"net/http"

	"klyniva/internal/auth"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
)

// Handler exposes the Service over HTTP. Every endpoint acts on the signed-in
// user's own notifications only.
type Handler struct{ svc *Service }

// NewHandler builds a Handler around svc.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func userFrom(c *gin.Context) (uint, bool) {
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "Invalid token"})
		return 0, false
	}
	return claims.UserID, true
}

func writeError(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}
	log.Printf("notification: internal error: %v", err)
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
}

// List godoc
// @Summary The current user's notifications (newest first) and unread count
// @Tags notifications
// @Security BearerAuth
// @Param unread query bool false "Only unread"
// @Success 200 {object} Page
// @Router /notifications/ [get]
func (h *Handler) List(c *gin.Context) {
	uid, ok := userFrom(c)
	if !ok {
		return
	}
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	page, err := h.svc.List(c.Request.Context(), uid, q)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// UnreadCount godoc
// @Summary Number of unread notifications (for the bell badge)
// @Tags notifications
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Router /notifications/unread-count/ [get]
func (h *Handler) UnreadCount(c *gin.Context) {
	uid, ok := userFrom(c)
	if !ok {
		return
	}
	n, err := h.svc.UnreadCount(c.Request.Context(), uid)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread": n})
}

// MarkRead godoc
// @Summary Mark one notification as read
// @Tags notifications
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 204
// @Router /notifications/{id}/read/ [post]
func (h *Handler) MarkRead(c *gin.Context) {
	uid, ok := userFrom(c)
	if !ok {
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), uid, c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// MarkAllRead godoc
// @Summary Mark all notifications as read
// @Tags notifications
// @Security BearerAuth
// @Success 204
// @Router /notifications/read-all/ [post]
func (h *Handler) MarkAllRead(c *gin.Context) {
	uid, ok := userFrom(c)
	if !ok {
		return
	}
	if err := h.svc.MarkAllRead(c.Request.Context(), uid); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
