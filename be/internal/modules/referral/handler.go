package referral

import (
	"context"
	"errors"
	"log"
	"net/http"

	"care-backend/internal/auth"
	"care-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Handler exposes the Service over HTTP.
type Handler struct {
	svc *Service
}

// NewHandler builds a Handler around svc.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom builds the Actor from the JWT claims set by auth.AuthMiddleware.
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: verr.Msg})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: err.Error()})
	case errors.Is(err, ErrInvalidTransition), errors.Is(err, ErrConflict):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "conflict", Message: err.Error()})
	default:
		// Never leak internal error details to clients; log them instead.
		log.Printf("referral: internal error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
	}
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
}

// List godoc
// @Summary List referrals visible to the current user
// @Tags referrals
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param patient query string false "Filter by patient ID"
// @Param limit query int false "Page size (max 100)"
// @Param offset query int false "Offset"
// @Success 200 {object} models.PaginatedResponse
// @Router /referrals/ [get]
func (h *Handler) List(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	var f ListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		badRequest(c, err)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), actor, f)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, models.PaginatedResponse{Count: int(total), Results: items})
}

// Create godoc
// @Summary Refer a patient to physiotherapy
// @Tags referrals
// @Security BearerAuth
// @Param referral body CreateRequest true "Referral"
// @Success 201 {object} Referral
// @Router /referrals/ [post]
func (h *Handler) Create(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	r, err := h.svc.Create(c.Request.Context(), actor, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

// Get godoc
// @Summary Get a referral
// @Tags referrals
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Success 200 {object} Referral
// @Router /referrals/{id}/ [get]
func (h *Handler) Get(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	r, err := h.svc.Get(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

// Accept godoc
// @Summary Accept a pending referral (physiotherapist)
// @Tags referrals
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Success 200 {object} Referral
// @Router /referrals/{id}/accept/ [post]
func (h *Handler) Accept(c *gin.Context) {
	h.simpleAction(c, h.svc.Accept)
}

// Start godoc
// @Summary Start treatment (assigned physiotherapist)
// @Tags referrals
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Success 200 {object} Referral
// @Router /referrals/{id}/start/ [post]
func (h *Handler) Start(c *gin.Context) {
	h.simpleAction(c, h.svc.Start)
}

// Decline godoc
// @Summary Decline a referral (assigned physiotherapist)
// @Tags referrals
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Param body body ReasonRequest true "Reason"
// @Success 200 {object} Referral
// @Router /referrals/{id}/decline/ [post]
func (h *Handler) Decline(c *gin.Context) {
	h.reasonAction(c, h.svc.Decline)
}

// Complete godoc
// @Summary Complete treatment with an outcome summary (assigned physiotherapist)
// @Tags referrals
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Param body body ReasonRequest true "Outcome summary in reason"
// @Success 200 {object} Referral
// @Router /referrals/{id}/complete/ [post]
func (h *Handler) Complete(c *gin.Context) {
	h.reasonAction(c, h.svc.Complete)
}

// Cancel godoc
// @Summary Cancel a referral (referring doctor or admin)
// @Tags referrals
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Param body body ReasonRequest true "Reason"
// @Success 200 {object} Referral
// @Router /referrals/{id}/cancel/ [post]
func (h *Handler) Cancel(c *gin.Context) {
	h.reasonAction(c, h.svc.Cancel)
}

// ListNotes godoc
// @Summary List the doctor ↔ therapist thread of a referral
// @Tags referrals
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Success 200 {array} Note
// @Router /referrals/{id}/notes/ [get]
func (h *Handler) ListNotes(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	notes, err := h.svc.Notes(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, notes)
}

// AddNote godoc
// @Summary Post a note on a referral thread
// @Tags referrals
// @Security BearerAuth
// @Param id path string true "Referral ID"
// @Param note body NoteRequest true "Note"
// @Success 201 {object} Note
// @Router /referrals/{id}/notes/ [post]
func (h *Handler) AddNote(c *gin.Context) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	var req NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	n, err := h.svc.AddNote(c.Request.Context(), actor, c.Param("id"), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, n)
}

// simpleAction runs a status change that needs no request body.
func (h *Handler) simpleAction(c *gin.Context,
	fn func(ctx context.Context, a Actor, id string) (*Referral, error)) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	r, err := fn(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

// reasonAction runs a status change that requires a {"reason": "..."} body.
func (h *Handler) reasonAction(c *gin.Context,
	fn func(ctx context.Context, a Actor, id, reason string) (*Referral, error)) {
	actor, ok := actorFrom(c)
	if !ok {
		return
	}
	var req ReasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	r, err := fn(c.Request.Context(), actor, c.Param("id"), req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}
