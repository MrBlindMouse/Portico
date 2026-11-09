package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestSessionStartEffect(t *testing.T) {
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		OnSessionStart(experience.SetLocalEffect("boot", true)).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	ap := h.Applier()
	sess := h.Session("t", "home")
	_, err = ap.Apply(sess, exp, exp.OnSessionStart())
	if err != nil {
		t.Fatal(err)
	}
	v, ok := sess.LocalCopy()["boot"]
	if !ok || v != true {
		t.Fatalf("boot local: %v ok=%v", v, ok)
	}
}
