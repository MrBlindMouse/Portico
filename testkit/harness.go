package testkit

import (
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/runtime"
)

// Register adds sess to the engine registry (required for Submit and many multi-session tests).
func (h *Harness) Register(sess *runtime.Session) {
	h.Engine.RegisterSession(sess)
}

// Do runs one input line through dispatch and apply (same as Engine.HandleInput).
func (h *Harness) Do(sess *runtime.Session, line string) error {
	return h.Engine.HandleInput(sess, line)
}

// MustDo calls Do and fails the test on error.
func (h *Harness) MustDo(t *testing.T, sess *runtime.Session, line string) {
	t.Helper()
	if err := h.Do(sess, line); err != nil {
		t.Fatalf("Do(%q): %v", line, err)
	}
}

// SessionsIn returns registered sessions that have joined space.
func (h *Harness) SessionsIn(space components.SpaceID) []*runtime.Session {
	var out []*runtime.Session
	for _, sess := range h.Engine.Sessions() {
		if sess.InSpace(space) {
			out = append(out, sess)
		}
	}
	return out
}
