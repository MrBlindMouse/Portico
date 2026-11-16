package installation_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
)

func TestValidateWireBadActionRef(t *testing.T) {
	err := installation.ValidateWire(installation.Wire{
		Root: "home",
		Scenes: map[string]installation.WireScene{
			"home": {Inputs: map[string]string{"x": "missing"}},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "undefined action") {
		t.Fatalf("err=%v", err)
	}
}

func TestValidateWireOK(t *testing.T) {
	err := installation.ValidateWire(installation.Wire{
		Root:    "home",
		Actions: []string{"go"},
		Scenes: map[string]installation.WireScene{
			"home": {Inputs: map[string]string{"go": "go"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstallationWireRoundTrip(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	run := func(components.ActionContext) []components.Effect { return nil }
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: render,
			Inputs: map[string]components.ActionID{"go": "go"},
		}).
		Scene("away", components.Scene{Render: render}).
		Action("go", components.Action{Run: run}).
		Space("lobby", components.Space{}).
		Asset("welcome", "hi").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	w := exp.Wire()
	if err := installation.ValidateWire(w); err != nil {
		t.Fatal(err)
	}
	if w.Scenes["home"].Inputs["go"] != "go" {
		t.Fatalf("inputs=%v", w.Scenes["home"].Inputs)
	}
	if len(w.Assets) != 1 || w.Assets[0] != "welcome" {
		t.Fatalf("assets=%v", w.Assets)
	}
	delete(w.Scenes, "away")
	preview := installation.PreviewReload(exp, w, []installation.LiveSession{
		{ID: "s1", Scene: "away"},
	})
	if preview.OK || len(preview.StrandedSessions) != 1 {
		t.Fatalf("%+v", preview)
	}
}
