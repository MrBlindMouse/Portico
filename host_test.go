package portico_test

import (
	"os"
	"path/filepath"
	"testing"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestNewHostLoadsAssetDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "banner.txt"), []byte("ASSET_OK"), 0o644); err != nil {
		t.Fatal(err)
	}
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string {
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
	exp := pingExperience(t)
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
	exp := pingExperience(t)
	var seen string
	host, err := portico.NewHost("127.0.0.1:0", exp, portico.Options{
		ActionRate: -1,
		ActionBefore: func(sess experience.SessionView, id experience.ActionID, line string) (string, error) {
			seen = sess.ID() + ":" + string(id) + ":" + line
			return line, nil
		},
		ActionAfter: func(_ experience.SessionView, _ experience.ActionID, err error) {
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

func pingExperience(t *testing.T) *experience.Experience {
	t.Helper()
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "ok" },
			Inputs: map[string]experience.ActionID{"ping": "ping"},
		}).
		Action("ping", experience.Action{
			Run: func(experience.ActionContext) []experience.Effect { return nil },
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	return exp
}
