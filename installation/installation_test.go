package installation_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
)

func TestBuildUndefinedAction(t *testing.T) {
	_, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "hi" },
			Inputs: map[string]components.ActionID{"x": "missing"},
		}).
		Build()
	if err == nil || !strings.Contains(err.Error(), "undefined action") {
		t.Fatalf("got err=%v", err)
	}
}

func TestBuildKeySceneRejectsWords(t *testing.T) {
	_, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "hi" },
			Keys:   true,
			Inputs: map[string]components.ActionID{"north": "go"},
		}).
		Action("go", components.Action{Run: func(components.ActionContext) []components.Effect { return nil }}).
		Build()
	if err == nil || !strings.Contains(err.Error(), "one character") {
		t.Fatalf("got err=%v", err)
	}
}

func TestBuildRequiresRoot(t *testing.T) {
	_, err := installation.NewBuilder().
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "hi" },
		}).
		Build()
	if err == nil || !strings.Contains(err.Error(), "root scene") {
		t.Fatalf("got err=%v", err)
	}
}

func TestBindInputMissingScene(t *testing.T) {
	_, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "hi" }}).
		BindInput("missing", "go", "go").
		Build()
	if err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("got err=%v", err)
	}
}

func TestBuildDoesNotShareInputMap(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	run := func(components.ActionContext) []components.Effect { return nil }
	b := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: render,
			Inputs: map[string]components.ActionID{"stay": "go"},
		}).
		Action("go", components.Action{Run: run})
	exp, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	b.BindInput("home", "go", "go")
	sc, _ := exp.Scene("home")
	if _, ok := sc.Inputs["go"]; ok {
		t.Fatal("BindInput changed the built installation")
	}
	sc.Inputs["extra"] = "go"
	again, _ := exp.Scene("home")
	if _, ok := again.Inputs["extra"]; ok {
		t.Fatal("Scene shared its input map")
	}
}

func TestBuildSnapshotsMergedInputs(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	run := func(components.ActionContext) []components.Effect { return nil }
	src := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: render, Inputs: map[string]components.ActionID{}}).
		Action("go", components.Action{Run: run})
	exp, err := installation.NewBuilder().Merge(src).Build()
	if err != nil {
		t.Fatal(err)
	}
	src.BindInput("home", "go", "go")
	sc, _ := exp.Scene("home")
	if _, ok := sc.Inputs["go"]; ok {
		t.Fatal("BindInput changed the built installation")
	}
}

func TestScenesAreSorted(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("b", components.Scene{Render: render}).
		Scene("a", components.Scene{Render: render}).
		Scene("home", components.Scene{Render: render}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	got := exp.Scenes()
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "home" {
		t.Fatalf("scenes=%v", got)
	}
}

func TestMergeAppendsSessionHooks(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	const room = components.SpaceID("room")
	base := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: render}).
		OnSessionStart(components.SetLocalEffect("boot", true)).
		OnSessionEnd(components.SetLocalEffect("bye", true))
	other := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: render}).
		Space(room, components.Space{}).
		OnSessionStart(components.JoinSpaceEffect(room)).
		OnSessionEnd(components.LeaveSpaceEffect(room))
	exp, err := base.Merge(other).Build()
	if err != nil {
		t.Fatal(err)
	}
	start := exp.OnSessionStart()
	if len(start) != 2 {
		t.Fatalf("start=%d", len(start))
	}
	if _, ok := start[0].(components.SetLocal); !ok {
		t.Fatalf("start[0]=%T", start[0])
	}
	if _, ok := start[1].(components.JoinSpace); !ok {
		t.Fatalf("start[1]=%T", start[1])
	}
	if len(exp.OnSessionEnd()) != 2 {
		t.Fatalf("end=%d", len(exp.OnSessionEnd()))
	}
}
