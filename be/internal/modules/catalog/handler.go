package catalog

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultLimit = 20
	maxLimit     = 50
)

// Handler exposes the catalog over HTTP.
type Handler struct{ c *Catalog }

// NewHandler creates the handler.
func NewHandler(c *Catalog) *Handler { return &Handler{c: c} }

// lang reads ?lang=, falling back to the Accept-Language header.
func lang(ctx *gin.Context) string {
	if l := ctx.Query("lang"); l != "" {
		return Lang(l)
	}
	return Lang(ctx.GetHeader("Accept-Language"))
}

// ICD10 handles GET /catalog/icd10/?search=&lang=&limit=
func (h *Handler) ICD10(ctx *gin.Context) {
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}
	ctx.JSON(http.StatusOK, h.c.Search(ctx.Query("search"), lang(ctx), min(limit, maxLimit)))
}

// Templates handles GET /catalog/referral-templates/?lang=
func (h *Handler) Templates(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.c.Templates(lang(ctx)))
}
