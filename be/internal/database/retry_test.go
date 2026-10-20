package database

import (
	"errors"
	"testing"
)

func TestRetry(t *testing.T) {
	t.Run("succeeds after transient failures", func(t *testing.T) {
		calls := 0
		err := retry(5, 0, func() error {
			calls++
			if calls < 3 {
				return errors.New("not ready")
			}
			return nil
		})
		if err != nil || calls != 3 {
			t.Fatalf("err=%v calls=%d, want nil and 3", err, calls)
		}
	})

	t.Run("returns last error when attempts run out", func(t *testing.T) {
		calls := 0
		want := errors.New("down")
		err := retry(4, 0, func() error { calls++; return want })
		if !errors.Is(err, want) || calls != 4 {
			t.Fatalf("err=%v calls=%d, want %v and 4", err, calls, want)
		}
	})
}
