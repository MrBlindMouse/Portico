package portico_test

import (
	"strings"
	"testing"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/plans"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestServeWarningsRegisterWithoutAccounts(t *testing.T) {
	exp, err := plans.Registration().Build()
	if err != nil {
		t.Fatal(err)
	}
	w := portico.ServeWarnings(exp, portico.Options{DataDir: ".portico-data"})
	if len(w) != 1 || !strings.Contains(w[0], "register") {
		t.Fatalf("%v", w)
	}
	store := runtime.NewAccountStore()
	w = portico.ServeWarnings(exp, portico.Options{
		AccountStore: store,
		Accounts:     &runtime.AccountConfig{Enabled: true, Accounts: store},
	})
	if len(w) != 0 {
		t.Fatalf("%v", w)
	}
}

func TestServeWarningsBlogWithoutAssets(t *testing.T) {
	exp, err := plans.Blog().Build()
	if err != nil {
		t.Fatal(err)
	}
	w := portico.ServeWarnings(exp, portico.Options{})
	if len(w) != 1 || !strings.Contains(w[0], "blog-list") {
		t.Fatalf("%v", w)
	}
	exp2, err := installation.NewBuilder().
		Root("blog-list").
		Scene("blog-list", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Asset("post-1.md", "hi").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	w = portico.ServeWarnings(exp2, portico.Options{})
	if len(w) != 0 {
		t.Fatalf("%v", w)
	}
	w = portico.ServeWarnings(exp2, portico.Options{AssetDir: "posts"})
	if len(w) != 0 {
		t.Fatalf("AssetDir alone: %v", w)
	}
}