package runtime_test

import (
	"bytes"
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestRenderAssetAny(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string {
				v, ok := ctx.AssetAny("logo")
				if !ok {
					return "missing"
				}
				buf, ok := v.([]byte)
				if !ok {
					return "bad type"
				}
				return string(buf)
			},
		}).
		AssetAny("logo", []byte("LOGO")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, nil)
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), "home")
	text, err := eng.RenderText(sess)
	if err != nil {
		t.Fatal(err)
	}
	if text != "LOGO" {
		t.Fatalf("render=%q", text)
	}
}

func TestUpdateAssetsKeepsAssetAny(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string {
				v, _ := ctx.AssetAny("pinned")
				buf, _ := v.([]byte)
				s, _ := ctx.Asset("poll.txt")
				return s + "|" + string(buf)
			},
		}).
		AssetAny("pinned", []byte("stay")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, nil)
	if err := eng.UpdateAssets(map[string]string{"poll.txt": "from dir"}); err != nil {
		t.Fatal(err)
	}
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), "home")
	text, err := eng.RenderText(sess)
	if err != nil {
		t.Fatal(err)
	}
	if text != "from dir|stay" {
		t.Fatalf("render=%q", text)
	}
}

func TestUpdateAssetAnySwapsMap(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string {
				v, ok := ctx.AssetAny("x")
				if !ok {
					return "missing"
				}
				buf, ok := v.([]byte)
				if !ok {
					return "bad"
				}
				return string(buf)
			},
		}).
		AssetAny("x", []byte("old")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, nil)
	if err := eng.UpdateAssetAny(map[string]any{"x": []byte("new")}); err != nil {
		t.Fatal(err)
	}
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), "home")
	text, err := eng.RenderText(sess)
	if err != nil {
		t.Fatal(err)
	}
	if text != "new" {
		t.Fatalf("render=%q", text)
	}
	v, ok := exp.AssetAny("x")
	if !ok {
		t.Fatal("original missing")
	}
	buf, ok := v.([]byte)
	if !ok || !bytes.Equal(buf, []byte("old")) {
		t.Fatalf("original installation mutated: %q", buf)
	}
}