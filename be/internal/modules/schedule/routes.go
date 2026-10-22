package schedule

import (
	"klyniva/internal/clinic"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes wires the time-board API onto an authenticated group.
// hook runs when sessions are held (e.g. inventory uses consumables).
// It returns the service (referrals ask it who covers whom).
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB, hook SessionHook) *Service {
	svc := NewService(NewGormRepository(db), clinic.NewGormResolver(db)).WithSessionHook(hook)
	mount(protected.Group("/schedule"), NewHandler(svc))
	return svc
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
	g.GET("/absences/", h.Absences)
	g.POST("/absences/", h.CreateAbsence)
	g.DELETE("/absences/:id/", h.DeleteAbsence)
	g.GET("/handover/", h.Handover)
	g.GET("/missed/", h.MissedFlags)
	g.GET("/missed/:patientId/", h.PatientMisses)
	g.DELETE("/series/:id/", h.CancelSeries)
	g.GET("/series/:id/", h.SeriesSessions)
	g.POST("/series/:id/reschedule/", h.ReschedulePlan)
	g.POST("/appointments/:id/move/", h.MoveAppointment)
	g.GET("/waiting/", h.WaitingList)
	g.POST("/waiting/", h.AddWaiting)
	g.PATCH("/waiting/:id/", h.UpdateWaiting)
}
