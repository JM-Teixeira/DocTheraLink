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
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
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
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *NotificationHandler) SendNotificationMessages(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}