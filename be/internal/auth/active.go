package auth

import (
	"sync"
	"time"
)

// ActiveChecker reports whether a user may still use the API (not left,
// not deactivated). Set by the server at start; nil means no check (tests).
var ActiveChecker func(userID uint) (bool, error)

// activeCacheTTL bounds how long a deactivated user keeps access with a
// token issued before: at most this long after the change.
const activeCacheTTL = 30 * time.Second

type activeEntry struct {
	active bool
	until  time.Time
}

var (
	activeMu    sync.Mutex
	activeCache = map[uint]activeEntry{}
	nowFunc     = time.Now
)

// userActive asks ActiveChecker, remembering answers for activeCacheTTL.
// A failed check lets the request through (the database being down fails
// the request anyway) rather than signing everyone out.
func userActive(id uint) bool {
	if ActiveChecker == nil {
		return true
	}
	activeMu.Lock()
	e, ok := activeCache[id]
	activeMu.Unlock()
	if ok && nowFunc().Before(e.until) {
		return e.active
	}
	active, err := ActiveChecker(id)
	if err != nil {
		return true
	}
	activeMu.Lock()
	activeCache[id] = activeEntry{active: active, until: nowFunc().Add(activeCacheTTL)}
	activeMu.Unlock()
	return active
}

// ForgetActive drops the cached answer for a user (call after changing their status).
func ForgetActive(id uint) {
	activeMu.Lock()
	delete(activeCache, id)
	activeMu.Unlock()
}
