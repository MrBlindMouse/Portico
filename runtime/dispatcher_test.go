package runtime_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestResolveInputPrefixCommand(t *testing.T) {
	exp, _ := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "" },
			Inputs: map[string]components.ActionID{"go": "go"},
		}).
		Action("go", components.Action{Run: func(components.ActionContext) []components.Effect { return nil }}).
		Build()
	aid, err := runtime.ResolveInput(exp, "home", "go elsewhere")
	if err != nil || aid != "go" {
		t.Fatalf("aid=%q err=%v", aid, err)
	}
}

func TestInputEligible(t *testing.T) {
	exp, _ := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "" },
			Inputs: map[string]components.ActionID{"go": "go", "OK": "ack"},
		}).
		Action("go", components.Action{Run: func(components.ActionContext) []components.Effect { return nil }}).
		Action("ack", components.Action{Run: func(components.ActionContext) []components.Effect { return nil }}).
		Build()
	if !runtime.InputEligible(exp, "home", "go there") {
		t.Fatal("prefix should be eligible")
	}
	if !runtime.InputEligible(exp, "home", "OK") {
		t.Fatal("OK registered — should be eligible")
	}
	exp2, _ := installation.NewBuilder().
		Root("landing").
		Scene("landing", components.Scene{
			Render: func(components.RenderContext) string { return "" },
			Inputs: map[string]components.ActionID{"greet": "g"},
		}).
		Action("g", components.Action{Run: func(components.ActionContext) []components.Effect { return nil }}).
		Build()
	if runtime.InputEligible(exp2, "landing", "OK") {
		t.Fatal("OK not registered on landing")
	}
	if runtime.InputEligible(exp2, "landing", "KITTY_DATA_START") {
		t.Fatal("kitty marker not a command")
	}
}

func TestResolveInputMissing(t *testing.T) {
	exp, _ := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "" },
		}).
		Build()
	_, err := runtime.ResolveInput(exp, "home", "nope")
	if err == nil || !strings.Contains(err.Error(), "no action") {
		t.Fatalf("err=%v", err)
	}
}
