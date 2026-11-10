package runtime_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestSessionStartEffect(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		OnSessionStart(components.SetLocalEffect("boot", true)).
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
