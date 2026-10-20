package handlers

import (
	"net/http"

	"klyniva/internal/config"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
)

// PatientHandler holds the patient endpoints not yet implemented.
// List/create/search/get/update are served by internal/modules/patient.
type PatientHandler struct {
	config *config.Config
}

func NewPatientHandler(cfg *config.Config) *PatientHandler {
	return &PatientHandler{
		config: cfg,
	}
}

func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Endpoint not implemented yet",
	})
}

func (h *PatientHandler) DownloadPatients(c *gin.Context) { notImplemented(c) }
