package codegen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/codegen"
)

func TestFromWireGolden(t *testing.T) {
	out := codegen.FromWire(installation.Wire{
		Root: "home",
		Scenes: map[string]installation.WireScene{
			"home": {Inputs: map[string]string{"go": "go"}},
		},
		Actions: []string{"go"},
	}, "main")
	if !strings.Contains(out, `Root("home")`) || !strings.Contains(out, `Action("go"`) {
		t.Fatalf("%s", out)
	}
	if !strings.Contains(out, `"github.com/MrBlindMouse/Portico/components"`) {
		t.Fatalf("missing components import:\n%s", out)
	}
}

func TestFromWireCompiles(t *testing.T) {
	out := codegen.FromWire(installation.Wire{
		Root: "home",
		Scenes: map[string]installation.WireScene{
			"home": {Inputs: map[string]string{"go": "go"}},
		},
		Actions: []string{"go"},
		Spaces:  []string{"room"},
	}, "site")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "wire.go"), []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	mod := "module example.com/wiregen\n\ngo 1.23\n\nrequire github.com/MrBlindMouse/Portico v0.0.0\n\nreplace github.com/MrBlindMouse/Portico => " + moduleRoot(t) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	if b, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("tidy: %v\n%s", err, b)
	}
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build: %v\n%s", err, b)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
