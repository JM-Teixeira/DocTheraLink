package attachment

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the attachment API onto an authenticated group.
func RegisterRoutes(protected *gin.RouterGroup, svc *Service) {
	mount(protected, NewHandler(svc))
}

// mount is the route table, shared with the handler tests. The referral
// routes use the same ":id" parameter name, as Gin requires.
func mount(g *gin.RouterGroup, h *Handler) {
	g.GET("/referrals/:id/attachments/", h.List)
	g.POST("/referrals/:id/attachments/", h.Upload)
	g.GET("/attachments/:id/download/", h.Download)
	g.DELETE("/attachments/:id/", h.Delete)
}
