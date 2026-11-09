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

func TestBuildKeySceneRejectsWords(t *testing.T) {
	_, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "hi" },
			Keys:   true,
			Inputs: map[string]experience.ActionID{"north": "go"},
		}).
		Action("go", experience.Action{Run: func(experience.ActionContext) []experience.Effect { return nil }}).
		Build()
	if err == nil || !strings.Contains(err.Error(), "one character") {
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

func TestBindInputMissingScene(t *testing.T) {
	_, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "hi" }}).
		BindInput("missing", "go", "go").
		Build()
	if err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("got err=%v", err)
	}
}

func TestBuildDoesNotShareInputMap(t *testing.T) {
	render := func(experience.RenderContext) string { return "" }
	run := func(experience.ActionContext) []experience.Effect { return nil }
	b := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: render,
			Inputs: map[string]experience.ActionID{"stay": "go"},
		}).
		Action("go", experience.Action{Run: run})
	exp, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	b.BindInput("home", "go", "go")
	sc, _ := exp.Scene("home")
	if _, ok := sc.Inputs["go"]; ok {
		t.Fatal("BindInput changed the built experience")
	}
	sc.Inputs["extra"] = "go"
	again, _ := exp.Scene("home")
	if _, ok := again.Inputs["extra"]; ok {
		t.Fatal("Scene shared its input map")
	}
}

func TestBuildSnapshotsMergedInputs(t *testing.T) {
	render := func(experience.RenderContext) string { return "" }
	run := func(experience.ActionContext) []experience.Effect { return nil }
	src := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: render, Inputs: map[string]experience.ActionID{}}).
		Action("go", experience.Action{Run: run})
	exp, err := experience.NewBuilder().Merge(src).Build()
	if err != nil {
		t.Fatal(err)
	}
	src.BindInput("home", "go", "go")
	sc, _ := exp.Scene("home")
	if _, ok := sc.Inputs["go"]; ok {
		t.Fatal("BindInput changed the built experience")
	}
}

func TestScenesAreSorted(t *testing.T) {
	render := func(experience.RenderContext) string { return "" }
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("b", experience.Scene{Render: render}).
		Scene("a", experience.Scene{Render: render}).
		Scene("home", experience.Scene{Render: render}).
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
	render := func(experience.RenderContext) string { return "" }
	const room = experience.SpaceID("room")
	base := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: render}).
		OnSessionStart(experience.SetLocalEffect("boot", true)).
		OnSessionEnd(experience.SetLocalEffect("bye", true))
	other := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: render}).
		Space(room, experience.Space{}).
		OnSessionStart(experience.JoinSpaceEffect(room)).
		OnSessionEnd(experience.LeaveSpaceEffect(room))
	exp, err := base.Merge(other).Build()
	if err != nil {
		t.Fatal(err)
	}
	start := exp.OnSessionStart()
	if len(start) != 2 {
		t.Fatalf("start=%d", len(start))
	}
	if _, ok := start[0].(experience.SetLocal); !ok {
		t.Fatalf("start[0]=%T", start[0])
	}
	if _, ok := start[1].(experience.JoinSpace); !ok {
		t.Fatalf("start[1]=%T", start[1])
	}
	if len(exp.OnSessionEnd()) != 2 {
		t.Fatalf("end=%d", len(exp.OnSessionEnd()))
	}
}
