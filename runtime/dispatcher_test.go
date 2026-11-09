package runtime_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestResolveInputPrefixCommand(t *testing.T) {
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "" },
			Inputs: map[string]experience.ActionID{"go": "go"},
		}).
		Action("go", experience.Action{Run: func(experience.ActionContext) []experience.Effect { return nil }}).
		Build()
	aid, err := runtime.ResolveInput(exp, "home", "go elsewhere")
	if err != nil || aid != "go" {
		t.Fatalf("aid=%q err=%v", aid, err)
	}
}

func TestResolveInputMissing(t *testing.T) {
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "" },
		}).
		Build()
	_, err := runtime.ResolveInput(exp, "home", "nope")
	if err == nil || !strings.Contains(err.Error(), "no action") {
		t.Fatalf("err=%v", err)
	}
}
