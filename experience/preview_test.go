package experience_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
)

func TestPreviewReloadStranded(t *testing.T) {
	cur, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Scene("other", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	wire := experience.Wire{
		Root:    "home",
		Scenes:  map[string]experience.WireScene{"home": {}},
		Actions: []string{},
	}
	preview := experience.PreviewReload(cur, wire, []experience.LiveSession{
		{ID: "s1", Scene: "other"},
	})
	if preview.OK || len(preview.StrandedSessions) != 1 {
		t.Fatalf("%+v", preview)
	}
}

func TestPreviewReloadOK(t *testing.T) {
	wire := experience.Wire{
		Root: "home",
		Scenes: map[string]experience.WireScene{
			"home": {Inputs: map[string]string{"go": "go"}},
		},
		Actions: []string{"go"},
	}
	preview := experience.PreviewReload(nil, wire, []experience.LiveSession{
		{ID: "s1", Scene: "home"},
	})
	if !preview.OK {
		t.Fatalf("%+v", preview)
	}
}
