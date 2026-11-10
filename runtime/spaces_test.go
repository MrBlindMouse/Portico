package runtime_test

import (
	"github.com/MrBlindMouse/Portico/components"
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
