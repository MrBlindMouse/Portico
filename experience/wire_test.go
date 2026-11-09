package experience_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
)

func TestValidateWireBadActionRef(t *testing.T) {
	err := experience.ValidateWire(experience.Wire{
		Root: "home",
		Scenes: map[string]experience.WireScene{
			"home": {Inputs: map[string]string{"x": "missing"}},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "undefined action") {
		t.Fatalf("err=%v", err)
	}
}

func TestValidateWireOK(t *testing.T) {
	err := experience.ValidateWire(experience.Wire{
		Root:    "home",
		Actions: []string{"go"},
		Scenes: map[string]experience.WireScene{
			"home": {Inputs: map[string]string{"go": "go"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
