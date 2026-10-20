package handlers

import (
	"net/http"

	"care-backend/internal/config"
	"care-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	config *config.Config
}

func NewNotificationHandler(cfg *config.Config) *NotificationHandler {
	return &NotificationHandler{
		config: cfg,
	}
}

// Placeholder implementations for all notification endpoints
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// Return empty notification list for now
	// This prevents 501 errors and ad blocker issues
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Count:    0,
		Next:     nil,
		Previous: nil,
		Results:  []interface{}{},
	})
}

func (h *NotificationHandler) GetNotificationData(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *NotificationHandler) GetPublicKey(c *gin.Context) {
	// Return empty public key for now
	// This prevents 501 errors during facility creation
	c.JSON(http.StatusOK, gin.H{
		"public_key": "",
	})
}

func (h *NotificationHandler) SendNotificationMessages(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}