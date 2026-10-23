package auth

import (
	"errors"
	"testing"
	"time"
)

func TestAccountStateCachesAndExpires(t *testing.T) {
	defer func() { AccountChecker, nowFunc = nil, time.Now }()
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	nowFunc = func() time.Time { return now }
	calls, active := 0, true
	AccountChecker = func(uint) (AccountState, error) { calls++; return AccountState{Active: active}, nil }

	if !accountState(1).Active || !accountState(1).Active || calls != 1 {
		t.Fatalf("cached answer: calls %d", calls)
	}
	active = false // the user leaves
	if !accountState(1).Active {
		t.Error("still cached for accountCacheTTL")
	}
	now = now.Add(accountCacheTTL + time.Second)
	if accountState(1).Active {
		t.Error("after the TTL the user must be refused")
	}
	active = true
	ForgetAccount(1)
	if !accountState(1).Active {
		t.Error("ForgetAccount must make a change visible at once")
	}
	AccountChecker = func(uint) (AccountState, error) { return AccountState{}, errors.New("db down") }
	if !accountState(2).Active {
		t.Error("a failed check must not sign people out")
	}
}

func TestAllowedBeforePasswordChange(t *testing.T) {
	for path, want := range map[string]bool{
		"/api/v1/users/getcurrentuser/": true,
		"/api/v1/password_change/":      true,
		"/api/v1/patient/":              false,
		"/api/v1/users/x/":              false,
	} {
		if allowedBeforePasswordChange(path) != want {
			t.Errorf("%s: want %v", path, want)
		}
	}
}
