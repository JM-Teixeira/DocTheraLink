package auth

import (
	"errors"
	"testing"
	"time"
)

func TestUserActiveCachesAndExpires(t *testing.T) {
	defer func() { ActiveChecker, nowFunc = nil, time.Now }()
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	nowFunc = func() time.Time { return now }
	calls, active := 0, true
	ActiveChecker = func(uint) (bool, error) { calls++; return active, nil }

	if !userActive(1) || !userActive(1) || calls != 1 {
		t.Fatalf("cached answer: calls %d", calls)
	}
	active = false // the user leaves
	if !userActive(1) {
		t.Error("still cached for activeCacheTTL")
	}
	now = now.Add(activeCacheTTL + time.Second)
	if userActive(1) {
		t.Error("after the TTL the user must be refused")
	}
	ForgetActive(2)
	active = true
	ActiveChecker = func(uint) (bool, error) { return false, errors.New("db down") }
	if !userActive(2) {
		t.Error("a failed check must not sign people out")
	}
}
