package notification

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the notification API onto an authenticated group.
func RegisterRoutes(protected *gin.RouterGroup, svc *Service) {
	mount(protected.Group("/notifications"), NewHandler(svc))
}

// mount is the route table, shared with the handler tests.
func mount(g *gin.RouterGroup, h *Handler) {
	g.GET("/", h.List)
	g.GET("/unread-count/", h.UnreadCount)
	g.POST("/read-all/", h.MarkAllRead)
	g.POST("/:id/read/", h.MarkRead)
}
