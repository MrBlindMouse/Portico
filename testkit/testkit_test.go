package testkit_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestHarnessRunLoop(t *testing.T) {
	exp := testkit.MustBuild(t, experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "home" },
			Inputs: map[string]experience.ActionID{"go": "goto-other"},
		}).
		Scene("other", experience.Scene{
			Render: func(experience.RenderContext) string { return "other" },
		}).
		Action("goto-other", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{experience.TransitionEffect("other")}
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
