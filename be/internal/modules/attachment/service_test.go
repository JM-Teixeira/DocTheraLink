package attachment

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"klyniva/internal/models"
	"klyniva/internal/modules/referral"
	"klyniva/internal/storage"
)

// fakeRepo is an in-memory Repository.
type fakeRepo struct{ rows map[string]Attachment }

func (f *fakeRepo) Create(_ context.Context, a *Attachment) error {
	a.CreatedAt = time.Now()
	f.rows[a.ID] = *a
	return nil
}
func (f *fakeRepo) Get(_ context.Context, id string) (*Attachment, error) {
	a, ok := f.rows[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &a, nil
}
func (f *fakeRepo) ListByReferral(_ context.Context, rid string) ([]Attachment, error) {
	var out []Attachment
	for _, a := range f.rows {
		if a.ReferralID == rid {
			out = append(out, a)
		}
	}
	return out, nil
}
func (f *fakeRepo) Delete(_ context.Context, id string) error {
	delete(f.rows, id)
	return nil
}

// fakeAccess: who can see referral r1 (doctor 1, therapist 10, nurse 20 in
// the same clinic sees nothing; admin 99 sees all).
type fakeAccess struct{ r *referral.Referral }

func (f fakeAccess) Get(_ context.Context, a referral.Actor, id string) (*referral.Referral, error) {
	if id != f.r.ID {
		return nil, referral.ErrNotFound
	}
	if referral.IsParticipant(a, f.r) || a.ID == outsiderTherapist.ID {
		return f.r, nil // the outsider sees it (open inbox) but is no participant
	}
	return nil, referral.ErrNotFound
}

var (
	doctor            = referral.Actor{ID: 1, Role: models.RoleDoctor}
	physio            = referral.Actor{ID: 10, Role: models.RolePhysioTherapist}
	outsiderTherapist = referral.Actor{ID: 11, Role: models.RolePhysioTherapist}
	nurse             = referral.Actor{ID: 20, Role: models.RoleNurse}
	admin             = referral.Actor{ID: 99, Role: models.RoleStateAdmin}
)

var pdf = []byte("%PDF-1.7\n1 0 obj << >> endobj\ntrailer\n%%EOF\n")

func newSvc(t *testing.T, max int64) (*Service, *fakeRepo, *storage.Local) {
	t.Helper()
	store, err := storage.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	therapist := physio.ID
	r := &referral.Referral{ID: "11111111-1111-4111-8111-111111111111", DoctorID: doctor.ID, TherapistID: &therapist}
	repo := &fakeRepo{rows: map[string]Attachment{}}
	return NewService(repo, store, fakeAccess{r}, max), repo, store
}

const rid = "11111111-1111-4111-8111-111111111111"

func TestUploadAndDownload(t *testing.T) {
	ctx := context.Background()
	s, _, _ := newSvc(t, 1<<20)

	a, err := s.Upload(ctx, doctor, rid, "../../etc/MRI report.pdf", bytes.NewReader(pdf))
	if err != nil {
		t.Fatal(err)
	}
	if a.FileName != "MRI report.pdf" || a.ContentType != "application/pdf" || a.Size != int64(len(pdf)) {
		t.Fatalf("metadata: %+v", a)
	}
	if !strings.HasPrefix(a.StorageKey, "referrals/"+rid+"/") {
		t.Errorf("storage key: %s", a.StorageKey)
	}

	// The therapist (participant) downloads the same bytes
	got, body, err := s.Open(ctx, physio, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(body)
	body.Close()
	if !bytes.Equal(b, pdf) || got.ID != a.ID {
		t.Error("downloaded content differs")
	}

	if items, _ := s.List(ctx, physio, rid); len(items) != 1 {
		t.Errorf("list: %d", len(items))
	}
}

func TestUploadRules(t *testing.T) {
	ctx := context.Background()
	exe := append([]byte("MZ\x90\x00"), make([]byte, 100)...) // Windows executable
	cases := []struct {
		name  string
		actor referral.Actor
		body  []byte
		max   int64
		want  error
	}{
		{"non-participant who can see it", outsiderTherapist, pdf, 1 << 20, ErrForbidden},
		{"cannot see the referral", nurse, pdf, 1 << 20, referral.ErrNotFound},
		{"executable renamed to .pdf", doctor, exe, 1 << 20, ErrUnsupportedType},
		{"empty file", doctor, nil, 1 << 20, ErrUnsupportedType},
		{"too large", doctor, append(pdf, bytes.Repeat([]byte("x"), 100)...), int64(len(pdf)), ErrTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, repo, _ := newSvc(t, tc.max)
			_, err := s.Upload(ctx, tc.actor, rid, "scan.pdf", bytes.NewReader(tc.body))
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
			if len(repo.rows) != 0 {
				t.Error("nothing may be stored on failure")
			}
		})
	}

	t.Run("images and text are accepted", func(t *testing.T) {
		s, _, _ := newSvc(t, 1<<20)
		png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
		for name, body := range map[string][]byte{"x.png": png, "notes.txt": []byte("Session 3: improving")} {
			if _, err := s.Upload(ctx, physio, rid, name, bytes.NewReader(body)); err != nil {
				t.Errorf("%s: %v", name, err)
			}
		}
	})
}

func TestDeleteRules(t *testing.T) {
	ctx := context.Background()
	s, repo, store := newSvc(t, 1<<20)
	a, _ := s.Upload(ctx, doctor, rid, "a.pdf", bytes.NewReader(pdf))

	if err := s.Delete(ctx, physio, a.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("other participant: want forbidden, got %v", err)
	}
	if err := s.Delete(ctx, nurse, a.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("cannot see it: want not found, got %v", err)
	}
	if err := s.Delete(ctx, admin, a.ID); err != nil {
		t.Fatal(err)
	}
	if len(repo.rows) != 0 {
		t.Error("row not deleted")
	}
	if _, err := store.Open(a.StorageKey); !errors.Is(err, storage.ErrNotFound) {
		t.Error("file not deleted")
	}
	if _, _, err := s.Open(ctx, doctor, "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Errorf("invalid id: %v", err)
	}
}

func TestSanitizeFileName(t *testing.T) {
	cases := map[string]string{
		"report.pdf":             "report.pdf",
		`C:\Users\x\scan.png`:    "scan.png",
		"../../etc/passwd":       "passwd",
		"a\x00b\nc.txt":          "abc.txt",
		`we"ird<na>me?.pdf`:      "weirdname.pdf",
		"..":                     "file",
		"   ":                    "file",
		strings.Repeat("é", 300): strings.Repeat("é", 200),
	}
	for in, want := range cases {
		if got := SanitizeFileName(in); got != want {
			t.Errorf("SanitizeFileName(%q) = %q, want %q", in, got, want)
		}
	}
}
