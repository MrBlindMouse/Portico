package runtime_test

import (
	"errors"
	"github.com/MrBlindMouse/Portico/components"
	"io"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestLoopTransition(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"go": "goto-other"},
		}).
		Scene("other", components.Scene{
			Render: func(ctx components.RenderContext) string { return "other" },
		}).
		Action("goto-other", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{components.TransitionEffect("other")}
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
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"go": "goto-other"},
		}).
		Scene("other", components.Scene{
			Render: func(ctx components.RenderContext) string { return "other" },
		}).
		Action("goto-other", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{components.TransitionEffect("other")}
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
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"go": "goto-other"},
		}).
		Scene("other", components.Scene{
			Render: func(ctx components.RenderContext) string { return "other" },
		}).
		Action("goto-other", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{components.TransitionEffect("other")}
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

func TestLoopOnLineOnEmpty(t *testing.T) {
	var (
		lastLine string
		emptyN   int
	)
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render:  func(ctx components.RenderContext) string { return "home" },
			OnLine:  "submit",
			OnEmpty: "cancel",
		}).
		Action("submit", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				lastLine = ctx.InputLine()
				return nil
			},
		}).
		Action("cancel", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				emptyN++
				return nil
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("t", "home")
	term := faketerm.New("", "typed blob")
	if err := h.Engine.RunLoop(sess, term, nil); err == nil || !strings.Contains(err.Error(), "no more inputs") {
		t.Logf("RunLoop err=%v", err)
	}
	if emptyN != 1 {
		t.Fatalf("emptyN=%d", emptyN)
	}
	if lastLine != "typed blob" {
		t.Fatalf("lastLine=%q", lastLine)
	}
}

type rejectTerm struct {
	frames  []string
	rejects []string
	inputs  []string
	idx     int
}

func (t *rejectTerm) WriteFrame(s string) error {
	t.frames = append(t.frames, s)
	return nil
}
func (t *rejectTerm) ReadInput() (string, error) {
	if t.idx >= len(t.inputs) {
		return "", io.EOF
	}
	v := t.inputs[t.idx]
	t.idx++
	return v, nil
}
func (t *rejectTerm) OnResize(func(int, int))      {}
func (t *rejectTerm) Close() error                 { return nil }
func (t *rejectTerm) BeginInputLine() error        { return nil }
func (t *rejectTerm) CommitInputLine(string) error { return nil }
func (t *rejectTerm) RejectInput(msg string) error {
	t.rejects = append(t.rejects, msg)
	return nil
}

func TestLoopApplyErrorStaysOnPrompt(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"bad": "bad"},
			ValidateLocal: func(string, any) error {
				return errors.New("rejected")
			},
		}).
		Action("bad", components.Action{
			Run: func(components.ActionContext) []components.Effect {
				return []components.Effect{components.SetLocalEffect("k", "v")}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("t", "home")
	term := &rejectTerm{inputs: []string{"bad"}}
	if err := h.Engine.RunLoop(sess, term, nil); err != nil && !strings.Contains(err.Error(), "EOF") {
		t.Fatal(err)
	}
	if len(term.rejects) != 1 || !strings.Contains(term.rejects[0], "rejected") {
		t.Fatalf("rejects=%v", term.rejects)
	}
	if len(term.frames) == 0 || !strings.Contains(term.frames[len(term.frames)-1], "home") {
		t.Fatalf("frames=%v", term.frames)
	}
	for _, f := range term.frames {
		if strings.Contains(f, "rejected") {
			t.Fatalf("error replaced the scene frame: %q", f)
		}
	}
}
