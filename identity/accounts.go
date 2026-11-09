package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// AccountRecord is persisted account metadata (no plaintext passwords).
type AccountRecord struct {
	Username           string   `json:"username"`
	PasswordHash       string   `json:"password_hash"` // bcrypt hash string
	Email              string   `json:"email,omitempty"`
	EmailVerified      bool     `json:"email_verified,omitempty"`
	LinkedFingerprints []string `json:"linked_fingerprints,omitempty"`
}

// AccountStore registers and verifies accounts.
type AccountStore interface {
	Register(username, passwordPlain string) error
	Verify(username, passwordPlain string) (bool, error)
	LinkKey(username, fingerprint string) error
	Lookup(username string) (AccountRecord, bool, error)
	LookupByFingerprint(fp string) (AccountRecord, bool, error)
	Export() ([]AccountRecord, error)
	Import(records []AccountRecord) error
	AuditLog() []string
}

// PlaintextAccountStore is a dev-only username/password map.
type PlaintextAccountStore struct {
	mu    sync.RWMutex
	users map[string]string
	audit []string
}

// NewPlaintextAccountStore creates an empty dev store.
func NewPlaintextAccountStore() *PlaintextAccountStore {
	return &PlaintextAccountStore{users: make(map[string]string)}
}

func (a *PlaintextAccountStore) Register(username, passwordPlain string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.users[username]; ok {
		return fmt.Errorf("portico: account %q already exists", username)
	}
	a.users[username] = passwordPlain
	a.audit = append(a.audit, "register:"+username)
	return nil
}

func (a *PlaintextAccountStore) Verify(username, passwordPlain string) (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	want, ok := a.users[username]
	if !ok {
		return false, nil
	}
	return want == passwordPlain, nil
}

func (a *PlaintextAccountStore) LinkKey(username, fingerprint string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.users[username]; !ok {
		return fmt.Errorf("portico: account %q not found", username)
	}
	a.audit = append(a.audit, "link:"+username+":"+fingerprint)
	return nil
}

func (a *PlaintextAccountStore) Lookup(username string) (AccountRecord, bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, ok := a.users[username]
	if !ok {
		return AccountRecord{}, false, nil
	}
	return AccountRecord{Username: username}, true, nil
}

func (a *PlaintextAccountStore) LookupByFingerprint(string) (AccountRecord, bool, error) {
	return AccountRecord{}, false, nil
}

func (a *PlaintextAccountStore) Export() ([]AccountRecord, error) {
	return nil, errors.New("portico: plaintext account export not supported; use file account store")
}

func (a *PlaintextAccountStore) Import([]AccountRecord) error {
	return errors.New("portico: plaintext account import not supported")
}

func (a *PlaintextAccountStore) AuditLog() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	cp := make([]string, len(a.audit))
	copy(cp, a.audit)
	return cp
}

// FileAccountStore persists bcrypt-hashed accounts under dataDir/accounts/.
type FileAccountStore struct {
	dir   string
	mu    sync.RWMutex
	audit []string
}

// OpenFileAccountStore opens or creates the accounts directory.
func OpenFileAccountStore(dataDir string) (*FileAccountStore, error) {
	dir := filepath.Join(dataDir, "accounts")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &FileAccountStore{dir: dir}, nil
}

func (f *FileAccountStore) path(username string) string {
	return filepath.Join(f.dir, username+".json")
}

func (f *FileAccountStore) load(username string) (AccountRecord, error) {
	b, err := os.ReadFile(f.path(username))
	if err != nil {
		return AccountRecord{}, err
	}
	var rec AccountRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return AccountRecord{}, err
	}
	return rec, nil
}

func (f *FileAccountStore) save(rec AccountRecord) error {
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f.path(rec.Username), b, 0o600)
}

func (f *FileAccountStore) Register(username, passwordPlain string) error {
	if username == "" || passwordPlain == "" {
		return fmt.Errorf("portico: username and password required")
	}
	if _, err := f.load(username); err == nil {
		return fmt.Errorf("portico: account %q already exists", username)
	} else if !os.IsNotExist(err) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(passwordPlain), bcryptCost)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.save(AccountRecord{Username: username, PasswordHash: string(hash)}); err != nil {
		return err
	}
	f.audit = append(f.audit, "register:"+username)
	return nil
}

func (f *FileAccountStore) Verify(username, passwordPlain string) (bool, error) {
	rec, err := f.load(username)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(rec.PasswordHash), []byte(passwordPlain))
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (f *FileAccountStore) LinkKey(username, fingerprint string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, err := f.load(username)
	if err != nil {
		return err
	}
	for _, fp := range rec.LinkedFingerprints {
		if fp == fingerprint {
			return nil
		}
	}
	rec.LinkedFingerprints = append(rec.LinkedFingerprints, fingerprint)
	if err := f.save(rec); err != nil {
		return err
	}
	f.audit = append(f.audit, "link:"+username+":"+fingerprint)
	return nil
}

func (f *FileAccountStore) Lookup(username string) (AccountRecord, bool, error) {
	rec, err := f.load(username)
	if os.IsNotExist(err) {
		return AccountRecord{}, false, nil
	}
	if err != nil {
		return AccountRecord{}, false, err
	}
	return rec, true, nil
}

func (f *FileAccountStore) LookupByFingerprint(fp string) (AccountRecord, bool, error) {
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return AccountRecord{}, false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		var rec AccountRecord
		b, err := os.ReadFile(filepath.Join(f.dir, e.Name()))
		if err != nil {
			continue
		}
		if json.Unmarshal(b, &rec) != nil {
			continue
		}
		for _, link := range rec.LinkedFingerprints {
			if link == fp {
				return rec, true, nil
			}
		}
	}
	return AccountRecord{}, false, nil
}

func (f *FileAccountStore) Export() ([]AccountRecord, error) {
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return nil, err
	}
	var out []AccountRecord
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(f.dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var rec AccountRecord
		if err := json.Unmarshal(b, &rec); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

func (f *FileAccountStore) Import(records []AccountRecord) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, rec := range records {
		if rec.Username == "" || rec.PasswordHash == "" {
			return fmt.Errorf("portico: invalid account record for %q", rec.Username)
		}
		if err := f.save(rec); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileAccountStore) AuditLog() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	cp := make([]string, len(f.audit))
	copy(cp, f.audit)
	return cp
}
