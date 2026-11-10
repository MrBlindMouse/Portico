package runtime_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestSetInstallationRejectsStrandedSession(t *testing.T) {
	exp, _ := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Scene("gone", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Build()
	sess := runtime.NewSessionForTest("s1", identity.Anonymous(), "gone")
	eng := runtime.NewEngine(exp, nil)
	eng.RegisterSession(sess)
	smaller, _ := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Build()
	err := eng.SetInstallation(smaller)
	if err == nil || !strings.Contains(err.Error(), "gone") {
		t.Fatalf("err=%v", err)
	}
}

func TestSetInstallationRejectsRemovedSpace(t *testing.T) {
	const sp = components.SpaceID("lobby")
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Space(sp, components.Space{}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s1", "home")
	h.Engine.RegisterSession(sess)
	if _, err := h.Applier().Apply(sess, exp, []components.Effect{components.JoinSpaceEffect(sp)}); err != nil {
		t.Fatal(err)
	}
	smaller, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	err = h.Engine.SetInstallation(smaller)
	if err == nil || !strings.Contains(err.Error(), "lobby") {
		t.Fatalf("err=%v", err)
	}
}

func TestRedrawSessions(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "one" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s", "home")
	h.Engine.RegisterSession(sess)
	term := faketerm.New()
	runtime.SetSessionTerminalForTest(sess, term)
	next, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "two" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.SetInstallation(next); err != nil {
		t.Fatal(err)
	}
	h.Engine.RedrawSessions()
	if !term.Contains("two") {
		t.Fatalf("frames=%v", term.Frames())
	}
}
