package runtime

import (
	"crypto/subtle"
	"sync"
)

// AccountStore holds username → password for development account auth (phase 8).
// ponytail: plaintext dev store; production sites should use password hashes via extension.
type AccountStore struct {
	mu    sync.RWMutex
	users map[string]string
	audit []string
}

// NewAccountStore creates an empty account store.
func NewAccountStore() *AccountStore {
	return &AccountStore{users: make(map[string]string)}
}

// Register adds or updates a user.
func (a *AccountStore) Register(user, password string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.users[user] = password
	a.audit = append(a.audit, "register:"+user)
}

// Verify checks credentials.
func (a *AccountStore) Verify(user, password string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	want, ok := a.users[user]
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(want), []byte(password)) == 1
}

// AuditLog returns privileged operation audit lines.
func (a *AccountStore) AuditLog() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	cp := make([]string, len(a.audit))
	copy(cp, a.audit)
	return cp
}
