package runtime_test

import (
	"errors"
	"github.com/MrBlindMouse/Portico/components"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestSpaceValidator(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Space("s", components.Space{
			ValidateSet: func(key string, old, new any) error {
				if key == "n" {
					return errors.New("deny")
				}
				return nil
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	ap := h.Applier()
	sess := h.Session("t", "home")
	_, err = ap.Apply(sess, exp, []components.Effect{
		components.JoinSpaceEffect("s"),
		components.SpaceSetEffect("s", "n", 1),
	})
	if err == nil {
		t.Fatal("expected validator error")
	}
}
