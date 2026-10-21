package notification

import (
	"context"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"klyniva/internal/mailer"
	"klyniva/internal/modules/referral"
)

const (
	defaultLimit = 20
	maxLimit     = 100
	excerptLen   = 160
)

// Service decides who is notified and delivers in-app + e-mail notifications.
type Service struct {
	repo      Repository
	mail      mailer.Mailer
	publicURL string
	now       func() time.Time
	// dispatch runs slow work (e-mail) off the request path; tests make it synchronous.
	dispatch func(func())
}

// NewService builds a Service. now may be nil to use time.Now.
func NewService(repo Repository, mail mailer.Mailer, publicURL string, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{repo: repo, mail: mail, publicURL: publicURL, now: now,
		dispatch: func(f func()) { go f() }}
}

// recipients returns who should hear about e: the people the referral
// concerns for that kind of event, never the actor themselves.
func (s *Service) recipients(ctx context.Context, e referral.Event) ([]uint, error) {
	r := e.Referral
	var ids []uint
	switch e.Kind {
	case referral.EventCreated, referral.EventReopened:
		if r.TherapistID != nil {
			ids = []uint{*r.TherapistID}
		} else {
			// Open referral: every physiotherapist of the clinic may take it.
			var err error
			if ids, err = s.repo.Therapists(ctx, r.FacilityID); err != nil {
				return nil, err
			}
		}
	case referral.EventAccepted, referral.EventDeclined, referral.EventCompleted:
		ids = []uint{r.DoctorID}
	case referral.EventCancelled:
		if r.TherapistID != nil {
			ids = []uint{*r.TherapistID}
		}
	case referral.EventReassigned:
		ids = []uint{r.DoctorID}
		if r.TherapistID != nil {
			ids = append(ids, *r.TherapistID)
		}
		if e.PreviousTherapistID != nil {
			ids = append(ids, *e.PreviousTherapistID)
		}
	case referral.EventMessage:
		ids = []uint{r.DoctorID}
		if r.TherapistID != nil {
			ids = append(ids, *r.TherapistID)
		}
	}
	return withoutActor(ids, e.ActorID), nil
}

// withoutActor removes the actor, zeros and duplicates, keeping order.
func withoutActor(ids []uint, actor uint) []uint {
	seen := map[uint]bool{actor: true, 0: true}
	out := ids[:0:0]
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// ReferralEvent implements referral.Notifier. It never returns an error:
// failures are logged so the referral operation itself is unaffected.
func (s *Service) ReferralEvent(ctx context.Context, e referral.Event) {
	if err := s.notify(ctx, e); err != nil {
		log.Printf("notification: %s for referral %s: %v", e.Kind, e.Referral.ID, err)
	}
}

func (s *Service) notify(ctx context.Context, e referral.Event) error {
	ids, err := s.recipients(ctx, e)
	if err != nil || len(ids) == 0 {
		return err
	}
	people, err := s.repo.People(ctx, append(ids, e.ActorID))
	if err != nil {
		return err
	}
	byID := make(map[uint]Person, len(people))
	for _, p := range people {
		byID[p.ID] = p
	}

	actorID := e.ActorID
	patient := ""
	if e.Referral.Patient != nil {
		patient = e.Referral.Patient.Name
	}
	var rows []Notification
	var emails []Person
	for _, id := range ids {
		p, ok := byID[id]
		if !ok {
			continue // inactive or deleted
		}
		rows = append(rows, Notification{
			UserID: id, Event: string(e.Kind), ReferralID: &e.Referral.ID,
			ActorID: &actorID, ActorName: byID[e.ActorID].FullName(),
			PatientName: patient, Excerpt: truncate(e.Excerpt, excerptLen),
		})
		if p.Email != "" {
			emails = append(emails, p)
		}
	}
	if err := s.repo.CreateMany(ctx, rows); err != nil {
		return err
	}

	link := s.publicURL + "/referrals/" + e.Referral.ID
	s.dispatch(func() {
		for _, p := range emails {
			if err := s.mail.Send(EmailFor(e.Kind, p, link)); err != nil {
				log.Printf("notification: e-mail to user %d failed: %v", p.ID, err)
			}
		}
	})
	return nil
}

// EmailFor builds the e-mail for an event. It deliberately contains no
// patient or clinical information — e-mail is not a safe channel for health
// data — only a generic line and a link that requires signing in.
func EmailFor(kind referral.EventKind, to Person, link string) mailer.Message {
	pt, en := emailLines(kind)
	body := strings.Join([]string{
		"Olá " + to.FirstName + ",", "", pt, "Abrir: " + link, "",
		"—", "", "Hello " + to.FirstName + ",", "", en, "Open: " + link, "",
		"Klyniva",
	}, "\n")
	return mailer.Message{To: to.Email, Subject: "Klyniva: " + en, Body: body}
}

func emailLines(kind referral.EventKind) (pt, en string) {
	switch kind {
	case referral.EventCreated:
		return "Tem um novo encaminhamento.", "You have a new referral."
	case referral.EventAccepted:
		return "Um encaminhamento seu foi aceite.", "One of your referrals was accepted."
	case referral.EventDeclined:
		return "Um encaminhamento seu foi recusado.", "One of your referrals was declined."
	case referral.EventCompleted:
		return "Um tratamento foi concluído.", "A treatment was completed."
	case referral.EventCancelled:
		return "Um encaminhamento foi cancelado.", "A referral was cancelled."
	case referral.EventReassigned:
		return "Um encaminhamento foi transferido.", "A referral was handed over."
	case referral.EventReopened:
		return "Um encaminhamento foi reaberto.", "A referral was re-opened."
	case referral.EventMessage:
		return "Tem uma nova mensagem num encaminhamento.", "You have a new message on a referral."
	}
	return "Tem uma atualização.", "You have an update."
}

// truncate shortens s to at most n runes, adding "…" when cut.
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

// List returns the user's notifications, newest first, with the unread count.
func (s *Service) List(ctx context.Context, userID uint, q ListQuery) (*Page, error) {
	if q.Limit <= 0 {
		q.Limit = defaultLimit
	}
	q.Limit = min(q.Limit, maxLimit)
	q.Offset = max(q.Offset, 0)
	items, total, err := s.repo.List(ctx, userID, q)
	if err != nil {
		return nil, err
	}
	unread, err := s.repo.UnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []Notification{}
	}
	return &Page{Count: total, Unread: unread, Results: items}, nil
}

// UnreadCount is the number on the bell.
func (s *Service) UnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.repo.UnreadCount(ctx, userID)
}

// MarkRead marks one of the user's notifications as read (idempotent).
func (s *Service) MarkRead(ctx context.Context, userID uint, id string) error {
	return s.repo.MarkRead(ctx, userID, id, s.now())
}

// MarkAllRead marks all of the user's notifications as read.
func (s *Service) MarkAllRead(ctx context.Context, userID uint) error {
	return s.repo.MarkAllRead(ctx, userID, s.now())
}
