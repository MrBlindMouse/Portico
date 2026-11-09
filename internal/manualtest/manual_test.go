package manualtest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestManualLinksExist parses docs/manual/README.md relative links and checks targets exist.
func TestManualLinksExist(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	manualDir := filepath.Join(root, "docs", "manual")
	readme := filepath.Join(manualDir, "README.md")
	body, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`\]\(([^)#]+)`)
	seen := make(map[string]struct{})
	for _, m := range re.FindAllStringSubmatch(string(body), -1) {
		rel := m[1]
		if strings.HasPrefix(rel, "http") {
			continue
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		seen[rel] = struct{}{}
		target := filepath.Join(manualDir, filepath.FromSlash(rel))
		if _, err := os.Stat(target); err != nil {
			t.Errorf("manual link %q -> %s: %v", rel, target, err)
		}
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
