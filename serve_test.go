package portico

import (
	"context"
	"log/slog"
	"testing"
)

func TestWithDefaultsLogLevel(t *testing.T) {
	o := Options{LogLevel: "debug"}.withDefaults()
	if o.Logger == nil {
		t.Fatal("expected logger")
	}
	if !o.Logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("LogLevel=debug should enable debug logs")
	}
}

func TestWithDefaultsInfoWhenUnset(t *testing.T) {
	o := Options{}.withDefaults()
	if o.Logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("default should be info, not debug")
	}
}
