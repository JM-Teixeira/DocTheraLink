package schedule

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"klyniva/internal/auth"
	"klyniva/internal/models"

	"github.com/gin-gonic/gin"
)

// Handler exposes the Service over HTTP.
type Handler struct{ svc *Service }

// NewHandler builds a Handler around svc.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func actorFrom(c *gin.Context) (Actor, bool) {
	claims, err := auth.GetCurrentUserClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized", Message: "Invalid token"})
		return Actor{}, false
	}
	return Actor{ID: claims.UserID, Role: claims.UserType}, true
}

func writeError(c *gin.Context, err error) {
	var verr *ValidationError
	var cerr *ConflictError
	switch {
	case errors.As(err, &verr):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "validation_error", Message: verr.Error(), Details: map[string]string{verr.Field: verr.Msg},
		})
	case errors.As(err, &cerr):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "conflict", Message: cerr.Msg})
	case errors.Is(err, ErrFull):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "hour_full", Message: err.Error()})
	case errors.Is(err, ErrOverCredits):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "over_credits", Message: err.Error()})
	case errors.Is(err, ErrPatientBusy):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "patient_busy", Message: err.Error()})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrNoClinic):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: err.Error()})
	default:
		log.Printf("schedule: internal error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Unexpected error"})
	}
}

// handle runs one call: actor, optional JSON body, service, response.
func handle[Req any](c *gin.Context, status int, body bool, run func(a Actor, req Req) (any, error)) {
	a, ok := actorFrom(c)
	if !ok {
		return
	}
	var req Req
	if body {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Message: err.Error()})
			return
		}
	}
	out, err := run(a, req)
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		c.Status(status)
		return
	}
	c.JSON(status, out)
}

type none struct{}

// Board handles GET /schedule/board/?date= (today by default).
func (h *Handler) Board(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		return h.svc.Board(c.Request.Context(), a, c.DefaultQuery("date", time.Now().Format(dateLayout)))
	})
}

// Range handles GET /schedule/range/?from=&to= (week and month views).
func (h *Handler) Range(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		return h.svc.Range(c.Request.Context(), a, c.Query("from"), c.Query("to"))
	})
}

// Year handles GET /schedule/year/?year= (this year by default).
func (h *Handler) Year(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		year, err := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
		if err != nil {
			return nil, &ValidationError{"year", "must be a number"}
		}
		return h.svc.Year(c.Request.Context(), a, year)
	})
}

// Series handles POST /schedule/series/ (dry_run previews).
func (h *Handler) Series(c *gin.Context) {
	status := http.StatusCreated
	handle(c, status, true, func(a Actor, req SeriesRequest) (any, error) {
		return h.svc.Series(c.Request.Context(), a, req)
	})
}

// CancelSeries handles DELETE /schedule/series/:id/?from= (today by default).
func (h *Handler) CancelSeries(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		n, err := h.svc.CancelSeries(c.Request.Context(), a, c.Param("id"), c.DefaultQuery("from", time.Now().Format(dateLayout)))
		return gin.H{"removed": n}, err
	})
}

// RepeatWeek handles POST /schedule/shifts/repeat-week/.
func (h *Handler) RepeatWeek(c *gin.Context) {
	handle(c, http.StatusOK, true, func(a Actor, req RepeatWeekRequest) (any, error) {
		n, err := h.svc.RepeatWeek(c.Request.Context(), a, req)
		return gin.H{"copied": n}, err
	})
}

// Absences handles GET /schedule/absences/?from=&to=
func (h *Handler) Absences(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		today := time.Now().Format(dateLayout)
		return h.svc.Absences(c.Request.Context(), a, c.DefaultQuery("from", today), c.DefaultQuery("to", today))
	})
}

