package render_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/render"
)

func TestFormatWriteupHeadings(t *testing.T) {
	md := "# Title\n\npara\n\n## Sub\n\nmore"
	out := render.FormatWriteup(md, 0, render.Plain)
	if !strings.Contains(out, "Title") || !strings.Contains(out, "Sub") {
		t.Fatalf("%q", out)
	}
	// spacing: title block separated from para by blank lines from h1 rule
	if strings.Count(out, "\n\n") < 2 {
		t.Fatalf("spacing %q", out)
	}
}

func TestFormatWriteupHeadingsBoldANSI(t *testing.T) {
	out := render.FormatWriteup("# Hi", 0, render.Default)
	if !strings.Contains(out, "\x1b[1m") {
		t.Fatalf("%q", out)
	}
}

func TestFormatWriteupNestedList(t *testing.T) {
	md := "- one\n  - two\n    - three\n- four"
	out := render.FormatWriteup(md, 0, render.Plain)
	if !strings.Contains(out, "  • one") {
		t.Fatalf("%q", out)
	}
	if !strings.Contains(out, "    • two") {
		t.Fatalf("%q", out)
	}
	if !strings.Contains(out, "      • three") {
		t.Fatalf("%q", out)
	}
	if !strings.Contains(out, "  • four") {
		t.Fatalf("%q", out)
	}
}

func TestFormatWriteupInline(t *testing.T) {
	out := render.FormatWriteup("say **bold** and `code`", 0, render.Default)
	if !strings.Contains(out, "\x1b[1m") || !strings.Contains(out, "\x1b[90m") {
		t.Fatalf("%q", out)
	}
}

func TestFormatWriteupWrap(t *testing.T) {
	out := render.FormatWriteup("hello world", 5, render.Plain)
	if out != "hello\nworld" {
		t.Fatalf("%q", out)
	}
}

func TestFormatWriteupCodeFence(t *testing.T) {
	md := "before\n```\nline1\nline2\n```\nafter"
	out := render.FormatWriteup(md, 0, render.Plain)
	if !strings.Contains(out, "line1") || strings.Contains(out, "```") {
		t.Fatalf("%q", out)
	}
}