package runtime

import (
	"github.com/MrBlindMouse/Portico/components"
	"sync"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/render"
)

// Session is one SSH connection's runtime state.
type Session struct {
	id             string
	identity       identity.Identity
	current        components.SceneID
	local          map[string]any
	spaces         map[components.SpaceID]struct{}
	lastRender     string
	lastRenderSize int
	termW          int
	termH          int
	slot           *render.Input
	commands       []string // successful lines; cleared on scene change unless kept

	mu sync.RWMutex
}

// commandCap bounds the session transcript.
// ponytail: fixed slice; raise it if a scene shows a longer log.
const commandCap = 32

func newSession(id string, idn identity.Identity, root components.SceneID) *Session {
	return &Session{
		id:       id,
		identity: idn,
		current:  root,
		local:    make(map[string]any),
		spaces:   make(map[components.SpaceID]struct{}),
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
func (s *Session) CurrentScene() components.SceneID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// InSpace reports membership.
func (s *Session) InSpace(space components.SpaceID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.spaces[space]
	return ok
}

// Spaces returns joined space ids.
func (s *Session) Spaces() []components.SpaceID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]components.SpaceID, 0, len(s.spaces))
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

func (s *Session) setSlot(in *render.Input) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if in == nil {
		s.slot = nil
		return
	}
	cp := *in
	s.slot = &cp
}

func (s *Session) inputSlot() (render.Input, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.slot == nil {
		return render.Input{}, false
	}
	return *s.slot, true
}

func (s *Session) setSize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.termW = width
	s.termH = height
}

func (s *Session) size() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, h := s.termW, s.termH
	if w <= 0 {
		w = render.DefaultWidth
	}
	if h <= 0 {
		h = render.DefaultHeight
	}
	return w, h
}

func (s *Session) noteCommand(line string) {
	if line == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commands = append(s.commands, line)
	if len(s.commands) > commandCap {
		s.commands = append([]string(nil), s.commands[len(s.commands)-commandCap:]...)
	}
}

func (s *Session) clearCommands() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commands = nil
}

// RecentCommands returns the last n successful lines, oldest first.
// n <= 0 returns the whole list. The caller may keep the slice.
func (s *Session) RecentCommands(n int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.commands) == 0 {
		return nil
	}
	src := s.commands
	if n > 0 && n < len(src) {
		src = src[len(src)-n:]
	}
	out := make([]string, len(src))
	copy(out, src)
	return out
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
