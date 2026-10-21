package catalog

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the reference-data API onto an authenticated group.
func RegisterRoutes(protected *gin.RouterGroup, c *Catalog) {
	mount(protected.Group("/catalog"), NewHandler(c))
}

// mount is the route table, shared with the handler tests.
func mount(g *gin.RouterGroup, h *Handler) {
	g.GET("/icd10/", h.ICD10)
	g.GET("/referral-templates/", h.Templates)
}
