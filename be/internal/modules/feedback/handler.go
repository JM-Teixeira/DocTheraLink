package feedback

import (
	"errors"
	"log"
	"net/http"

	"klyniva/internal/auth"
	"klyniva/internal/clinic"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires /suggestions onto an authenticated group.
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	h := &handler{svc: NewService(NewGormRepository(db), clinic.NewGormResolver(db))}
	g := protected.Group("/suggestions")
	g.POST("/", h.send)
	g.GET("/mine/", h.mine)
	g.GET("/", h.all)
	g.GET("/count/", h.count)
	g.PATCH("/:id/", h.review)
}

type handler struct{ svc *Service }

func actorFrom(c *gin.Context) (Actor, bool) {
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "Invalid token"})
		return Actor{}, false
	}
	return Actor{ID: claims.UserID, Role: claims.UserType}, true
}

func writeError(c *gin.Context, err error) {
	var v *ValidationError
	switch {
	case errors.As(err, &v):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: v.Error(),
			Details: map[string]string{v.Field: v.Msg}})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: err.Error()})
	default:
		log.Printf("feedback: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
	}
}

// send godoc
// @Summary Send a suggestion about the app
// @Tags suggestions
// @Security BearerAuth
// @Router /suggestions/ [post]
func (h *handler) send(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	s, err := h.svc.Send(c.Request.Context(), a, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, s)
}

func (h *handler) mine(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	items, err := h.svc.Mine(c.Request.Context(), a)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *handler) all(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	items, err := h.svc.All(c.Request.Context(), a, c.Query("status"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *handler) count(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	n, err := h.svc.CountNew(c.Request.Context(), a)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"new": n})
}

func (h *handler) review(c *gin.Context) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}
	s, err := h.svc.Review(c.Request.Context(), a, c.Param("id"), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, s)
}
