package experience_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
)

func TestBuildUndefinedAction(t *testing.T) {
	_, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "hi" },
			Inputs: map[string]experience.ActionID{"x": "missing"},
		}).
		Build()
	if err == nil || !strings.Contains(err.Error(), "undefined action") {
		t.Fatalf("got err=%v", err)
	}
}

func TestBuildRequiresRoot(t *testing.T) {
	_, err := experience.NewBuilder().
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "hi" },
		}).
		Build()
	if err == nil || !strings.Contains(err.Error(), "root scene") {
		t.Fatalf("got err=%v", err)
	}
}
