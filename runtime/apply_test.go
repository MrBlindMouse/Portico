package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestApplyTransitionAndLocal(t *testing.T) {
	exp, err := experience.NewBuilder().
		Root("a").
		Scene("a", experience.Scene{Render: func(experience.RenderContext) string { return "a" }}).
		Scene("b", experience.Scene{Render: func(experience.RenderContext) string { return "b" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	ap := h.Applier()
	sess := h.Session("1", "a")
	res, err := ap.Apply(sess, exp, []experience.Effect{
		experience.SetLocalEffect("x", 1),
		experience.TransitionEffect("b"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Transitioned || sess.CurrentScene() != "b" {
		t.Fatalf("scene=%s transitioned=%v", sess.CurrentScene(), res.Transitioned)
	}
}

func TestApplySpaceOnJoin(t *testing.T) {
	const sp = experience.SpaceID("lobby")
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Space(sp, experience.Space{
			OnJoin: []experience.Effect{experience.SetLocalEffect("joined", true)},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	ap := h.Applier()
	sess := h.Session("1", "home")
	_, err = ap.Apply(sess, exp, []experience.Effect{experience.JoinSpaceEffect(sp)})
	if err != nil {
		t.Fatal(err)
	}
	local := sess.LocalCopy()
	if local["joined"] != true {
		t.Fatalf("local=%v", local)
	}
	_, err = ap.Apply(sess, exp, []experience.Effect{
		experience.SetLocalEffect("joined", false),
		experience.JoinSpaceEffect(sp),
	})
	if err != nil {
		t.Fatal(err)
	}
	if sess.LocalCopy()["joined"] != false {
		t.Fatal("second join re-ran OnJoin")
	}
}

func TestApplyJoinCycleTerminates(t *testing.T) {
	render := func(experience.RenderContext) string { return "" }
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: render}).
		Space("a", experience.Space{OnJoin: []experience.Effect{experience.JoinSpaceEffect("b")}}).
		Space("b", experience.Space{OnJoin: []experience.Effect{experience.JoinSpaceEffect("a")}}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("1", "home")
	if _, err := h.Applier().Apply(sess, exp, []experience.Effect{experience.JoinSpaceEffect("a")}); err != nil {
		t.Fatal(err)
	}
	if !sess.InSpace("a") || !sess.InSpace("b") {
		t.Fatalf("spaces=%v", sess.Spaces())
	}
}

func TestApplyLeaveUnknownAndRepeat(t *testing.T) {
	render := func(experience.RenderContext) string { return "" }
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: render}).
		Space("s", experience.Space{
			OnLeave: []experience.Effect{experience.LeaveSpaceEffect("s"), experience.SetLocalEffect("left", true)},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("1", "home")
	ap := h.Applier()
	if _, err := ap.Apply(sess, exp, []experience.Effect{experience.LeaveSpaceEffect("s")}); err != nil {
		t.Fatal(err)
	}
	if sess.LocalCopy()["left"] == true {
		t.Fatal("leave ran OnLeave for a session that was not a member")
	}
	if _, err := ap.Apply(sess, exp, []experience.Effect{experience.JoinSpaceEffect("s"), experience.LeaveSpaceEffect("s")}); err != nil {
		t.Fatal(err)
	}
	if sess.InSpace("s") || sess.LocalCopy()["left"] != true {
		t.Fatalf("in=%v local=%v", sess.InSpace("s"), sess.LocalCopy())
	}
}

func TestApplyAppendAndCASRequireSpace(t *testing.T) {
	render := func(experience.RenderContext) string { return "" }
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: render}).
		Space("s", experience.Space{}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("1", "home")
	ap := h.Applier()
	if _, err := ap.Apply(sess, exp, []experience.Effect{experience.AppendSpaceListEffect("missing", "k", "v")}); err == nil {
		t.Fatal("append to unknown space")
	}
	if _, err := ap.Apply(sess, exp, []experience.Effect{experience.SpaceCAS{Space: "missing", Key: "k", Value: "v"}}); err == nil {
		t.Fatal("cas on unknown space")
	}
	if _, err := ap.Apply(sess, exp, []experience.Effect{
		experience.SpaceSetEffect("s", "ready", "no"),
		experience.SpaceCASEffect("s", "ready", "no", "yes"),
	}); err != nil {
		t.Fatal(err)
	}
}
