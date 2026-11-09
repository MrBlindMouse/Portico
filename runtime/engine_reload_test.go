package runtime_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestSetExperienceRejectsStrandedSession(t *testing.T) {
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Scene("gone", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	sess := runtime.NewSessionForTest("s1", identity.Anonymous(), "gone")
	eng := runtime.NewEngine(exp, nil)
	eng.RegisterSession(sess)
	smaller, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	err := eng.SetExperience(smaller)
	if err == nil || !strings.Contains(err.Error(), "gone") {
		t.Fatalf("err=%v", err)
	}
}

func TestSetExperienceRejectsRemovedSpace(t *testing.T) {
	const sp = experience.SpaceID("lobby")
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Space(sp, experience.Space{}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s1", "home")
	h.Engine.RegisterSession(sess)
	if _, err := h.Applier().Apply(sess, exp, []experience.Effect{experience.JoinSpaceEffect(sp)}); err != nil {
		t.Fatal(err)
	}
	smaller, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	err = h.Engine.SetExperience(smaller)
	if err == nil || !strings.Contains(err.Error(), "lobby") {
		t.Fatalf("err=%v", err)
	}
}

func TestRedrawSessions(t *testing.T) {
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "one" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s", "home")
	h.Engine.RegisterSession(sess)
	term := faketerm.New()
	runtime.SetSessionTerminalForTest(sess, term)
	next, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "two" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.SetExperience(next); err != nil {
		t.Fatal(err)
	}
	h.Engine.RedrawSessions()
	if !term.Contains("two") {
		t.Fatalf("frames=%v", term.Frames())
	}
}
