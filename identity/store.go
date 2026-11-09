package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store persists per-identity key-value data and nicknames.
type Store interface {
	Get(id Identity, key string) (any, bool, error)
	Set(id Identity, key string, value any) error
	ListKeys(id Identity) ([]string, error)
	ReserveNickname(id Identity, nickname string) error
	NicknameOwner(nickname string) (Identity, bool, error)
	ExportAll() (map[string]map[string]any, error)
	ImportAll(data map[string]map[string]any) error
}

// FileStore stores JSON files under dataDir/identities/<fingerprint>.json and nicknames.json.
type FileStore struct {
	dir string
	mu  sync.Mutex
}

// OpenFileStore creates or opens a file-backed store.
func OpenFileStore(dataDir string) (*FileStore, error) {
	idDir := filepath.Join(dataDir, "identities")
	if err := os.MkdirAll(idDir, 0o750); err != nil {
		return nil, err
	}
	return &FileStore{dir: dataDir}, nil
}

type idFile struct {
	Values map[string]any `json:"values"`
}

func (s *FileStore) identityPath(id Identity) (string, error) {
	fp := id.Fingerprint()
	if fp == "" {
		return "", fmt.Errorf("portico: cannot persist anonymous identity")
	}
	return filepath.Join(s.dir, "identities", fp+".json"), nil
}

func (s *FileStore) loadID(id Identity) (idFile, error) {
	path, err := s.identityPath(id)
	if err != nil {
		return idFile{}, err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return idFile{Values: make(map[string]any)}, nil
	}
	if err != nil {
		return idFile{}, err
	}
	var f idFile
	if err := json.Unmarshal(b, &f); err != nil {
		return idFile{}, err
	}
	if f.Values == nil {
		f.Values = make(map[string]any)
	}
	return f, nil
}

func (s *FileStore) saveID(id Identity, f idFile) error {
	path, err := s.identityPath(id)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o640)
}

func (s *FileStore) Get(id Identity, key string) (any, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.loadID(id)
	if err != nil {
		return nil, false, err
	}
	v, ok := f.Values[key]
	return v, ok, nil
}

func (s *FileStore) Set(id Identity, key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.loadID(id)
	if err != nil {
		return err
	}
	f.Values[key] = value
	return s.saveID(id, f)
}

func (s *FileStore) ListKeys(id Identity) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.loadID(id)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(f.Values))
	for k := range f.Values {
		out = append(out, k)
	}
	return out, nil
}

type nickFile struct {
	ByNickname map[string]string `json:"by_nickname"`
}

func (s *FileStore) nickPath() string {
	return filepath.Join(s.dir, "nicknames.json")
}

func (s *FileStore) loadNicks() (nickFile, error) {
	b, err := os.ReadFile(s.nickPath())
	if errors.Is(err, os.ErrNotExist) {
		return nickFile{ByNickname: make(map[string]string)}, nil
	}
	if err != nil {
		return nickFile{}, err
	}
	var f nickFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nickFile{}, err
	}
	if f.ByNickname == nil {
		f.ByNickname = make(map[string]string)
	}
	return f, nil
}

func (s *FileStore) saveNicks(f nickFile) error {
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.nickPath(), b, 0o640)
}

func (s *FileStore) ReserveNickname(id Identity, nickname string) error {
	if err := ValidateNickname(nickname); err != nil {
		return err
	}
	fp := id.Fingerprint()
	if fp == "" {
		return fmt.Errorf("portico: cannot reserve nickname for anonymous identity")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	nf, err := s.loadNicks()
	if err != nil {
		return err
	}
	if owner, ok := nf.ByNickname[nickname]; ok && owner != fp {
		return fmt.Errorf("portico: nickname %q already taken", nickname)
	}
	nf.ByNickname[nickname] = fp
	return s.saveNicks(nf)
}

func (s *FileStore) NicknameOwner(nickname string) (Identity, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nf, err := s.loadNicks()
	if err != nil {
		return Identity{}, false, err
	}
	fp, ok := nf.ByNickname[nickname]
	if !ok {
		return Identity{}, false, nil
	}
	return Identity{fingerprint: fp}, true, nil
}

func (s *FileStore) ExportAll() (map[string]map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idDir := filepath.Join(s.dir, "identities")
	entries, err := os.ReadDir(idDir)
	if err != nil {
		return nil, err
	}
	out := make(map[string]map[string]any)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		fp := e.Name()[:len(e.Name())-5]
		id := Identity{fingerprint: fp}
		f, err := s.loadID(id)
		if err != nil {
			return nil, err
		}
		out[fp] = f.Values
	}
	return out, nil
}

func (s *FileStore) ImportAll(data map[string]map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for fp, vals := range data {
		id := Identity{fingerprint: fp}
		if err := s.saveID(id, idFile{Values: vals}); err != nil {
			return err
		}
	}
	return nil
}

// MemoryStore is an in-memory Store for tests.
type MemoryStore struct {
	mu      sync.Mutex
	byID    map[string]map[string]any
	nicks   map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byID:  make(map[string]map[string]any),
		nicks: make(map[string]string),
	}
}

func (m *MemoryStore) Get(id Identity, key string) (any, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	vals := m.byID[id.Fingerprint()]
	if vals == nil {
		return nil, false, nil
	}
	v, ok := vals[key]
	return v, ok, nil
}

func (m *MemoryStore) Set(id Identity, key string, value any) error {
	fp := id.Fingerprint()
	if fp == "" {
		return fmt.Errorf("portico: cannot persist anonymous identity")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byID[fp] == nil {
		m.byID[fp] = make(map[string]any)
	}
	m.byID[fp][key] = value
	return nil
}

func (m *MemoryStore) ListKeys(id Identity) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	vals := m.byID[id.Fingerprint()]
	out := make([]string, 0, len(vals))
	for k := range vals {
		out = append(out, k)
	}
	return out, nil
}

func (m *MemoryStore) ReserveNickname(id Identity, nickname string) error {
	if err := ValidateNickname(nickname); err != nil {
		return err
	}
	fp := id.Fingerprint()
	if fp == "" {
		return fmt.Errorf("portico: cannot reserve nickname for anonymous identity")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if owner, ok := m.nicks[nickname]; ok && owner != fp {
		return fmt.Errorf("portico: nickname %q already taken", nickname)
	}
	m.nicks[nickname] = fp
	return nil
}

func (m *MemoryStore) NicknameOwner(nickname string) (Identity, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fp, ok := m.nicks[nickname]
	if !ok {
		return Identity{}, false, nil
	}
	return Identity{fingerprint: fp}, true, nil
}

func (m *MemoryStore) ExportAll() (map[string]map[string]any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]map[string]any, len(m.byID))
	for fp, vals := range m.byID {
		cp := make(map[string]any, len(vals))
		for k, v := range vals {
			cp[k] = v
		}
		out[fp] = cp
	}
	return out, nil
}

func (m *MemoryStore) ImportAll(data map[string]map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for fp, vals := range data {
		cp := make(map[string]any, len(vals))
		for k, v := range vals {
			cp[k] = v
		}
		m.byID[fp] = cp
	}
	return nil
}
