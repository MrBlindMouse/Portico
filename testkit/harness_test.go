package testkit_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestHarnessDoAssertSceneAndLocal(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"go": "goto"},
		}).
		Scene("other", components.Scene{
			Render: func(ctx components.RenderContext) string {
				v, _ := ctx.Local("flag")
				return "other:" + v.(string)
			},
		}).
		Action("goto", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{
					components.SetLocalEffect("flag", "yes"),
					components.TransitionEffect("other"),
				}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s", "home")
	h.Register(sess)
	h.MustDo(t, sess, "go")
	h.AssertScene(t, sess, "other")
	h.AssertLocal(t, sess, "flag", "yes")
	h.AssertRenderedContains(t, sess, "other:yes")
}

func TestHarnessAssertPersisted(t *testing.T) {
	idn := identity.ForFingerprint("feedface")
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "" },
			Inputs: map[string]components.ActionID{"save": "save"},
		}).
		Action("save", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{components.PersistEffect("note", "hello")}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.SessionWithIdentity("s", idn, "home")
	h.Register(sess)
	h.MustDo(t, sess, "save")
	h.AssertPersisted(t, idn, "note", "hello")
}

func TestHarnessBroadcastMemberRender(t *testing.T) {
	const sp = components.SpaceID("room")
	exp, err := installation.NewBuilder().
		Root("home").
		Space(sp, components.Space{}).
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string {
				list := ctx.SpaceList(sp, "log")
				if len(list) == 0 {
					return "empty"
				}
				return list[len(list)-1].(string)
			},
			Inputs: map[string]components.ActionID{"post": "post"},
		}).
		Action("post", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				line := ctx.InputLine()
				return []components.Effect{
					components.JoinSpaceEffect(sp),
					components.AppendSpaceListEffect(sp, "log", line),
					components.BroadcastEffect(sp),
				}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	s1 := h.Session("a", "home")
	s2 := h.Session("b", "home")
	h.Register(s1)
	h.Register(s2)
	if _, err := h.Applier().Apply(s1, exp, []components.Effect{components.JoinSpaceEffect(sp)}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Applier().Apply(s2, exp, []components.Effect{components.JoinSpaceEffect(sp)}); err != nil {
		t.Fatal(err)
	}
	h.MustDo(t, s1, "post hi")
	h.AssertMemberRenderedContains(t, sp, "hi")
	if members := h.SessionsIn(sp); len(members) != 2 {
		t.Fatalf("SessionsIn=%d", len(members))
	}
}
