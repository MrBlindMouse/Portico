package runtime

import (
	"reflect"
	"sync"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
)

type spaceState struct {
	mu      sync.RWMutex
	members map[string]*Session
	values  map[string]any
	lists   map[string][]any
}

type spaceRegistry struct {
	mu     sync.RWMutex
	spaces map[experience.SpaceID]*spaceState
}

func newSpaceRegistry() *spaceRegistry {
	return &spaceRegistry{spaces: make(map[experience.SpaceID]*spaceState)}
}

func (r *spaceRegistry) get(space experience.SpaceID) *spaceState {
	r.mu.Lock()
	defer r.mu.Unlock()
	st, ok := r.spaces[space]
	if !ok {
		st = &spaceState{
			members: make(map[string]*Session),
			values:  make(map[string]any),
			lists:   make(map[string][]any),
		}
		r.spaces[space] = st
	}
	return st
}

func (r *spaceRegistry) join(space experience.SpaceID, sess *Session) {
	st := r.get(space)
	st.mu.Lock()
	defer st.mu.Unlock()
	st.members[sess.ID()] = sess
}

func (r *spaceRegistry) leave(space experience.SpaceID, sess *Session) {
	st := r.get(space)
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.members, sess.ID())
}

func (r *spaceRegistry) members(space experience.SpaceID) []*Session {
	st := r.get(space)
	st.mu.RLock()
	defer st.mu.RUnlock()
	out := make([]*Session, 0, len(st.members))
	for _, s := range st.members {
		out = append(out, s)
	}
	return out
}

func (r *spaceRegistry) memberIdentities(space experience.SpaceID) []identity.Identity {
	sessions := r.members(space)
	out := make([]identity.Identity, len(sessions))
	for i, s := range sessions {
		out[i] = s.Identity()
	}
	return out
}

func (r *spaceRegistry) getValue(space experience.SpaceID, key string) (any, bool) {
	st := r.get(space)
	st.mu.RLock()
	defer st.mu.RUnlock()
	v, ok := st.values[key]
	return v, ok
}

func (r *spaceRegistry) set(space experience.SpaceID, key string, value any) {
	st := r.get(space)
	st.mu.Lock()
	defer st.mu.Unlock()
	st.values[key] = value
}

func (r *spaceRegistry) cas(space experience.SpaceID, key string, expect, value any) bool {
	st := r.get(space)
	st.mu.Lock()
	defer st.mu.Unlock()
	cur, ok := st.values[key]
	if !ok && expect == nil {
		st.values[key] = value
		return true
	}
	if ok && reflect.DeepEqual(cur, expect) {
		st.values[key] = value
		return true
	}
	return false
}

func (r *spaceRegistry) appendList(space experience.SpaceID, key string, value any) {
	st := r.get(space)
	st.mu.Lock()
	defer st.mu.Unlock()
	st.lists[key] = append(st.lists[key], value)
}

func (r *spaceRegistry) list(space experience.SpaceID, key string) []any {
	st := r.get(space)
	st.mu.RLock()
	defer st.mu.RUnlock()
	cp := make([]any, len(st.lists[key]))
	copy(cp, st.lists[key])
	return cp
}

func (r *spaceRegistry) snapshot() map[string]SpaceSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]SpaceSnapshot)
	for id, st := range r.spaces {
		st.mu.RLock()
		snap := SpaceSnapshot{
			MemberCount: len(st.members),
			Keys:        make([]string, 0, len(st.values)),
		}
		for k := range st.values {
			snap.Keys = append(snap.Keys, k)
		}
		st.mu.RUnlock()
		out[string(id)] = snap
	}
	return out
}

// SpaceSnapshot is a read-only space summary for inspection.
type SpaceSnapshot struct {
	MemberCount int      `json:"member_count"`
	Keys        []string `json:"keys"`
}
