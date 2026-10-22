// Package timeline puts together everything that happened to one patient:
// registration, referrals and their steps, thread messages, files and
// sessions. It owns no tables; it reads through the other modules' services,
// so each source keeps its own access rules (a receptionist sees sessions but
// no referrals, for example).
//
//	service.go - merging and ordering the entries
//	handler.go - GET /patient/:id/timeline/
//	routes.go  - adapters from the other modules' services
package timeline

import (
	"context"
	"sort"
	"strings"
	"time"

	"klyniva/internal/models"
	"klyniva/internal/modules/attachment"
	"klyniva/internal/modules/referral"
	"klyniva/internal/modules/schedule"
)

// Actor is the authenticated user asking.
type Actor struct {
	ID   uint
	Role models.UserRole
}

// Kind of a timeline entry.
type Kind string

const (
	KindRegistered Kind = "registered" // the patient was registered
	KindReferral   Kind = "referral"   // a referral step (Status: created, accepted, ...)
	KindNote       Kind = "note"       // a message on a referral thread (Status: note kind)
	KindFile       Kind = "file"       // a file attached to a referral
	KindSession    Kind = "session"    // a physio session (Status: booked, attended, missed)
)

// Entry is one thing that happened. Texts are short excerpts; the full
// content stays on the referral page.
type Entry struct {
	At         time.Time `json:"at"`
	Kind       Kind      `json:"kind"`
	Status     string    `json:"status,omitempty"`
	ReferralID string    `json:"referral_id,omitempty"`
	ItemID     string    `json:"item_id,omitempty"` // note, file or session id
	Text       string    `json:"text,omitempty"`
	Who        string    `json:"who,omitempty"`
	Pain       *int      `json:"pain,omitempty"`
	Date       string    `json:"date,omitempty"` // sessions: the day (YYYY-MM-DD)
	Hour       *int      `json:"hour,omitempty"` // sessions: start hour
	Justified  *bool     `json:"justified,omitempty"`
}

// Timeline is the answer: sessions still to come, then the history
// (newest first).
type Timeline struct {
	Upcoming []Entry `json:"upcoming"`
	History  []Entry `json:"history"`
}

// Sources are the other modules, seen through what the timeline needs. Each
// returns nothing (not an error) when the actor may not see that kind.
type Sources interface {
	Patient(ctx context.Context, a Actor, id string) (*models.Patient, error)
	Referrals(ctx context.Context, a Actor, patientID string) ([]referral.Referral, error)
	Notes(ctx context.Context, a Actor, referralID string) ([]referral.Note, error)
	Files(ctx context.Context, a Actor, referralID string) ([]attachment.Attachment, error)
	Sessions(ctx context.Context, a Actor, patientID string) ([]schedule.Appointment, error)
}

const excerptLen = 240

// Service builds timelines.
type Service struct {
	src Sources
	now func() time.Time
}

// NewService builds a Service; now may be nil to use time.Now.
func NewService(src Sources, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{src: src, now: now}
}

// For returns the timeline of a patient the actor can see.
func (s *Service) For(ctx context.Context, a Actor, patientID string) (*Timeline, error) {
	p, err := s.src.Patient(ctx, a, patientID)
	if err != nil {
		return nil, err
	}
	out := &Timeline{Upcoming: []Entry{}, History: []Entry{{
		At: p.CreatedAt, Kind: KindRegistered,
	}}}

	refs, err := s.src.Referrals(ctx, a, patientID)
	if err != nil {
		return nil, err
	}
	for i := range refs {
		r := &refs[i]
		out.History = append(out.History, referralSteps(r)...)
		notes, err := s.src.Notes(ctx, a, r.ID)
		if err != nil {
			return nil, err
		}
		for _, n := range notes {
			if n.Kind == referral.NoteSystem { // already shown as a referral step
				continue
			}
			out.History = append(out.History, Entry{
				At: n.CreatedAt, Kind: KindNote, Status: string(n.Kind), ReferralID: r.ID, ItemID: n.ID,
				Text: excerpt(n.Body), Who: summaryName(n.Author), Pain: n.Pain,
			})
		}
		files, err := s.src.Files(ctx, a, r.ID)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			out.History = append(out.History, Entry{
				At: f.CreatedAt, Kind: KindFile, ReferralID: r.ID, ItemID: f.ID,
				Text: f.FileName, Who: summaryName(f.Uploader),
			})
		}
	}

	sessions, err := s.src.Sessions(ctx, a, patientID)
	if err != nil {
		return nil, err
	}
	today := s.now().Format(time.DateOnly)
	for _, ap := range sessions {
		day := ap.Date.Format(time.DateOnly)
		hour := ap.Hour
		e := Entry{
			At:   time.Date(ap.Date.Year(), ap.Date.Month(), ap.Date.Day(), ap.Hour, 0, 0, 0, time.UTC),
			Kind: KindSession, Status: ap.Status, ItemID: ap.ID, Date: day, Hour: &hour,
			Justified: ap.MissJustified, Text: ap.MissReason,
		}
		if ap.ReferralID != nil {
			e.ReferralID = *ap.ReferralID
		}
		if ap.Therapist != nil {
			e.Who = strings.TrimSpace(ap.Therapist.FirstName + " " + ap.Therapist.LastName)
		}
		if ap.Status == schedule.StatusBooked && day >= today {
			out.Upcoming = append(out.Upcoming, e)
		} else {
			out.History = append(out.History, e)
		}
	}

	sort.SliceStable(out.Upcoming, func(i, j int) bool { return out.Upcoming[i].At.Before(out.Upcoming[j].At) })
	sort.SliceStable(out.History, func(i, j int) bool { return out.History[i].At.After(out.History[j].At) })
	return out, nil
}

// referralSteps turns a referral into its dated steps. Declined and
// cancelled have no own timestamp, so they use the last change.
func referralSteps(r *referral.Referral) []Entry {
	step := func(at time.Time, status, text, who string) Entry {
		return Entry{At: at, Kind: KindReferral, Status: status, ReferralID: r.ID, Text: excerpt(text), Who: who}
	}
	doctor, physio := summaryName(r.Doctor), summaryName(r.Therapist)
	out := []Entry{step(r.CreatedAt, "created", r.Diagnosis, doctor)}
	if r.AcceptedAt != nil {
		out = append(out, step(*r.AcceptedAt, "accepted", "", physio))
	}
	if r.StartedAt != nil {
		out = append(out, step(*r.StartedAt, "started", "", physio))
	}
	switch r.Status {
	case referral.StatusCompleted:
		at := r.UpdatedAt
		if r.CompletedAt != nil {
			at = *r.CompletedAt
		}
		out = append(out, step(at, "completed", deref(r.OutcomeSummary), physio))
	case referral.StatusDeclined:
		out = append(out, step(r.UpdatedAt, "declined", deref(r.DeclineReason), physio))
	case referral.StatusCancelled:
		out = append(out, step(r.UpdatedAt, "cancelled", deref(r.CancelReason), doctor))
	}
	return out
}

// excerpt shortens long texts on a word boundary.
func excerpt(s string) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= excerptLen {
		return s
	}
	cut := string(r[:excerptLen])
	if i := strings.LastIndexAny(cut, " \n"); i > excerptLen/2 {
		cut = cut[:i]
	}
	return strings.TrimSpace(cut) + "…"
}

func summaryName(u *referral.UserSummary) string {
	if u == nil {
		return ""
	}
	if n := strings.TrimSpace(u.FirstName + " " + u.LastName); n != "" {
		return n
	}
	return u.Username
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
