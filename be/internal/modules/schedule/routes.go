package schedule

import (
	"klyniva/internal/clinic"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires the time-board API onto an authenticated group.
// hook runs when sessions are held (e.g. inventory uses consumables).
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB, hook SessionHook) {
	svc := NewService(NewGormRepository(db), clinic.NewGormResolver(db)).WithSessionHook(hook)
	mount(protected.Group("/schedule"), NewHandler(svc))
}

// mount is the route table, shared with the handler tests.
func mount(g *gin.RouterGroup, h *Handler) {
	g.GET("/board/", h.Board)
	g.GET("/range/", h.Range)
	g.GET("/year/", h.Year)

	g.GET("/divisions/", h.Divisions)
	g.POST("/divisions/", h.CreateDivision)
	g.PUT("/divisions/:id/", h.UpdateDivision)
	g.DELETE("/divisions/:id/", h.RemoveDivision)

	g.POST("/shifts/", h.CreateShift)
	g.POST("/shifts/copy/", h.CopyDay)
	g.POST("/shifts/repeat-week/", h.RepeatWeek)
	g.PUT("/shifts/:id/", h.UpdateShift)
	g.DELETE("/shifts/:id/", h.DeleteShift)

	g.POST("/appointments/", h.Book)
	g.PATCH("/appointments/:id/", h.SetStatus)
	g.DELETE("/appointments/:id/", h.Unbook)

	g.POST("/series/", h.Series)
	g.DELETE("/series/:id/", h.CancelSeries)
}
