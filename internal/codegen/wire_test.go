package codegen_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/internal/codegen"
)

func TestFromWireGolden(t *testing.T) {
	out := codegen.FromWire(experience.Wire{
		Root: "home",
		Scenes: map[string]experience.WireScene{
			"home": {Inputs: map[string]string{"go": "go"}},
		},
		Actions: []string{"go"},
	}, "main")
	if !strings.Contains(out, `Root("home")`) || !strings.Contains(out, `Action("go"`) {
		t.Fatalf("%s", out)
	}
}
