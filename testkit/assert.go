package testkit

import (
	"reflect"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
)

// AssertScene fails when the session's current scene is not want.
func (h *Harness) AssertScene(t *testing.T, sess *runtime.Session, want components.SceneID) {
	t.Helper()
	got := sess.CurrentScene()
	if got != want {
		t.Fatalf("session %q scene: got %q want %q", sess.ID(), got, want)
	}
}

// AssertLocal fails when session-local key does not deep-equal want.
func (h *Harness) AssertLocal(t *testing.T, sess *runtime.Session, key string, want any) {
	t.Helper()
	got, ok := sess.LocalCopy()[key]
	if !ok && want != nil {
		t.Fatalf("session %q local %q: missing, want %v", sess.ID(), key, want)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("session %q local %q: got %v want %v", sess.ID(), key, got, want)
	}
}

// AssertPersisted fails when identity store key does not deep-equal want.
func (h *Harness) AssertPersisted(t *testing.T, idn identity.Identity, key string, want any) {
	t.Helper()
	got, ok, err := h.Store.Get(idn, key)
	if err != nil {
		t.Fatalf("session store get %q: %v", key, err)
	}
	if !ok && want != nil {
		t.Fatalf("identity %q persisted %q: missing, want %v", idn.Fingerprint(), key, want)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("identity %q persisted %q: got %v want %v", idn.Fingerprint(), key, got, want)
	}
}

// AssertRenderedContains renders the session scene and fails when substr is absent.
func (h *Harness) AssertRenderedContains(t *testing.T, sess *runtime.Session, substr string) {
	t.Helper()
	text, err := h.Engine.RenderText(sess)
	if err != nil {
		t.Fatalf("session %q render: %v", sess.ID(), err)
	}
	if !strings.Contains(text, substr) {
		excerpt := text
		if len(excerpt) > 120 {
			excerpt = excerpt[:120] + "..."
		}
		t.Fatalf("session %q render missing %q in:\n%s", sess.ID(), substr, excerpt)
	}
}

// AssertMemberRenderedContains renders each member of space and fails if none contain substr.
func (h *Harness) AssertMemberRenderedContains(t *testing.T, space components.SpaceID, substr string) {
	t.Helper()
	members := h.SessionsIn(space)
	if len(members) == 0 {
		t.Fatalf("space %q has no registered members", space)
	}
	for _, sess := range members {
		text, err := h.Engine.RenderText(sess)
		if err != nil {
			t.Fatalf("session %q render: %v", sess.ID(), err)
		}
		if strings.Contains(text, substr) {
			return
		}
	}
	ids := make([]string, len(members))
	for i, s := range members {
		ids[i] = s.ID()
	}
	t.Fatalf("space %q: no member render contained %q (sessions %v)", space, substr, ids)
}
