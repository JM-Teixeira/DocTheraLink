package staff

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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: verr.Error(),
			Details: map[string]string{verr.Field: verr.Msg}})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: err.Error()})
	default:
		log.Printf("staff: internal error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
	}
}

// page wraps a list in the {count, results} shape the user pages expect.
func page[T any](items []T) models.PaginatedResponse {
	if items == nil {
		items = []T{}
	}
	return models.PaginatedResponse{Count: len(items), Results: items}
}

// Skills handles GET /skill/ (the catalogue).
func (h *Handler) Skills(c *gin.Context) {
	skills, err := h.svc.Skills(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, page(skills))
}

// UserSkills handles GET /users/:username/skill/
func (h *Handler) UserSkills(c *gin.Context) {
	skills, err := h.svc.UserSkills(c.Request.Context(), c.Param("username"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, page(skills))
}

// AddSkill handles POST /users/:username/skill/ {skill}
func (h *Handler) AddSkill(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req struct {
		Skill string `json:"skill" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	us, err := h.svc.AddSkill(c.Request.Context(), a, c.Param("username"), req.Skill)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, us)
}

// RemoveSkill handles DELETE /users/:username/skill/:id/
func (h *Handler) RemoveSkill(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	if err := h.svc.RemoveSkill(c.Request.Context(), a, c.Param("username"), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Clinics handles GET /users/:username/get_facilities/
func (h *Handler) Clinics(c *gin.Context) {
	clinics, err := h.svc.Clinics(c.Request.Context(), c.Param("username"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, page(clinics))
}

// SetClinic handles PUT /users/:username/add_facility/ {facility}
func (h *Handler) SetClinic(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req struct {
		Facility string `json:"facility" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	u, err := h.svc.SetClinic(c.Request.Context(), a, c.Param("username"), req.Facility)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

// ClearClinic handles DELETE /users/:username/delete_facility/ and
// /users/:username/clear_home_facility/
func (h *Handler) ClearClinic(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	if _, err := h.svc.SetClinic(c.Request.Context(), a, c.Param("username"), ""); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetEmployment handles PATCH /users/:username/employment/
func (h *Handler) SetEmployment(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req EmploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	u, err := h.svc.SetEmployment(c.Request.Context(), a, c.Param("username"), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

// RegisterRoutes wires the staff API onto the authenticated group (the
// paths of the existing user pages).
func RegisterRoutes(protected *gin.RouterGroup, svc *Service) {
	mount(protected, NewHandler(svc))
}

// mount is the route table, shared with the handler tests.
func mount(g *gin.RouterGroup, h *Handler) {
	g.GET("/skill/", h.Skills)
	users := g.Group("/users/:username")
	users.GET("/skill/", h.UserSkills)
	users.POST("/skill/", h.AddSkill)
	users.DELETE("/skill/:id/", h.RemoveSkill)
	users.GET("/get_facilities/", h.Clinics)
	users.PUT("/add_facility/", h.SetClinic)
	users.DELETE("/delete_facility/", h.ClearClinic)
	users.DELETE("/clear_home_facility/", h.ClearClinic)
	users.PATCH("/employment/", h.SetEmployment)
}
