//go:build integration

package integration

import (
	"context"
	"testing"

	"klyniva/internal/clinic"
	"klyniva/internal/database"
	"klyniva/internal/mailer"
	"klyniva/internal/models"
	"klyniva/internal/modules/notification"
	"klyniva/internal/modules/referral"
)

// TestReferralNotificationsOnPostgres: an open referral notifies every active
// physiotherapist of the doctor's clinic — not other clinics — and acting on
// it notifies the doctor.
func TestReferralNotificationsOnPostgres(t *testing.T) {
	startDB(t)
	ctx := context.Background()
	db := database.GetDB()

	notes := notification.NewService(notification.NewGormRepository(db), mailer.LogMailer{}, "http://test", nil)
	refs := referral.NewService(referral.NewGormRepository(db), clinic.NewGormResolver(db), nil).WithNotifier(notes)

	other := models.Facility{Name: "Elsewhere"}
	db.Create(&other)
	far := newUser(t, "far-physio", models.RolePhysioTherapist, &other.ID)
	doctor := userID(t, "doctor")
	sarah := userID(t, "physiotherapist")
	tiago := userID(t, "physiotherapist2")

	p := models.Patient{Name: "Notified Patient"}
	db.Create(&p)
	r, err := refs.Create(ctx, doctor, referral.CreateRequest{PatientID: p.ID, Diagnosis: "Sprain", Reason: "Rehab"})
	if err != nil {
		t.Fatal(err)
	}

	unread := func(a referral.Actor) int64 {
		n, err := notes.UnreadCount(ctx, a.ID)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	if unread(sarah) != 1 || unread(tiago) != 1 {
		t.Errorf("clinic physios should be notified: sarah=%d tiago=%d", unread(sarah), unread(tiago))
	}
	if unread(far) != 0 || unread(doctor) != 0 {
		t.Errorf("other clinic and the actor must not be notified: far=%d doctor=%d", unread(far), unread(doctor))
	}

	if _, err := refs.Accept(ctx, sarah, r.ID); err != nil {
		t.Fatal(err)
	}
	page, err := notes.List(ctx, doctor.ID, notification.ListQuery{})
	if err != nil || page.Unread != 1 || page.Results[0].Event != string(referral.EventAccepted) ||
		page.Results[0].ActorName != "Sarah Johnson" || page.Results[0].PatientName != "Notified Patient" {
		t.Fatalf("doctor notification: %+v %v", page, err)
	}
	if err := notes.MarkRead(ctx, doctor.ID, page.Results[0].ID); err != nil {
		t.Fatal(err)
	}
	if unread(doctor) != 0 {
		t.Error("mark read did not persist")
	}
}
