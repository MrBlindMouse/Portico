package runtime_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"io"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestSpaceBroadcast(t *testing.T) {
	const sp = components.SpaceID("lobby")
	exp, err := installation.NewBuilder().
		Root("chat").
		Space(sp, components.Space{}).
		Scene("chat", components.Scene{
			Render: func(ctx components.RenderContext) string {
				n := len(ctx.SpaceList(sp, "msgs"))
				if n == 0 {
					return "empty"
				}
				return "msgs"
			},
			Inputs: map[string]components.ActionID{"send": "send"},
		}).
		Action("send", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{
					components.JoinSpaceEffect(sp),
					components.AppendSpaceListEffect(sp, "msgs", "hi"),
					components.BroadcastEffect(sp),
				}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	s1 := h.Session("1", "chat")
	s2 := h.Session("2", "chat")
	t1 := faketerm.New()
	t2 := faketerm.New()
	h.Engine.RegisterSession(s1)
	h.Engine.RegisterSession(s2)
	runtime.SetSessionTerminalForTest(s1, t1)
	runtime.SetSessionTerminalForTest(s2, t2)
	_, _ = h.Engine.Applier().Apply(s2, exp, []components.Effect{components.JoinSpaceEffect(sp)})
	_ = h.Engine.Render(s2, t2)
	if err := h.Engine.HandleInput(s1, "send"); err != nil {
		t.Fatal(err)
	}
	_ = h.Engine.Render(s1, t1)
	if t2.LastFrame() != "msgs" {
		t.Fatalf("s2=%q want msgs", t2.LastFrame())
	}
}

type editorTerm struct {
	frames   []string
	restores int
}

func (t *editorTerm) WriteFrame(s string) error {
	t.frames = append(t.frames, s)
	return nil
}
func (t *editorTerm) ReadInput() (string, error)      { return "", io.EOF }
func (t *editorTerm) OnResize(func(int, int))         {}
func (t *editorTerm) Close() error                    { return nil }
func (t *editorTerm) SetRedraw(func())                {}
func (t *editorTerm) SetComplete(func(string) string) {}
func (t *editorTerm) RestoreInputLine()               { t.restores++ }

func TestBroadcastRestoresLineEditor(t *testing.T) {
	const sp = components.SpaceID("lobby")
	exp, err := installation.NewBuilder().
		Root("chat").
		Space(sp, components.Space{}).
		Scene("chat", components.Scene{
			Render: func(components.RenderContext) string { return "chat" },
			Inputs: map[string]components.ActionID{"send": "send"},
		}).
		Action("send", components.Action{
			Run: func(components.ActionContext) []components.Effect {
				return []components.Effect{components.BroadcastEffect(sp)}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	s1 := h.Session("1", "chat")
	s2 := h.Session("2", "chat")
	h.Engine.RegisterSession(s1)
	h.Engine.RegisterSession(s2)
	if _, err := h.Engine.Applier().Apply(s1, exp, []components.Effect{components.JoinSpaceEffect(sp)}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Engine.Applier().Apply(s2, exp, []components.Effect{components.JoinSpaceEffect(sp)}); err != nil {
		t.Fatal(err)
	}
	other := &editorTerm{}
	runtime.SetSessionTerminalForTest(s2, other)
	if err := h.Engine.HandleInput(s1, "send"); err != nil {
		t.Fatal(err)
	}
	if other.restores != 1 || len(other.frames) != 1 {
		t.Fatalf("restores=%d frames=%v", other.restores, other.frames)
	}
}
