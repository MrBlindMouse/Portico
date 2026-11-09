package runtime

import (
	"sync"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
)

// Session is one SSH connection's runtime state.
type Session struct {
	id       string
	identity identity.Identity
	current  experience.SceneID
	local    map[string]any
	spaces   map[experience.SpaceID]struct{}
	lastRender     string
	lastRenderSize int

	mu sync.RWMutex
}

func newSession(id string, idn identity.Identity, root experience.SceneID) *Session {
	return &Session{
		id:       id,
		identity: idn,
		current:  root,
		local:    make(map[string]any),
		spaces:   make(map[experience.SpaceID]struct{}),
	}
}

// ID returns the session identifier.
func (s *Session) ID() string { return s.id }

// Identity returns the visitor identity.
func (s *Session) Identity() identity.Identity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.identity
}

// CurrentScene returns the active scene id.
func (s *Session) CurrentScene() experience.SceneID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// InSpace reports membership.
func (s *Session) InSpace(space experience.SpaceID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.spaces[space]
	return ok
}

// Spaces returns joined space ids.
func (s *Session) Spaces() []experience.SpaceID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]experience.SpaceID, 0, len(s.spaces))
	for id := range s.spaces {
		out = append(out, id)
	}
	return out
}

// LocalCopy returns a snapshot of session-local state.
func (s *Session) LocalCopy() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]any, len(s.local))
	for k, v := range s.local {
		out[k] = v
	}
	return out
}

func (s *Session) setIdentity(idn identity.Identity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.identity = idn
}

// SetIdentityForTest updates identity in tests.
func (s *Session) SetIdentityForTest(idn identity.Identity) {
	s.setIdentity(idn)
}

func (s *Session) setLastRender(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastRender = text
	s.lastRenderSize = len(text)
}

// LastRender returns the most recent rendered frame and its byte size.
func (s *Session) LastRender() (text string, size int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRender, s.lastRenderSize
}
