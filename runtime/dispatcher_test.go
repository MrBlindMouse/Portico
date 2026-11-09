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

func TestInputEligible(t *testing.T) {
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "" },
			Inputs: map[string]experience.ActionID{"go": "go", "OK": "ack"},
		}).
		Action("go", experience.Action{Run: func(experience.ActionContext) []experience.Effect { return nil }}).
		Action("ack", experience.Action{Run: func(experience.ActionContext) []experience.Effect { return nil }}).
		Build()
	if !runtime.InputEligible(exp, "home", "go there") {
		t.Fatal("prefix should be eligible")
	}
	if !runtime.InputEligible(exp, "home", "OK") {
		t.Fatal("OK registered — should be eligible")
	}
	exp2, _ := experience.NewBuilder().
		Root("landing").
		Scene("landing", experience.Scene{
			Render: func(experience.RenderContext) string { return "" },
			Inputs: map[string]experience.ActionID{"greet": "g"},
		}).
		Action("g", experience.Action{Run: func(experience.ActionContext) []experience.Effect { return nil }}).
		Build()
	if runtime.InputEligible(exp2, "landing", "OK") {
		t.Fatal("OK not registered on landing")
	}
	if runtime.InputEligible(exp2, "landing", "KITTY_DATA_START") {
		t.Fatal("kitty marker not a command")
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
