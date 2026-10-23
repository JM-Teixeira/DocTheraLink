package patient

import (
	"context"
	"net/http"
	"strings"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Consent kinds a clinic records (GDPR). Labels live in the frontend.
var consentKinds = map[string]bool{
	"data_processing":   true, // health data for the treatment (required to treat)
	"share_with_doctor": true, // progress shared with the referring doctor
	"contact_email":     true,
	"contact_sms":       true,
	"photos":            true, // photos / video for the clinical record
}

// How the consent was given.
var consentMethods = map[string]bool{"signed_paper": true, "verbal": true, "digital": true}

// Consent is one record: given or withdrawn. Records are never edited; the
// newest of each kind is the current state, the rest is the history.
type Consent struct {
	ID           string        `json:"id" gorm:"primaryKey;type:uuid"`
	PatientID    string        `json:"patient_id" gorm:"type:uuid;not null;index"`
	Kind         string        `json:"kind" gorm:"size:30;not null"`
	Granted      bool          `json:"granted"`
	Method       string        `json:"method" gorm:"size:20;not null"`
	Note         string        `json:"note" gorm:"size:300"`
	RecordedByID uint          `json:"recorded_by_id" gorm:"not null"`
	RecordedBy   *consentActor `json:"recorded_by,omitempty" gorm:"foreignKey:RecordedByID;->;-:migration"`
	CreatedAt    time.Time     `json:"created_date"`
}

func (Consent) TableName() string { return "patient_consents" }

// BeforeCreate assigns a UUID.
func (c *Consent) BeforeCreate(*gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

// consentActor is who recorded a consent (read-only view of users).
type consentActor struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (consentActor) TableName() string { return "users" }

// Models lists the tables owned by this module besides models.Patient.
func Models() []any { return []any{&Consent{}} }

// ConsentRequest records a consent being given or withdrawn.
type ConsentRequest struct {
	Kind    string `json:"kind" binding:"required"`
	Granted *bool  `json:"granted" binding:"required"`
	Method  string `json:"method" binding:"required"`
	Note    string `json:"note" binding:"max=300"`
}

// ConsentView is the current state per kind plus the full history (newest first).
type ConsentView struct {
	Current map[string]Consent `json:"current"`
	History []Consent          `json:"history"`
}

// Consents returns a visible patient's consents.
func (s *Service) Consents(ctx context.Context, a Actor, patientID string) (*ConsentView, error) {
	if _, err := s.Get(ctx, a, patientID); err != nil {
		return nil, err
	}
	history, err := s.repo.Consents(ctx, patientID)
	if err != nil {
		return nil, err
	}
	return consentView(history), nil
}

// consentView takes the newest record of each kind as current (history is newest first).
func consentView(history []Consent) *ConsentView {
	v := &ConsentView{Current: map[string]Consent{}, History: history}
	if v.History == nil {
		v.History = []Consent{}
	}
	for _, c := range history {
		if _, seen := v.Current[c.Kind]; !seen {
			v.Current[c.Kind] = c
		}
	}
	return v
}

// RecordConsent adds a record. Whoever may edit the patient may record it.
func (s *Service) RecordConsent(ctx context.Context, a Actor, patientID string, req ConsentRequest) (*ConsentView, error) {
	if !auth.CanModifyResource(a.Role) {
		return nil, ErrForbidden
	}
	if _, err := s.Get(ctx, a, patientID); err != nil {
		return nil, err
	}
	if !consentKinds[req.Kind] {
		return nil, &ValidationError{"kind", "is not a known consent"}
	}
	if !consentMethods[req.Method] {
		return nil, &ValidationError{"method", "must be signed_paper, verbal or digital"}
	}
	c := &Consent{
		PatientID: patientID, Kind: req.Kind, Granted: *req.Granted, Method: req.Method,
		Note: strings.TrimSpace(req.Note), RecordedByID: a.ID,
	}
	if err := s.repo.AddConsent(ctx, c); err != nil {
		return nil, err
	}
	return s.Consents(ctx, a, patientID)
}

// ListConsents godoc
// @Summary A patient's consents: current state per kind and history
// @Tags patients
// @Security BearerAuth
// @Param id path string true "Patient ID"
// @Success 200 {object} ConsentView
// @Router /patient/{id}/consents/ [get]
func (h *Handler) ListConsents(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	v, err := h.svc.Consents(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// RecordConsent godoc
// @Summary Record a consent being given or withdrawn
// @Tags patients
// @Security BearerAuth
// @Param id path string true "Patient ID"
// @Param consent body ConsentRequest true "Consent"
// @Success 201 {object} ConsentView
// @Router /patient/{id}/consents/ [post]
func (h *Handler) RecordConsent(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	var req ConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	v, err := h.svc.RecordConsent(c.Request.Context(), actor, c.Param("id"), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, v)
}
