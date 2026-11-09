package runtime_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestRateLimitAppliesOnlyAfterResolve(t *testing.T) {
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string { return "home" },
			Inputs: map[string]experience.ActionID{"go": "goto"},
		}).
		Action("goto", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect { return nil },
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
