package config_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/config"
)

func TestNewLogger(t *testing.T) {
	if config.NewLogger("info") == nil || config.NewLogger("debug") == nil {
		t.Fatal("expected logger")
	}
}
