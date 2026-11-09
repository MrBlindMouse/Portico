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
