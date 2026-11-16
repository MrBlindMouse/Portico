package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
)

func TestHostKeyReadFailureDoesNotReplaceKey(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, nil)

	dir := t.TempDir()
	_, err = NewServer(ServerConfig{HostKeyPath: dir}, eng)
	if err == nil || !strings.Contains(err.Error(), "host key") {
		t.Fatalf("dir path err=%v", err)
	}

	path := filepath.Join(t.TempDir(), "ssh_host_rsa")
	if err := os.WriteFile(path, []byte("not a key"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = NewServer(ServerConfig{HostKeyPath: path}, eng)
	if err == nil || !strings.Contains(err.Error(), "host key") {
		t.Fatalf("bad key err=%v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "not a key" {
		t.Fatalf("host key overwritten: %q", b)
	}
}
