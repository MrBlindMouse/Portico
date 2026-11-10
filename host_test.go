package portico_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"os"
	"path/filepath"
	"testing"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestNewHostLoadsAssetDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "banner.txt"), []byte("ASSET_OK"), 0o644); err != nil {
		t.Fatal(err)
	}
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string {
				s, ok := ctx.Asset("banner.txt")
				if !ok {
					return "missing"
				}
				return s
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	host, err := portico.NewHost(":0", exp, portico.Options{AssetDir: dir, InspectAddr: "", StatusAddr: ""})
	if err != nil {
		t.Fatal(err)
	}
	sess := runtime.NewSessionForTest("t", identity.Anonymous(), "home")
	text, err := host.Server.Engine().RenderText(sess)
	if err != nil {
		t.Fatal(err)
	}
	if text != "ASSET_OK" {
		t.Fatalf("render=%q", text)
	}
}

func TestNegativeActionRateDisablesLimit(t *testing.T) {
	exp := pingInstallation(t)
	host, err := portico.NewHost("127.0.0.1:0", exp, portico.Options{ActionRate: -1})
	if err != nil {
		t.Fatal(err)
	}
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), "home")
	for i := 0; i < 30; i++ {
		if err := host.Server.Engine().HandleInput(sess, "ping"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestActionHooksDoNotNeedRuntimeImport(t *testing.T) {
	exp := pingInstallation(t)
	var seen string
	host, err := portico.NewHost("127.0.0.1:0", exp, portico.Options{
		ActionRate: -1,
		ActionBefore: func(sess components.SessionView, id components.ActionID, line string) (string, error) {
			seen = sess.ID() + ":" + string(id) + ":" + line
			return line, nil
		},
		ActionAfter: func(_ components.SessionView, _ components.ActionID, err error) {
			if err != nil {
				t.Errorf("after: %v", err)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), "home")
	if err := host.Server.Engine().HandleInput(sess, "ping"); err != nil {
		t.Fatal(err)
	}
	if seen != "s:ping:ping" {
		t.Fatalf("seen=%q", seen)
	}
}

func pingInstallation(t *testing.T) *installation.Installation {
	t.Helper()
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "ok" },
			Inputs: map[string]components.ActionID{"ping": "ping"},
		}).
		Action("ping", components.Action{
			Run: func(components.ActionContext) []components.Effect { return nil },
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	return exp
}
