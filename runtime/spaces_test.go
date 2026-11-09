package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestSpaceBroadcast(t *testing.T) {
	const sp = experience.SpaceID("lobby")
	exp, err := experience.NewBuilder().
		Root("chat").
		Space(sp, experience.Space{}).
		Scene("chat", experience.Scene{
			Render: func(ctx experience.RenderContext) string {
				n := len(ctx.SpaceList(sp, "msgs"))
				if n == 0 {
					return "empty"
				}
				return "msgs"
			},
			Inputs: map[string]experience.ActionID{"send": "send"},
		}).
		Action("send", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{
					experience.JoinSpaceEffect(sp),
					experience.AppendSpaceListEffect(sp, "msgs", "hi"),
					experience.BroadcastEffect(sp),
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
	_, _ = h.Engine.Applier().Apply(s2, exp, []experience.Effect{experience.JoinSpaceEffect(sp)})
	_ = h.Engine.Render(s2, t2)
	if err := h.Engine.HandleInput(s1, "send"); err != nil {
		t.Fatal(err)
	}
	_ = h.Engine.Render(s1, t1)
	if t2.LastFrame() != "msgs" {
		t.Fatalf("s2=%q want msgs", t2.LastFrame())
	}
}
