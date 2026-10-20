package referral

import (
	"klyniva/internal/clinic"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires the module onto an authenticated router group.
// The group must already run auth.AuthMiddleware.
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(NewService(NewGormRepository(db), clinic.NewGormResolver(db), nil))
	mount(protected.Group("/referrals"), h)
}

// mount is the route table, shared by RegisterRoutes and the handler tests.
func mount(g *gin.RouterGroup, h *Handler) {
	g.GET("/", h.List)
	g.POST("/", h.Create)
	g.GET("/counts/", h.Counts)
	g.GET("/:id/", h.Get)
	g.POST("/:id/accept/", h.Accept)
	g.POST("/:id/decline/", h.Decline)
	g.POST("/:id/start/", h.Start)
	g.POST("/:id/complete/", h.Complete)
	g.POST("/:id/cancel/", h.Cancel)
	g.POST("/:id/reassign/", h.Reassign)
	g.POST("/:id/reopen/", h.Reopen)
	g.GET("/:id/notes/", h.ListNotes)
	g.POST("/:id/notes/", h.AddNote)
}
