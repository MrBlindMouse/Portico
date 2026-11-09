package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
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
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	s1 := runtime.NewSessionForTest("1", identity.Anonymous(), "chat")
	s2 := runtime.NewSessionForTest("2", identity.Anonymous(), "chat")
	t1 := faketerm.New()
	t2 := faketerm.New()
	eng.RegisterSession(s1)
	eng.RegisterSession(s2)
	runtime.SetSessionTerminalForTest(s1, t1)
	runtime.SetSessionTerminalForTest(s2, t2)
	_, _ = eng.Applier().Apply(s2, exp, []experience.Effect{experience.JoinSpaceEffect(sp)})
	_ = eng.Render(s2, t2)
	if err := eng.HandleInput(s1, "send"); err != nil {
		t.Fatal(err)
	}
	_ = eng.Render(s1, t1)
	if t2.LastFrame() != "msgs" {
		t.Fatalf("s2=%q want msgs", t2.LastFrame())
	}
}
