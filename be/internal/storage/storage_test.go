package storage

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestLocalRoundTrip(t *testing.T) {
	s, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	n, err := s.Save("referrals/abc-123/file-1", strings.NewReader("hello"))
	if err != nil || n != 5 {
		t.Fatalf("save: n=%d err=%v", n, err)
	}
	f, err := s.Open("referrals/abc-123/file-1")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(f)
	f.Close()
	if string(b) != "hello" {
		t.Errorf("content: %q", b)
	}
	if err := s.Delete("referrals/abc-123/file-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Open("referrals/abc-123/file-1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("after delete: want ErrNotFound, got %v", err)
	}
	if err := s.Delete("referrals/abc-123/file-1"); err != nil {
		t.Errorf("deleting a missing file should not fail: %v", err)
	}
}

func TestRejectsUnsafeKeys(t *testing.T) {
	s, _ := NewLocal(t.TempDir())
	for _, key := range []string{"../etc/passwd", "/abs", "a/../../b", "UPPER", "a b", ""} {
		if _, err := s.Save(key, strings.NewReader("x")); err == nil {
			t.Errorf("key %q should be rejected", key)
		}
	}
}
