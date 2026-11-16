package portico_test

import (
	"testing"
	"time"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/config"
)

func TestMergeConfig(t *testing.T) {
	o := portico.Options{}.MergeConfig(config.File{
		MaxConns:       10,
		MaxConnsPerIP:  3,
		ActionRate:     2,
		AssetDir:       "/assets",
		AssetPollEvery: 5 * time.Second,
		AllowKeys:      []string{"fp1"},
		DenyKeys:       []string{"fp2"},
	})
	if o.MaxConns != 10 || o.MaxConnsPerIP != 3 || o.ActionRate != 2 {
		t.Fatalf("%+v", o)
	}
	if o.AssetDir != "/assets" || o.AssetPollEvery != 5*time.Second {
		t.Fatalf("%+v", o)
	}
	if len(o.AllowKeys) != 1 || len(o.DenyKeys) != 1 {
		t.Fatalf("%+v", o)
	}
	off := portico.Options{}.MergeConfig(config.File{ActionRate: -1})
	if off.ActionRate != -1 {
		t.Fatalf("rate=%v", off.ActionRate)
	}
	withLog := portico.Options{}.MergeConfig(config.File{LogLevel: "debug", IdleTimeout: -time.Second})
	if withLog.LogLevel != "debug" || withLog.IdleTimeout != -time.Second {
		t.Fatalf("log=%q idle=%v", withLog.LogLevel, withLog.IdleTimeout)
	}
	if portico.TranscriptKeep == 0 || portico.TranscriptOff == 0 {
		t.Fatal("transcript constants")
	}
}
