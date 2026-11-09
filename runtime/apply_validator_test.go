package runtime_test

import (
	"errors"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestSpaceValidator(t *testing.T) {
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Space("s", experience.Space{
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
	ap := runtime.NewApplier(nil, identity.NewMemoryStore())
	sess := runtime.NewSessionForTest("t", identity.Anonymous(), "home")
	_, err = ap.Apply(sess, exp, []experience.Effect{
		experience.JoinSpaceEffect("s"),
		experience.SpaceSetEffect("s", "n", 1),
	})
	if err == nil {
		t.Fatal("expected validator error")
	}
}
