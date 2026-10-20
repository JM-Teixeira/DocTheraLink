package referral

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires the module onto an authenticated router group.
// The group must already run auth.AuthMiddleware.
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(NewService(NewGormRepository(db), nil))

	g := protected.Group("/referrals")
	g.GET("/", h.List)
	g.POST("/", h.Create)
	g.GET("/:id/", h.Get)
	g.POST("/:id/accept/", h.Accept)
	g.POST("/:id/decline/", h.Decline)
	g.POST("/:id/start/", h.Start)
	g.POST("/:id/complete/", h.Complete)
	g.POST("/:id/cancel/", h.Cancel)
	g.GET("/:id/notes/", h.ListNotes)
	g.POST("/:id/notes/", h.AddNote)
}
