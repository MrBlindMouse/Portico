package runtime_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestLoopTransition(t *testing.T) {
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string { return "home" },
			Inputs: map[string]experience.ActionID{"go": "goto-other"},
		}).
		Scene("other", experience.Scene{
			Render: func(ctx experience.RenderContext) string { return "other" },
		}).
		Action("goto-other", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{experience.TransitionEffect("other")}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("t", identity.Anonymous(), "home")
	term := faketerm.New("go")
	if err := eng.RunLoop(sess, term, nil); err == nil || !strings.Contains(err.Error(), "no more inputs") {
		// RunLoop reads until error after last input
	}
	if !term.Contains("other") {
		t.Fatalf("frames=%v", term.Frames())
	}
}
