package runtime_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestRateLimitAppliesOnlyAfterResolve(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"go": "goto"},
		}).
		Action("goto", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect { return nil },
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	h.Engine.EnableRateLimit(1, 1)
	sess := h.Session("t", "home")

	for i := 0; i < 20; i++ {
		_ = h.Engine.HandleInput(sess, "not-a-command")
	}
	if err := h.Engine.HandleInput(sess, "go"); err != nil {
		if strings.Contains(err.Error(), "rate limit") {
			t.Fatal("unknown input should not consume rate limit tokens")
		}
		t.Fatal(err)
	}
}
