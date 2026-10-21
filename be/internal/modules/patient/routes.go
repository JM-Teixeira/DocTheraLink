package patient

import (
	"klyniva/internal/clinic"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires the module's endpoints onto an existing /patient group.
// The group must already run auth.AuthMiddleware.
func RegisterRoutes(patients *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(NewService(NewGormRepository(db), clinic.NewGormResolver(db), nil))

	patients.GET("/", h.List)
	patients.POST("/", h.Create)
	// search/ is kept for the frontend duplicate check; it accepts the same filters as the list.
	patients.GET("/search/", h.List)
	patients.GET("/stats/", h.Stats)
	patients.GET("/:id/", h.Get)
	patients.PUT("/:id/", h.Update)
	patients.PATCH("/:id/", h.Update)
}
