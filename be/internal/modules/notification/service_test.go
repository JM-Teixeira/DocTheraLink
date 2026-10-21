package notification

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"klyniva/internal/mailer"
	"klyniva/internal/modules/referral"
)

// fakeRepo is an in-memory Repository.
type fakeRepo struct {
	mu         sync.Mutex
	rows       []Notification
	people     map[uint]Person
	therapists map[string][]uint // clinic → physiotherapists ("" = all)
}

func (f *fakeRepo) CreateMany(_ context.Context, ns []Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range ns {
		_ = ns[i].BeforeCreate(nil)
		ns[i].CreatedAt = time.Now()
	}
	f.rows = append(f.rows, ns...)
	return nil
}

func (f *fakeRepo) List(_ context.Context, userID uint, q ListQuery) ([]Notification, int64, error) {
	var out []Notification
	for _, n := range f.rows {
		if n.UserID == userID && (!q.Unread || n.ReadAt == nil) {
			out = append(out, n)
		}
	}
	return out, int64(len(out)), nil
}

func (f *fakeRepo) UnreadCount(ctx context.Context, userID uint) (int64, error) {
	_, n, err := f.List(ctx, userID, ListQuery{Unread: true})
	return n, err
}

func (f *fakeRepo) MarkRead(_ context.Context, userID uint, id string, at time.Time) error {
	for i := range f.rows {
		if f.rows[i].ID == id && f.rows[i].UserID == userID {
			if f.rows[i].ReadAt == nil {
				f.rows[i].ReadAt = &at
			}
			return nil
		}
	}
	return ErrNotFound
}

func (f *fakeRepo) MarkAllRead(_ context.Context, userID uint, at time.Time) error {
	for i := range f.rows {
		if f.rows[i].UserID == userID && f.rows[i].ReadAt == nil {
			f.rows[i].ReadAt = &at
		}
	}
	return nil
}

