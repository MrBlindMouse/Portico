package testkit_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestHarnessRunLoop(t *testing.T) {
	exp := testkit.MustBuild(t, installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"go": "goto-other"},
		}).
		Scene("other", components.Scene{
			Render: func(components.RenderContext) string { return "other" },
		}).
		Action("goto-other", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{components.TransitionEffect("other")}
			},
		}))
	h := testkit.New(exp)
	sess := h.Session("t", "home")
	term := faketerm.New("go")
	_ = h.Engine.RunLoop(sess, term, nil)
	if !term.Contains("other") {
		t.Fatalf("frames=%v", term.Frames())
	}
}
