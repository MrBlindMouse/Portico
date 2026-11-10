package archtest_test

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ponytail: naive import scan via build.Import; sufficient for boundary checks.

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

func importsUnder(t *testing.T, root, importPath string) []string {
	t.Helper()
	pkg, err := build.Import(importPath, root, 0)
	if err != nil {
		t.Fatalf("import %s: %v", importPath, err)
	}
	return pkg.Imports
}

func TestInstallationDoesNotImportRuntime(t *testing.T) {
	root := moduleRoot(t)
	const exp = "github.com/MrBlindMouse/Portico/installation"
	for _, imp := range importsUnder(t, root, exp) {
		if strings.Contains(imp, "/runtime") {
			t.Fatalf("installation must not import runtime, found %s", imp)
		}
	}
}

func TestRuntimeMayImportInstallation(t *testing.T) {
	root := moduleRoot(t)
	const rt = "github.com/MrBlindMouse/Portico/runtime"
	imps := importsUnder(t, root, rt)
	found := false
	for _, imp := range imps {
		if imp == "github.com/MrBlindMouse/Portico/installation" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("runtime should import installation for orchestration")
	}
}

func TestComponentsDoNotImportInstallationOrRuntime(t *testing.T) {
	root := moduleRoot(t)
	const pkg = "github.com/MrBlindMouse/Portico/components"
	for _, imp := range importsUnder(t, root, pkg) {
		if strings.Contains(imp, "/installation") || strings.Contains(imp, "/runtime") {
			t.Fatalf("components must not import %s", imp)
		}
	}
}

func TestForbiddenInstallationImports(t *testing.T) {
	root := moduleRoot(t)
	const exp = "github.com/MrBlindMouse/Portico/installation"
	for _, imp := range importsUnder(t, root, exp) {
		if strings.Contains(imp, "/internal/plans") || strings.Contains(imp, "/examples") {
			t.Fatalf("installation must not import %s", imp)
		}
	}
}

func TestSessionStateMutationInApplyOnly(t *testing.T) {
	root := moduleRoot(t)
	runtimeDir := filepath.Join(root, "runtime")
	allowed := map[string]bool{
		"apply.go":   true,
		"session.go": true, // identity/render metadata only
		"context.go": true, // read-only session access for scene/action contexts
	}
	err := filepath.Walk(runtimeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		base := filepath.Base(path)
		if allowed[base] {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		body := string(b)
		for _, needle := range []string{"sess.current =", "sess.local[", "delete(sess.spaces"} {
			if strings.Contains(body, needle) {
				t.Errorf("%s must not mutate session fields directly (use apply.go)", base)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInspectDoesNotImportPlans(t *testing.T) {
	root := moduleRoot(t)
	const insp = "github.com/MrBlindMouse/Portico/inspect"
	for _, imp := range importsUnder(t, root, insp) {
		if strings.Contains(imp, "/internal/plans") {
			t.Fatalf("inspect must not import plans, found %s", imp)
		}
	}
}

func TestAllPackagesCompile(t *testing.T) {
	root := moduleRoot(t)
	pkgs := []string{
		"github.com/MrBlindMouse/Portico",
		"github.com/MrBlindMouse/Portico/components",
		"github.com/MrBlindMouse/Portico/installation",
		"github.com/MrBlindMouse/Portico/runtime",
		"github.com/MrBlindMouse/Portico/identity",
		"github.com/MrBlindMouse/Portico/render",
		"github.com/MrBlindMouse/Portico/inspect",
		"github.com/MrBlindMouse/Portico/internal/plans",
		"github.com/MrBlindMouse/Portico/asset",
		"github.com/MrBlindMouse/Portico/config",
		"github.com/MrBlindMouse/Portico/testkit",
	}
	for _, p := range pkgs {
		if _, err := build.Import(p, root, 0); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
	fmt.Fprintf(os.Stderr, "archtest: checked %d packages\n", len(pkgs))
}
