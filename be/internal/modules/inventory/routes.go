package inventory

import (
	"klyniva/internal/clinic"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires the inventory API onto an authenticated group.
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	svc := NewService(NewGormRepository(db), clinic.NewGormResolver(db))
	mount(protected.Group("/inventory"), NewHandler(svc))
}

// mount is the route table, shared with the handler tests.
func mount(g *gin.RouterGroup, h *Handler) {
	g.GET("/summary/", h.Summary)
	g.GET("/items/", h.List)
	g.POST("/items/", h.Create)
	g.PUT("/items/:id/", h.Update)
	g.DELETE("/items/:id/", h.Archive)
	g.GET("/items/:id/movements/", h.History)
	g.POST("/items/:id/movements/", h.Move)
}
