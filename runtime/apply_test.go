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
}