// CreateAbsence handles POST /schedule/absences/ (dry_run previews).
func (h *Handler) CreateAbsence(c *gin.Context) {
	handle(c, http.StatusCreated, true, func(a Actor, req AbsenceRequest) (any, error) {
		return h.svc.CreateAbsence(c.Request.Context(), a, req)
	})
}

// DeleteAbsence handles DELETE /schedule/absences/:id/
func (h *Handler) DeleteAbsence(c *gin.Context) {
	handle(c, http.StatusNoContent, false, func(a Actor, _ none) (any, error) {
		return nil, h.svc.DeleteAbsence(c.Request.Context(), a, c.Param("id"))
	})
}

// MissedFlags handles GET /schedule/missed/ (patients who miss too often).
func (h *Handler) MissedFlags(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		return h.svc.MissedFlags(c.Request.Context(), a)
	})
}

// PatientMisses handles GET /schedule/missed/:patientId/
func (h *Handler) PatientMisses(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		return h.svc.PatientMisses(c.Request.Context(), a, c.Param("patientId"))
	})
}

// Handover handles GET /schedule/handover/?patient=
func (h *Handler) Handover(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		return h.svc.Handover(c.Request.Context(), a, c.Query("patient"))
	})
}

func (h *Handler) Divisions(c *gin.Context) {
	handle(c, http.StatusOK, false, func(a Actor, _ none) (any, error) {
		return h.svc.Divisions(c.Request.Context(), a)
	})
}

func (h *Handler) CreateDivision(c *gin.Context) {
	handle(c, http.StatusCreated, true, func(a Actor, req DivisionRequest) (any, error) {
		return h.svc.SaveDivision(c.Request.Context(), a, "", req)
	})
}

func (h *Handler) UpdateDivision(c *gin.Context) {
	handle(c, http.StatusOK, true, func(a Actor, req DivisionRequest) (any, error) {
		return h.svc.SaveDivision(c.Request.Context(), a, c.Param("id"), req)
	})
}

func (h *Handler) RemoveDivision(c *gin.Context) {
	handle(c, http.StatusNoContent, false, func(a Actor, _ none) (any, error) {
		return nil, h.svc.RemoveDivision(c.Request.Context(), a, c.Param("id"))
	})
}

func (h *Handler) CreateShift(c *gin.Context) {
	handle(c, http.StatusCreated, true, func(a Actor, req ShiftRequest) (any, error) {
		return h.svc.SaveShift(c.Request.Context(), a, "", req)
	})
}

func (h *Handler) UpdateShift(c *gin.Context) {
	handle(c, http.StatusOK, true, func(a Actor, req ShiftRequest) (any, error) {
		return h.svc.SaveShift(c.Request.Context(), a, c.Param("id"), req)
	})
}

func (h *Handler) DeleteShift(c *gin.Context) {
	handle(c, http.StatusNoContent, false, func(a Actor, _ none) (any, error) {
		return nil, h.svc.DeleteShift(c.Request.Context(), a, c.Param("id"))
	})
}

func (h *Handler) CopyDay(c *gin.Context) {
	handle(c, http.StatusOK, true, func(a Actor, req CopyRequest) (any, error) {
		n, err := h.svc.CopyDay(c.Request.Context(), a, req)
		return gin.H{"copied": n}, err
	})
}

func (h *Handler) Book(c *gin.Context) {
	handle(c, http.StatusCreated, true, func(a Actor, req AppointmentRequest) (any, error) {
		return h.svc.Book(c.Request.Context(), a, req)
	})
}

func (h *Handler) SetStatus(c *gin.Context) {
	handle(c, http.StatusOK, true, func(a Actor, req StatusRequest) (any, error) {
		return h.svc.SetStatus(c.Request.Context(), a, c.Param("id"), req)
	})
}

func (h *Handler) Unbook(c *gin.Context) {
	handle(c, http.StatusNoContent, false, func(a Actor, _ none) (any, error) {
		return nil, h.svc.Unbook(c.Request.Context(), a, c.Param("id"))
	})
}
