package runtime

import (
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
)

func TestUnregisterSessionDropsRateBucket(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, NewApplier(nil, identity.NewMemoryStore()))
	eng.EnableRateLimit(5, 10)
	sess := newSession("s1", identity.Anonymous(), "home")
	eng.RegisterSession(sess)
	if err := eng.limits.allow(sess.ID()); err != nil {
		t.Fatal(err)
	}
	eng.UnregisterSession(sess.ID())
	eng.limits.mu.Lock()
	defer eng.limits.mu.Unlock()
	if _, ok := eng.limits.perSess[sess.ID()]; ok {
		t.Fatal("rate bucket survived unregister")
	}
}
