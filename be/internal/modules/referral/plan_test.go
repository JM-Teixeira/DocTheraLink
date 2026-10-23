package referral

import (
	"context"
	"errors"
	"strings"
	"testing"

	"klyniva/internal/clinic"
	"klyniva/internal/models"
)

func intp(v int) *int { return &v }

func TestTreatmentPlan(t *testing.T) {
	ctx := context.Background()
	s, _ := newTestService(t)
	coordinator := Actor{ID: 30, Role: models.RoleCoordinator}
	s.clinics = clinic.Static{doctor.ID: clinicA, physio.ID: clinicA, otherPhysio.ID: clinicA, coordinator.ID: clinicA}
	r := mustCreate(t, s, ptrUint(physio.ID))
	steps := PlanRequest{Items: []PlanItem{
		{Activity: " Ultrassons ", Minutes: intp(10), Equipment: "Máquina 2"},
		{Activity: "Massagem", Minutes: intp(5)},
		{Activity: "  "}, // empty rows are dropped
		{Activity: "Exercícios de Codman", Dose: "3x10"},
	}}

	// Nothing to plan before the physio accepts
	if v, err := s.Plan(ctx, physio, r.ID); err != nil || v.CanEdit || v.Current != nil {
		t.Fatalf("pending: %v %+v", err, v)
	}
	if _, err := s.SavePlan(ctx, physio, r.ID, steps); !errors.Is(err, ErrForbidden) {
		t.Errorf("pending save: %v", err)
	}
	if _, err := s.Accept(ctx, physio, r.ID); err != nil {
		t.Fatal(err)
	}

	v, err := s.SavePlan(ctx, physio, r.ID, steps)
	if err != nil {
		t.Fatal(err)
	}
	if !v.CanEdit || len(v.Current.Items) != 3 || v.Current.Items[0].Activity != "Ultrassons" || v.Current.AuthorID != physio.ID {
		t.Errorf("saved: %+v", v.Current)
	}
	// A new save is a new version; the newest is current
	steps.Items = steps.Items[:1]
	if v, err = s.SavePlan(ctx, coordinator, r.ID, steps); err != nil || len(v.Versions) != 2 || len(v.Current.Items) != 1 {
		t.Errorf("second version: %v %+v", err, v)
	}

	// The doctor reads but does not edit; other physios do not even see it
	if dv, err := s.Plan(ctx, doctor, r.ID); err != nil || dv.CanEdit || dv.Current == nil {
		t.Errorf("doctor view: %v %+v", err, dv)
	}
	if _, err := s.SavePlan(ctx, doctor, r.ID, steps); !errors.Is(err, ErrForbidden) {
		t.Errorf("doctor save: %v", err)
	}
	if _, err := s.Plan(ctx, otherPhysio, r.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("other physio: %v", err)
	}

	for name, bad := range map[string]PlanItem{
		"minutes": {Activity: "x", Minutes: intp(0)},
		"long":    {Activity: strings.Repeat("a", 121)},
	} {
		var ve *ValidationError
		if _, err := s.SavePlan(ctx, physio, r.ID, PlanRequest{Items: []PlanItem{bad}}); !errors.As(err, &ve) {
			t.Errorf("%s: %v", name, err)
		}
	}
	many := make([]PlanItem, maxPlanItems+1)
	for i := range many {
		many[i].Activity = "x"
	}
	var ve *ValidationError
	if _, err := s.SavePlan(ctx, physio, r.ID, PlanRequest{Items: many}); !errors.As(err, &ve) {
		t.Errorf("too many: %v", err)
	}
}
