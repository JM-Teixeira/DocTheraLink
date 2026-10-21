package referral

import "context"

// EventKind names a referral event that may notify people.
type EventKind string

const (
	EventCreated    EventKind = "referral_created"
	EventAccepted   EventKind = "referral_accepted"
	EventDeclined   EventKind = "referral_declined"
	EventCompleted  EventKind = "referral_completed"
	EventCancelled  EventKind = "referral_cancelled"
	EventReassigned EventKind = "referral_reassigned"
	EventReopened   EventKind = "referral_reopened"
	EventMessage    EventKind = "referral_message"
)

// Event describes something that happened to a referral.
type Event struct {
	Kind     EventKind
	Referral *Referral // state after the change, with Patient/Doctor/Therapist loaded
	ActorID  uint      // who did it (never notified about their own action)
	// PreviousTherapistID is set on hand-over so the old therapist is told.
	PreviousTherapistID *uint
	// Excerpt is the message text (EventMessage) or the given reason.
	Excerpt string
}

// Notifier receives referral events. Implementations must not fail the
// operation that produced the event: errors are theirs to log.
// The notification module implements it; the referral module does not
// depend on it.
type Notifier interface {
	ReferralEvent(ctx context.Context, e Event)
}

// noopNotifier is used when no notifier is configured (tests, tools).
type noopNotifier struct{}

func (noopNotifier) ReferralEvent(context.Context, Event) {}
