package feedback

import (
	"context"
	"errors"
	"strings"
	"testing"

	"klyniva/internal/clinic"
	"klyniva/internal/models"
)

type fakeRepo struct{ items []Suggestion }

func (f *fakeRepo) Add(_ context.Context, s *Suggestion) error {
	_ = s.BeforeCreate(nil)
	f.items = append(f.items, *s)
	return nil
}
func (f *fakeRepo) List(_ context.Context, authorID uint, status string) ([]Suggestion, error) {
	var out []Suggestion
	for _, s := range f.items {
		if (authorID == 0 || s.AuthorID == authorID) && (status == "" || s.Status == status) {
			out = append(out, s)
		}
	}
	return out, nil
}
func (f *fakeRepo) Get(_ context.Context, id string) (*Suggestion, error) {
	for i := range f.items {
		if f.items[i].ID == id {
			s := f.items[i]
			return &s, nil
		}
	}
	return nil, ErrNotFound
}
func (f *fakeRepo) Update(_ context.Context, id string, fields map[string]any) error {
	for i := range f.items {
		if f.items[i].ID == id {
			if v, ok := fields["status"]; ok {
				f.items[i].Status = v.(string)
			}
			if v, ok := fields["admin_note"]; ok {
				f.items[i].AdminNote = v.(string)
			}
		}
	}
	return nil
}
func (f *fakeRepo) CountNew(context.Context) (int64, error) {
	var n int64
	for _, s := range f.items {
		if s.Status == "new" {
			n++
		}
	}
	return n, nil
}

var (
	physio = Actor{ID: 3, Role: models.RolePhysioTherapist}
	nurse  = Actor{ID: 4, Role: models.RoleNurse}
	admin  = Actor{ID: 1, Role: models.RoleStateAdmin}
)

func TestSuggestions(t *testing.T) {
	ctx := context.Background()
	s := NewService(&fakeRepo{}, clinic.Static{physio.ID: "c1", nurse.ID: "c1"})

	sg, err := s.Send(ctx, physio, SendRequest{Kind: "idea", Page: "/time-board", Text: "  Imprimir o dia do fisioterapeuta  "})
	if err != nil || sg.Status != "new" || sg.Text != "Imprimir o dia do fisioterapeuta" || *sg.ClinicID != "c1" {
		t.Fatalf("send: %+v %v", sg, err)
	}
	if _, err := s.Send(ctx, nurse, SendRequest{Kind: "problem", Text: "O botão X não funciona"}); err != nil {
		t.Fatal(err)
	}
	for name, req := range map[string]SendRequest{
		"short": {Kind: "idea", Text: "ok"},
		"long":  {Kind: "idea", Text: strings.Repeat("a", maxText+1)},
		"kind":  {Kind: "spam", Text: "uma ideia qualquer"},
	} {
		var ve *ValidationError
		if _, err := s.Send(ctx, physio, req); !errors.As(err, &ve) {
			t.Errorf("%s: %v", name, err)
		}
	}

	// Each person sees their own; only admins see all and triage
	if mine, _ := s.Mine(ctx, physio); len(mine) != 1 {
		t.Errorf("mine: %d", len(mine))
	}
	if _, err := s.All(ctx, physio, ""); !errors.Is(err, ErrForbidden) {
		t.Errorf("physio lists all: %v", err)
	}
	if n, _ := s.CountNew(ctx, admin); n != 2 {
		t.Errorf("new: %d", n)
	}
	note := " Boa ideia, fica para a próxima versão "
	got, err := s.Review(ctx, admin, sg.ID, ReviewRequest{Status: "planned", AdminNote: &note})
	if err != nil || got.Status != "planned" || got.AdminNote != "Boa ideia, fica para a próxima versão" {
		t.Errorf("review: %+v %v", got, err)
	}
	if all, _ := s.All(ctx, admin, "new"); len(all) != 1 {
		t.Errorf("filter new: %d", len(all))
	}
	if _, err := s.Review(ctx, physio, sg.ID, ReviewRequest{Status: "done"}); !errors.Is(err, ErrForbidden) {
		t.Errorf("physio reviews: %v", err)
	}
	if _, err := s.Review(ctx, admin, sg.ID, ReviewRequest{Status: "lost"}); err == nil {
		t.Error("unknown status")
	}
	if _, err := s.Review(ctx, admin, "nope", ReviewRequest{}); !errors.Is(err, ErrNotFound) {
		t.Errorf("bad id: %v", err)
	}
}
