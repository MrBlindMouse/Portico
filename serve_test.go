package portico

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/render"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestResolveTheme(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	if resolveTheme(Options{}).Bold("a") != render.Default.Bold("a") {
		t.Fatal("default theme")
	}
	t.Setenv("NO_COLOR", "1")
	if resolveTheme(Options{}).Bold("a") != "a" {
		t.Fatal("NO_COLOR")
	}
	custom := render.Default
	if resolveTheme(Options{Theme: &custom}).Bold("a") != render.Default.Bold("a") {
		t.Fatal("explicit theme")
	}
}

func TestWithDefaultsLogLevel(t *testing.T) {
	o := Options{LogLevel: "debug"}.withDefaults()
	if o.Logger == nil {
		t.Fatal("expected logger")
	}
	if !o.Logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("LogLevel=debug should enable debug logs")
	}
}

func TestServerConfigKeepsIdleDefault(t *testing.T) {
	cfg, err := serverConfig(":2222", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IdleTimeout != 30*time.Minute {
		t.Fatalf("idle=%v", cfg.IdleTimeout)
	}
	off, err := serverConfig(":2222", Options{IdleTimeout: -1})
	if err != nil {
		t.Fatal(err)
	}
	if off.IdleTimeout != 0 {
		t.Fatalf("idle=%v", off.IdleTimeout)
	}
	custom, err := serverConfig(":2222", Options{IdleTimeout: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if custom.IdleTimeout != time.Minute {
		t.Fatalf("idle=%v", custom.IdleTimeout)
	}
}

func TestHashedAccountsOpenErrorDoesNotFallThrough(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	plain := runtime.NewAccountStore()
	_, err := resolveAccountStore(Options{
		HashedAccounts: true,
		DataDir:        blocker,
		AccountStore:   plain,
	})
	if err == nil {
		t.Fatal("expected hashed account open to fail")
	}
	if _, err := resolveAccountStore(Options{HashedAccounts: true}); err == nil {
		t.Fatal("expected missing DataDir to fail")
	}
	store, err := resolveAccountStore(Options{HashedAccounts: true, DataDir: t.TempDir(), AccountStore: plain})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store.(*identity.FileAccountStore); !ok {
		t.Fatalf("store=%T", store)
	}
}

func TestWithDefaultsInfoWhenUnset(t *testing.T) {
	o := Options{}.withDefaults()
	if o.Logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("default should be info, not debug")
	}
}
