package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteTemplateSkipExists(t *testing.T) {
	dir := t.TempDir()
	if err := writeTemplate(dir, "AGENTS.md", "AGENTS.md", false); err != nil {
		t.Fatal(err)
	}
	if err := writeTemplate(dir, "AGENTS.md", "AGENTS.md", false); err == nil {
		t.Fatal("expected error when file exists")
	}
}

func TestWriteTemplateForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeTemplate(dir, "main.go", "main-scratch.tmpl", true); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 50 {
		t.Fatalf("short main.go: %q", b)
	}
}
