package auth

import (
	"strings"
	"sync"
	"time"
)

// AccountState is what the API checks on every request besides the token.
type AccountState struct {
	Active             bool // false: left or deactivated, the token no longer works
	MustChangePassword bool // first sign-in: only the password change is allowed
}

// AccountChecker reads a user's AccountState. Set by the server at start;
// nil means no check (tests).
var AccountChecker func(userID uint) (AccountState, error)

// accountCacheTTL bounds how long a change takes to reach a token issued
// before it: at most this long (ForgetAccount makes it immediate).
const accountCacheTTL = 30 * time.Second

type accountEntry struct {
	state AccountState
	until time.Time
}

var (
	accountMu    sync.Mutex
	accountCache = map[uint]accountEntry{}
	nowFunc      = time.Now
)

// accountState asks AccountChecker, remembering answers for accountCacheTTL.
// A failed check lets the request through (the database being down fails
// the request anyway) rather than signing everyone out.
func accountState(id uint) AccountState {
	if AccountChecker == nil {
		return AccountState{Active: true}
	}
	accountMu.Lock()
	e, ok := accountCache[id]
	accountMu.Unlock()
	if ok && nowFunc().Before(e.until) {
		return e.state
	}
	state, err := AccountChecker(id)
	if err != nil {
		return AccountState{Active: true}
	}
	accountMu.Lock()
	accountCache[id] = accountEntry{state: state, until: nowFunc().Add(accountCacheTTL)}
	accountMu.Unlock()
	return state
}

// ForgetAccount drops the cached state of a user (call after changing their
// status or password).
func ForgetAccount(id uint) {
	accountMu.Lock()
	delete(accountCache, id)
	accountMu.Unlock()
}

// allowedBeforePasswordChange: what a person who must change their password
// may still call (who am I, change it).
func allowedBeforePasswordChange(path string) bool {
	return strings.HasSuffix(path, "/users/getcurrentuser/") || strings.HasSuffix(path, "/password_change/")
}
