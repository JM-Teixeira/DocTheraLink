package attachment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"unicode"

	"klyniva/internal/auth"
	"klyniva/internal/modules/referral"
	"klyniva/internal/storage"

	"github.com/google/uuid"
)

var (
	// ErrForbidden means the actor may see the referral but not change its files.
	ErrForbidden = errors.New("not allowed to perform this action")
	// ErrTooLarge means the file exceeds the configured limit.
	ErrTooLarge = errors.New("file is too large")
	// ErrUnsupportedType means the file content is not an allowed type.
	ErrUnsupportedType = errors.New("file type not allowed (PDF, JPEG, PNG, WebP or text only)")
)

// allowedTypes are detected from the file content, never from the client.
var allowedTypes = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"text/plain":      true,
}

// ReferralAccess is the part of the referral service this module needs:
// Get returns referral.ErrNotFound when the actor may not see it.
type ReferralAccess interface {
	Get(ctx context.Context, actor referral.Actor, id string) (*referral.Referral, error)
}

// Service holds the attachment rules.
type Service struct {
	repo     Repository
	files    storage.Store
	access   ReferralAccess
	maxBytes int64
}

// NewService builds a Service; maxBytes is the per-file size limit.
func NewService(repo Repository, files storage.Store, access ReferralAccess, maxBytes int64) *Service {
	return &Service{repo: repo, files: files, access: access, maxBytes: maxBytes}
}

// List returns the files of a referral the actor can see.
func (s *Service) List(ctx context.Context, actor referral.Actor, referralID string) ([]Attachment, error) {
	if _, err := s.access.Get(ctx, actor, referralID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListByReferral(ctx, referralID)
	if items == nil {
		items = []Attachment{}
	}
	return items, err
}

// Upload stores a file on a referral. Only participants may upload.
func (s *Service) Upload(ctx context.Context, actor referral.Actor, referralID, fileName string, body io.Reader) (*Attachment, error) {
	r, err := s.access.Get(ctx, actor, referralID)
	if err != nil {
		return nil, err
	}
	if !referral.IsParticipant(actor, r) {
		return nil, ErrForbidden
	}

	// Sniff the real type from the first bytes, then stream the rest.
	head := make([]byte, 512)
	n, err := io.ReadFull(body, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return nil, err
	}
	head = head[:n]
	if n == 0 {
		return nil, ErrUnsupportedType
	}
	contentType, ok := detectType(head)
	if !ok {
		return nil, ErrUnsupportedType
	}

	a := &Attachment{
		ReferralID: r.ID, UploaderID: actor.ID,
		FileName: SanitizeFileName(fileName), ContentType: contentType,
	}
	a.ID = uuid.NewString()
	a.StorageKey = fmt.Sprintf("referrals/%s/%s", r.ID, a.ID)

	// Read at most maxBytes+1 so an oversized upload is detected without
	// storing all of it.
	limited := io.LimitReader(io.MultiReader(bytes.NewReader(head), body), s.maxBytes+1)
	written, err := s.files.Save(a.StorageKey, limited)
	if err != nil {
		return nil, err
	}
	if written > s.maxBytes {
		_ = s.files.Delete(a.StorageKey)
		return nil, ErrTooLarge
	}
	a.Size = written
	if err := s.repo.Create(ctx, a); err != nil {
		_ = s.files.Delete(a.StorageKey)
		return nil, err
	}
	return a, nil
}

// Open returns an attachment and its content if the actor can see the referral.
func (s *Service) Open(ctx context.Context, actor referral.Actor, id string) (*Attachment, io.ReadCloser, error) {
	a, err := s.visible(ctx, actor, id)
	if err != nil {
		return nil, nil, err
	}
	f, err := s.files.Open(a.StorageKey)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, nil, ErrNotFound
	}
	return a, f, err
}

// Delete removes an attachment: its uploader or an admin.
func (s *Service) Delete(ctx context.Context, actor referral.Actor, id string) error {
	a, err := s.visible(ctx, actor, id)
	if err != nil {
		return err
	}
	if a.UploaderID != actor.ID && !auth.IsAdmin(actor.Role) {
		return ErrForbidden
	}
	if err := s.repo.Delete(ctx, a.ID); err != nil {
		return err
	}
	if err := s.files.Delete(a.StorageKey); err != nil {
		log.Printf("attachment: deleted row %s but not file: %v", a.ID, err)
	}
	return nil
}

// visible loads an attachment the actor may see (via its referral).
func (s *Service) visible(ctx context.Context, actor referral.Actor, id string) (*Attachment, error) {
	if uuid.Validate(id) != nil {
		return nil, ErrNotFound
	}
	a, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.access.Get(ctx, actor, a.ReferralID); err != nil {
		if errors.Is(err, referral.ErrNotFound) {
			return nil, ErrNotFound // do not reveal that it exists
		}
		return nil, err
	}
	return a, nil
}

// detectType sniffs the content type and checks it against the allow-list.
func detectType(head []byte) (string, bool) {
	mt, _, err := mime.ParseMediaType(http.DetectContentType(head))
	if err != nil || !allowedTypes[mt] {
		return "", false
	}
	return mt, true
}

// SanitizeFileName keeps a safe display name: base name only, no control or
// path characters, at most 200 characters.
func SanitizeFileName(name string) string {
	name = path.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || strings.ContainsRune(`/\:*?"<>|`, r) {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if r := []rune(name); len(r) > 200 {
		name = string(r[:200])
	}
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	return name
}
