package runtime_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/testkit"
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
	h := testkit.New(exp)
	sess := h.Session("t", "home")
	term := faketerm.New("go")
	if err := h.Engine.RunLoop(sess, term, nil); err == nil || !strings.Contains(err.Error(), "no more inputs") {
		// RunLoop reads until error after last input
	}
	if !term.Contains("other") {
		t.Fatalf("frames=%v", term.Frames())
	}
}

func TestLoopIgnoresIneligibleAndProtocol(t *testing.T) {
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
	h := testkit.New(exp)
	sess := h.Session("t", "home")
	term := faketerm.New("OK", "KITTY_DATA_START", "go")
	if err := h.Engine.RunLoop(sess, term, nil); err == nil || !strings.Contains(err.Error(), "no more inputs") {
		t.Logf("RunLoop err=%v", err)
	}
	if !term.Contains("other") {
		t.Fatalf("expected transition; frames=%v", term.Frames())
	}
	for _, f := range term.Frames() {
		if strings.Contains(f, "no action for input") || strings.Contains(f, "Error:") {
			t.Fatalf("ineligible/protocol lines should not dispatch: %q", f)
		}
	}
}

func TestLoopIgnoresTerminalProtocolLines(t *testing.T) {
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
	h := testkit.New(exp)
	h.Engine.EnableRateLimit(1, 2) // burst 2: protocol spam would exceed without filter
	sess := h.Session("t", "home")
	noise := []string{
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"KITTY_DATA_END",
		"go",
	}
	term := faketerm.New(noise...)
	if err := h.Engine.RunLoop(sess, term, nil); err == nil || !strings.Contains(err.Error(), "no more inputs") {
		t.Logf("RunLoop err=%v", err)
	}
	if !term.Contains("other") {
		t.Fatalf("expected transition; frames=%v", term.Frames())
	}
	for _, f := range term.Frames() {
		if strings.Contains(f, "rate limit exceeded") {
			t.Fatalf("protocol noise tripped rate limit: %q", f)
		}
	}
}
