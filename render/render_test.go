package render_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/render"
)

func TestBoxContainsTitle(t *testing.T) {
	out := render.Box("Hello", "body", 24)
	if !strings.Contains(out, "Hello") || !strings.Contains(out, "body") {
		t.Fatalf("box output=%q", out)
	}
}
