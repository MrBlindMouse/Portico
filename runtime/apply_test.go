package runtime_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestApplyTransitionAndLocal(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("a").
		Scene("a", components.Scene{Render: func(components.RenderContext) string { return "a" }}).
		Scene("b", components.Scene{Render: func(components.RenderContext) string { return "b" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	ap := h.Applier()
	sess := h.Session("1", "a")
	res, err := ap.Apply(sess, exp, []components.Effect{
		components.SetLocalEffect("x", 1),
		components.TransitionEffect("b"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Transitioned || sess.CurrentScene() != "b" {
		t.Fatalf("scene=%s transitioned=%v", sess.CurrentScene(), res.Transitioned)
	}
}

func TestApplySpaceOnJoin(t *testing.T) {
	const sp = components.SpaceID("lobby")
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Space(sp, components.Space{
			OnJoin: []components.Effect{components.SetLocalEffect("joined", true)},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	ap := h.Applier()
	sess := h.Session("1", "home")
	_, err = ap.Apply(sess, exp, []components.Effect{components.JoinSpaceEffect(sp)})
	if err != nil {
		t.Fatal(err)
	}
	local := sess.LocalCopy()
	if local["joined"] != true {
		t.Fatalf("local=%v", local)
	}
	_, err = ap.Apply(sess, exp, []components.Effect{
		components.SetLocalEffect("joined", false),
		components.JoinSpaceEffect(sp),
	})
	if err != nil {
		t.Fatal(err)
	}
	if sess.LocalCopy()["joined"] != false {
		t.Fatal("second join re-ran OnJoin")
	}
}

func TestApplyJoinCycleTerminates(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: render}).
		Space("a", components.Space{OnJoin: []components.Effect{components.JoinSpaceEffect("b")}}).
		Space("b", components.Space{OnJoin: []components.Effect{components.JoinSpaceEffect("a")}}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("1", "home")
	if _, err := h.Applier().Apply(sess, exp, []components.Effect{components.JoinSpaceEffect("a")}); err != nil {
		t.Fatal(err)
	}
	if !sess.InSpace("a") || !sess.InSpace("b") {
		t.Fatalf("spaces=%v", sess.Spaces())
	}
}

func TestApplyLeaveUnknownAndRepeat(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: render}).
		Space("s", components.Space{
			OnLeave: []components.Effect{components.LeaveSpaceEffect("s"), components.SetLocalEffect("left", true)},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("1", "home")
	ap := h.Applier()
	if _, err := ap.Apply(sess, exp, []components.Effect{components.LeaveSpaceEffect("s")}); err != nil {
		t.Fatal(err)
	}
	if sess.LocalCopy()["left"] == true {
		t.Fatal("leave ran OnLeave for a session that was not a member")
	}
	if _, err := ap.Apply(sess, exp, []components.Effect{components.JoinSpaceEffect("s"), components.LeaveSpaceEffect("s")}); err != nil {
		t.Fatal(err)
	}
	if sess.InSpace("s") || sess.LocalCopy()["left"] != true {
		t.Fatalf("in=%v local=%v", sess.InSpace("s"), sess.LocalCopy())
	}
}

func TestApplyAppendAndCASRequireSpace(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: render}).
		Space("s", components.Space{}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("1", "home")
	ap := h.Applier()
	if _, err := ap.Apply(sess, exp, []components.Effect{components.AppendSpaceListEffect("missing", "k", "v")}); err == nil {
		t.Fatal("append to unknown space")
	}
	if _, err := ap.Apply(sess, exp, []components.Effect{components.SpaceCAS{Space: "missing", Key: "k", Value: "v"}}); err == nil {
		t.Fatal("cas on unknown space")
	}
	if _, err := ap.Apply(sess, exp, []components.Effect{
		components.SpaceSetEffect("s", "ready", "no"),
		components.SpaceCASEffect("s", "ready", "no", "yes"),
	}); err != nil {
		t.Fatal(err)
	}
}
