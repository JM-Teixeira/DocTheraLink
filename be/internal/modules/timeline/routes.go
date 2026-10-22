package timeline

import (
	"context"
	"errors"

	"klyniva/internal/models"
	"klyniva/internal/modules/attachment"
	"klyniva/internal/modules/patient"
	"klyniva/internal/modules/referral"
	"klyniva/internal/modules/schedule"

	"github.com/gin-gonic/gin"
)

// Services are the modules the timeline reads from.
type Services struct {
	Patients  *patient.Service
	Referrals *referral.Service
	Files     *attachment.Service
	Board     *schedule.Service
}

// RegisterRoutes wires GET /patient/:id/timeline/ onto an authenticated group.
func RegisterRoutes(protected *gin.RouterGroup, s Services) {
	protected.GET("/patient/:id/timeline/", NewHandler(NewService(sources{s}, nil)).Get)
}

// sources adapts the services. A source the actor may not use (no clinic,
// a role without referrals) gives no entries instead of failing the page.
type sources struct{ s Services }

func hidden(err error) bool {
	return errors.Is(err, referral.ErrForbidden) || errors.Is(err, referral.ErrNotFound) ||
		errors.Is(err, attachment.ErrForbidden) ||
		errors.Is(err, schedule.ErrForbidden) || errors.Is(err, schedule.ErrNoClinic)
}

func visible[T any](items []T, err error) ([]T, error) {
	if hidden(err) {
		return nil, nil
	}
	return items, err
}

func (x sources) Patient(ctx context.Context, a Actor, id string) (*models.Patient, error) {
	return x.s.Patients.Get(ctx, patient.Actor{ID: a.ID, Role: a.Role}, id)
}

func (x sources) Referrals(ctx context.Context, a Actor, patientID string) ([]referral.Referral, error) {
	items, _, err := x.s.Referrals.List(ctx, referral.Actor{ID: a.ID, Role: a.Role},
		referral.ListFilter{PatientID: patientID, Limit: 100})
	return visible(items, err)
}

func (x sources) Notes(ctx context.Context, a Actor, referralID string) ([]referral.Note, error) {
	return visible(x.s.Referrals.Notes(ctx, referral.Actor{ID: a.ID, Role: a.Role}, referralID))
}

func (x sources) Files(ctx context.Context, a Actor, referralID string) ([]attachment.Attachment, error) {
	return visible(x.s.Files.List(ctx, referral.Actor{ID: a.ID, Role: a.Role}, referralID))
}

func (x sources) Sessions(ctx context.Context, a Actor, patientID string) ([]schedule.Appointment, error) {
	return visible(x.s.Board.PatientSessions(ctx, schedule.Actor{ID: a.ID, Role: a.Role}, patientID))
}
