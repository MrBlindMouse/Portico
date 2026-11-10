package runtime

import (
	"fmt"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
)

func TestSubmitAllAppendsOnceAndPaints(t *testing.T) {
	const sp = components.SpaceID("room")
	exp, err := installation.NewBuilder().
		Root("home").
		Space(sp, components.Space{}).
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string {
				return fmt.Sprintf("n=%d", len(ctx.SpaceList(sp, "log")))
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, NewApplier(nil, identity.NewMemoryStore()))
	s1 := newSession("s1", identity.Anonymous(), "home")
	s2 := newSession("s2", identity.Anonymous(), "home")
	eng.RegisterSession(s1)
	eng.RegisterSession(s2)
	if _, err := eng.Applier().Apply(s1, exp, []components.Effect{components.JoinSpaceEffect(sp)}); err != nil {
		t.Fatal(err)
	}
	t1 := &slotTerm{}
	t2 := &slotTerm{}
	SetSessionTerminalForTest(s1, t1)
	SetSessionTerminalForTest(s2, t2)

	err = eng.SubmitAll([]components.Effect{
		components.SetLocalEffect("k", 1),
		components.AppendSpaceListEffect(sp, "log", "x"),
		components.BroadcastEffect(sp),
	})
	if err != nil {
		t.Fatal(err)
	}
	if s1.LocalCopy()["k"] != 1 || s2.LocalCopy()["k"] != 1 {
		t.Fatalf("local s1=%v s2=%v", s1.LocalCopy(), s2.LocalCopy())
	}
	if got := eng.spaces.list(sp, "log"); len(got) != 1 {
		t.Fatalf("list %#v", got)
	}
	if len(t1.frames) != 1 || t1.frames[0] != "n=1" {
		t.Fatalf("s1 frames %#v", t1.frames)
	}
	if len(t2.frames) != 1 || t2.frames[0] != "n=1" {
		t.Fatalf("s2 frames %#v", t2.frames)
	}
}

func TestSubmitUnknownSession(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, nil)
	err = eng.Submit("missing", []components.Effect{components.SetLocalEffect("k", 1)})
	if err == nil || !strings.Contains(err.Error(), "no such session") {
		t.Fatalf("err=%v", err)
	}
}

func TestSubmitValidateLocal(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "" },
			ValidateLocal: func(string, any) error {
				return fmt.Errorf("portico: rejected")
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, nil)
	sess := newSession("s", identity.Anonymous(), "home")
	eng.RegisterSession(sess)
	err = eng.Submit(sess.ID(), []components.Effect{components.SetLocalEffect("k", 1)})
	if err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Fatalf("err=%v", err)
	}
	if _, ok := sess.LocalCopy()["k"]; ok {
		t.Fatal("local was written")
	}
}

func TestRepaintScenePaintsOnlyThatScene(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "home" },
		}).
		Scene("away", components.Scene{
			Render: func(components.RenderContext) string { return "away" },
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, nil)
	home := newSession("home-sess", identity.Anonymous(), "home")
	away := newSession("away-sess", identity.Anonymous(), "home")
	eng.RegisterSession(home)
	eng.RegisterSession(away)
	if _, err := eng.Applier().Apply(away, exp, []components.Effect{components.TransitionEffect("away")}); err != nil {
		t.Fatal(err)
	}
	homeTerm := &slotTerm{}
	awayTerm := &slotTerm{}
	SetSessionTerminalForTest(home, homeTerm)
	SetSessionTerminalForTest(away, awayTerm)
	if err := eng.RepaintScene("home"); err != nil {
		t.Fatal(err)
	}
	if len(homeTerm.frames) != 1 || homeTerm.frames[0] != "home" {
		t.Fatalf("home frames %#v", homeTerm.frames)
	}
	if len(awayTerm.frames) != 0 {
		t.Fatalf("away frames %#v", awayTerm.frames)
	}
	if err := eng.RepaintScene("missing"); err != nil {
		t.Fatal(err)
	}
	if len(homeTerm.frames) != 1 {
		t.Fatalf("missing scene painted %#v", homeTerm.frames)
	}
}
