package inspect_test

import (
	"bytes"
	"encoding/json"
	"github.com/MrBlindMouse/Portico/components"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/inspect"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestSnapshotJSON(t *testing.T) {
	exp, _ := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "hello-preview" },
		}).
		Build()
	eng := runtime.NewEngine(exp, nil)
	sess := runtime.NewSessionForTest("t1", identity.Anonymous(), "home")
	eng.RegisterSession(sess)
	h := inspect.NewHandler(eng)
	req := httptest.NewRequest(http.MethodGet, "/inspect/snapshot", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var snap inspect.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Installation.Root != "home" || len(snap.Sessions) != 1 {
		t.Fatalf("%+v", snap)
	}
	if snap.Sessions[0].RenderSize == 0 {
		t.Fatalf("expected render metadata")
	}
}

func TestReloadPreviewEndpoint(t *testing.T) {
	exp, _ := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "" },
		}).
		Scene("gone", components.Scene{
			Render: func(components.RenderContext) string { return "" },
		}).
		Build()
	eng := runtime.NewEngine(exp, nil)
	eng.RegisterSession(runtime.NewSessionForTest("s1", identity.Anonymous(), "gone"))
	h := inspect.NewHandler(eng)
	body := []byte(`{"root":"home","scenes":{"home":{}}}`)
	req := httptest.NewRequest(http.MethodPost, "/inspect/reload-preview", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var preview installation.ReloadPreview
	if err := json.Unmarshal(rec.Body.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	if preview.OK || len(preview.StrandedSessions) != 1 {
		t.Fatalf("%+v", preview)
	}
}

func TestValidateWireEndpoint(t *testing.T) {
	eng := runtime.NewEngine(mustExp(t), nil)
	h := inspect.NewHandler(eng)
	body := []byte(`{"root":"home","actions":["go"],"scenes":{"home":{"inputs":{"go":"go"}}}}`)
	req := httptest.NewRequest(http.MethodPost, "/inspect/validate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["ok"] != true {
		t.Fatalf("%v", res)
	}
}

func mustExp(t *testing.T) *installation.Installation {
	t.Helper()
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	return exp
}