func (f *fakeRepo) People(_ context.Context, ids []uint) ([]Person, error) {
	var out []Person
	for _, id := range ids {
		if p, ok := f.people[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakeRepo) Therapists(_ context.Context, clinicID *string) ([]uint, error) {
	key := ""
	if clinicID != nil {
		key = *clinicID
	}
	return f.therapists[key], nil
}

// fakeMailer records sent messages.
type fakeMailer struct{ sent []mailer.Message }

func (m *fakeMailer) Send(msg mailer.Message) error {
	m.sent = append(m.sent, msg)
	return nil
}

const (
	doctorID, sarahID, tiagoID, otherClinicPhysio = 1, 10, 11, 12
	clinicA                                       = "clinic-a"
)

func newTestService() (*Service, *fakeRepo, *fakeMailer) {
	repo := &fakeRepo{
		people: map[uint]Person{
			doctorID: {ID: doctorID, Email: "doctor@klyniva.test", FirstName: "John", LastName: "Smith"},
			sarahID:  {ID: sarahID, Email: "sarah@klyniva.test", FirstName: "Sarah", LastName: "Johnson"},
			tiagoID:  {ID: tiagoID, Email: "", FirstName: "Tiago", LastName: "Martins"}, // no e-mail
		},
		therapists: map[string][]uint{clinicA: {sarahID, tiagoID}},
	}
	m := &fakeMailer{}
	s := NewService(repo, m, "https://app.klyniva.test", nil)
	s.dispatch = func(f func()) { f() } // synchronous for tests
	return s, repo, m
}

func ref(therapist *uint) *referral.Referral {
	clinic := clinicA
	return &referral.Referral{
		ID: "r1", DoctorID: doctorID, TherapistID: therapist, FacilityID: &clinic,
		Patient: &referral.PatientSummary{Name: "Maria Silva"},
	}
}

func uintp(v uint) *uint { return &v }

func recipientsOf(repo *fakeRepo) []uint {
	var out []uint
	for _, n := range repo.rows {
		out = append(out, n.UserID)
	}
	return out
}

func TestRecipients(t *testing.T) {
	cases := []struct {
		name string
		ev   referral.Event
		want []uint
	}{
		{"open referral → all physios of the clinic",
			referral.Event{Kind: referral.EventCreated, Referral: ref(nil), ActorID: doctorID}, []uint{sarahID, tiagoID}},
		{"addressed referral → that physio",
			referral.Event{Kind: referral.EventCreated, Referral: ref(uintp(sarahID)), ActorID: doctorID}, []uint{sarahID}},
		{"accepted → doctor",
			referral.Event{Kind: referral.EventAccepted, Referral: ref(uintp(sarahID)), ActorID: sarahID}, []uint{doctorID}},
		{"cancelled → assigned therapist",
			referral.Event{Kind: referral.EventCancelled, Referral: ref(uintp(sarahID)), ActorID: doctorID}, []uint{sarahID}},
		{"cancelled while unassigned → nobody",
			referral.Event{Kind: referral.EventCancelled, Referral: ref(nil), ActorID: doctorID}, nil},
		{"hand-over by old therapist → doctor and new therapist",
			referral.Event{Kind: referral.EventReassigned, Referral: ref(uintp(tiagoID)), ActorID: sarahID,
				PreviousTherapistID: uintp(sarahID)}, []uint{doctorID, tiagoID}},
		{"hand-over by doctor → new and previous therapist",
			referral.Event{Kind: referral.EventReassigned, Referral: ref(uintp(tiagoID)), ActorID: doctorID,
				PreviousTherapistID: uintp(sarahID)}, []uint{tiagoID, sarahID}},
		{"message from doctor → therapist only",
			referral.Event{Kind: referral.EventMessage, Referral: ref(uintp(sarahID)), ActorID: doctorID}, []uint{sarahID}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, repo, _ := newTestService()
			s.ReferralEvent(context.Background(), tc.ev)
			if got := recipientsOf(repo); !slices.Equal(got, tc.want) {
				t.Fatalf("recipients = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNotificationContentAndEmail(t *testing.T) {
	s, repo, m := newTestService()
	s.ReferralEvent(context.Background(), referral.Event{
		Kind: referral.EventMessage, Referral: ref(uintp(sarahID)), ActorID: doctorID,
		Excerpt: strings.Repeat("x", 300),
	})
	if len(repo.rows) != 1 {
		t.Fatalf("rows: %d", len(repo.rows))
	}
	n := repo.rows[0]
	if n.ActorName != "John Smith" || n.PatientName != "Maria Silva" || *n.ReferralID != "r1" ||
		len([]rune(n.Excerpt)) != excerptLen || !strings.HasSuffix(n.Excerpt, "…") {
		t.Errorf("notification snapshot: %+v", n)
	}

	if len(m.sent) != 1 || m.sent[0].To != "sarah@klyniva.test" {
		t.Fatalf("emails: %+v", m.sent)
	}
	body := m.sent[0].Body + m.sent[0].Subject
	if strings.Contains(body, "Maria") || strings.Contains(body, "xxx") {
		t.Error("e-mail must not contain patient names or message content")
	}
	if !strings.Contains(body, "https://app.klyniva.test/referrals/r1") || !strings.Contains(body, "Olá Sarah") {
		t.Errorf("e-mail body: %s", body)
	}
}

func TestSkipsUsersWithoutEmailOrInactive(t *testing.T) {
	s, repo, m := newTestService()
	repo.therapists[clinicA] = append(repo.therapists[clinicA], 99) // inactive/unknown
	s.ReferralEvent(context.Background(), referral.Event{Kind: referral.EventCreated, Referral: ref(nil), ActorID: doctorID})
	if got := recipientsOf(repo); !slices.Equal(got, []uint{sarahID, tiagoID}) {
		t.Errorf("unknown users must be skipped, got %v", got)
	}
	if len(m.sent) != 1 {
		t.Errorf("only users with an e-mail get one, sent %d", len(m.sent))
	}
}

func TestReadState(t *testing.T) {
	ctx := context.Background()
	s, _, _ := newTestService()
	s.ReferralEvent(ctx, referral.Event{Kind: referral.EventCreated, Referral: ref(nil), ActorID: doctorID})
	s.ReferralEvent(ctx, referral.Event{Kind: referral.EventMessage, Referral: ref(uintp(sarahID)), ActorID: doctorID})

	page, _ := s.List(ctx, sarahID, ListQuery{})
	if page.Count != 2 || page.Unread != 2 {
		t.Fatalf("page: %+v", page)
	}
	if err := s.MarkRead(ctx, sarahID, page.Results[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkRead(ctx, tiagoID, page.Results[1].ID); err != ErrNotFound {
		t.Errorf("marking someone else's notification: want ErrNotFound, got %v", err)
	}
	if n, _ := s.UnreadCount(ctx, sarahID); n != 1 {
		t.Errorf("unread after one read: %d", n)
	}
	_ = s.MarkAllRead(ctx, sarahID)
	if n, _ := s.UnreadCount(ctx, sarahID); n != 0 {
		t.Errorf("unread after read-all: %d", n)
	}
	if n, _ := s.UnreadCount(ctx, tiagoID); n != 1 {
		t.Errorf("other users are unaffected, tiago unread: %d", n)
	}
}

func TestEmailLinesCoverEveryEvent(t *testing.T) {
	for _, k := range []referral.EventKind{referral.EventCreated, referral.EventAccepted, referral.EventDeclined,
		referral.EventCompleted, referral.EventCancelled, referral.EventReassigned, referral.EventReopened,
		referral.EventMessage} {
		if pt, en := emailLines(k); pt == "Tem uma atualização." || en == "You have an update." {
			t.Errorf("event %s has no specific e-mail text", k)
		}
	}
}
