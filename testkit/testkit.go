// Package testkit provides in-memory engine fixtures for framework and site tests.
package testkit

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
)

// Harness wires an Experience to an in-memory store and Engine (real apply path).
type Harness struct {
	Engine *runtime.Engine
	Store  identity.Store
}

// New returns a harness with identity.NewMemoryStore().
func New(exp *experience.Experience) *Harness {
	store := identity.NewMemoryStore()
	ap := runtime.NewApplier(nil, store)
	return &Harness{Engine: runtime.NewEngine(exp, ap), Store: store}
}

// NewWithStore uses an existing store (for persistence or pre-seeded identity data).
func NewWithStore(exp *experience.Experience, store identity.Store) *Harness {
	ap := runtime.NewApplier(nil, store)
	return &Harness{Engine: runtime.NewEngine(exp, ap), Store: store}
}

// Applier returns the engine's effect applier.
func (h *Harness) Applier() *runtime.Applier {
	return h.Engine.Applier()
}

// Session is an anonymous session at scene.
func (h *Harness) Session(id string, scene experience.SceneID) *runtime.Session {
	return runtime.NewSessionForTest(id, identity.Anonymous(), scene)
}

// SessionWithIdentity binds a specific identity at scene.
func (h *Harness) SessionWithIdentity(id string, idn identity.Identity, scene experience.SceneID) *runtime.Session {
	return runtime.NewSessionForTest(id, idn, scene)
}

// MustBuild builds b or fails the test.
func MustBuild(t *testing.T, b *experience.Builder) *experience.Experience {
	t.Helper()
	exp, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	return exp
}
