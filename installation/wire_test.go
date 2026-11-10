package installation_test

import (
	"strings"
	"testing"

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
