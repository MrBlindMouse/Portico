package installation_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
)

func TestBuildAssetAndAssetAny(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Asset("note.txt", "hello").
		AssetAny("logo", []byte{1, 2, 3}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := exp.Asset("note.txt"); !ok || s != "hello" {
		t.Fatalf("asset=%q ok=%v", s, ok)
	}
	v, ok := exp.AssetAny("logo")
	if !ok {
		t.Fatal("missing logo")
	}
	buf, ok := v.([]byte)
	if !ok || !bytes.Equal(buf, []byte{1, 2, 3}) {
		t.Fatalf("logo=%T %v", v, v)
	}
}

func TestCloneWithAssetsKeepsAssetAny(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Asset("keep.txt", "gone").
		AssetAny("pinned", []byte("stay")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	next := exp.CloneWithAssets(map[string]string{"dir.txt": "from poll"})
	if _, ok := next.Asset("keep.txt"); ok {
		t.Fatal("string asset should have been replaced")
	}
	if s, ok := next.Asset("dir.txt"); !ok || s != "from poll" {
		t.Fatalf("dir=%q ok=%v", s, ok)
	}
	v, ok := next.AssetAny("pinned")
	if !ok {
		t.Fatal("any asset dropped")
	}
	buf, ok := v.([]byte)
	if !ok || string(buf) != "stay" {
		t.Fatalf("pinned=%v", v)
	}
}

func TestMergeAssetAny(t *testing.T) {
	render := func(components.RenderContext) string { return "" }
	a := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: render}).
		AssetAny("a", 1)
	b := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: render}).
		AssetAny("b", 2).
		AssetAny("a", 9)
	exp, err := a.Merge(b).Build()
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := exp.AssetAny("a"); !ok || v.(int) != 9 {
		t.Fatalf("a=%v ok=%v", v, ok)
	}
	if v, ok := exp.AssetAny("b"); !ok || v.(int) != 2 {
		t.Fatalf("b=%v ok=%v", v, ok)
	}
}

func TestAssetAnyInvalidName(t *testing.T) {
	_, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		AssetAny("sub/name", []byte("x")).
		Build()
	if err == nil || !strings.Contains(err.Error(), "invalid asset name") {
		t.Fatalf("err=%v", err)
	}
}