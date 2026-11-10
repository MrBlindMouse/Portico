package installation_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
)

func TestPreviewReloadStranded(t *testing.T) {
	cur, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Scene("other", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	wire := installation.Wire{
		Root:    "home",
		Scenes:  map[string]installation.WireScene{"home": {}},
		Actions: []string{},
	}
	preview := installation.PreviewReload(cur, wire, []installation.LiveSession{
		{ID: "s1", Scene: "other"},
	})
	if preview.OK || len(preview.StrandedSessions) != 1 {
		t.Fatalf("%+v", preview)
	}
}

func TestPreviewReloadOK(t *testing.T) {
	wire := installation.Wire{
		Root: "home",
		Scenes: map[string]installation.WireScene{
			"home": {Inputs: map[string]string{"go": "go"}},
		},
		Actions: []string{"go"},
	}
	preview := installation.PreviewReload(nil, wire, []installation.LiveSession{
		{ID: "s1", Scene: "home"},
	})
	if !preview.OK {
		t.Fatalf("%+v", preview)
	}
}

func TestPreviewReloadStrandsRemovedSpace(t *testing.T) {
	wire := installation.Wire{
		Root:   "home",
		Scenes: map[string]installation.WireScene{"home": {}},
	}
	preview := installation.PreviewReload(nil, wire, []installation.LiveSession{
		{ID: "s1", Scene: "home", Spaces: []components.SpaceID{"lobby"}},
	})
	if preview.OK || len(preview.StrandedSessions) != 1 {
		t.Fatalf("%+v", preview)
	}
}
