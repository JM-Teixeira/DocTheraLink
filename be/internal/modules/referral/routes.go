package referral

import (
	"klyniva/internal/clinic"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires the module onto an authenticated router group and
// returns the service so other modules (e.g. attachments) can reuse its
// access rules. The group must already run auth.AuthMiddleware. notifier is
// told about referral events (in-app notifications, e-mail).
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB, notifier Notifier) *Service {
	svc := NewService(NewGormRepository(db), clinic.NewGormResolver(db), nil).WithNotifier(notifier)
	mount(protected.Group("/referrals"), NewHandler(svc))
	return svc
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
	g.GET("/:id/plan/", h.GetPlan)
	g.PUT("/:id/plan/", h.SavePlan)
	g.POST("/:id/notes/", h.AddNote)
}
