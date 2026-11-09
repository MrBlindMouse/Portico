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

func TestTruncateCountsRunes(t *testing.T) {
	if got := render.Truncate("héllo", 4); got != "hél…" {
		t.Fatalf("got %q", got)
	}
	if got := render.Pad("é", 3); got != "é  " {
		t.Fatalf("pad %q", got)
	}
}

func TestWrapAndAlign(t *testing.T) {
	if got := render.Wrap("hello world", 5); got != "hello\nworld" {
		t.Fatalf("wrap %q", got)
	}
	if got := render.Wrap("abcdef", 3); got != "abc\ndef" {
		t.Fatalf("hard wrap %q", got)
	}
	if got := render.AlignRight("ab", 5); got != "   ab" {
		t.Fatalf("right %q", got)
	}
	if got := render.AlignCenter("ab", 5); got != " ab  " {
		t.Fatalf("center %q", got)
	}
}

func TestListTableProgressSpinner(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	if got := render.List(items, 4, 2, 0, nil); got != "  d\n> e" {
		t.Fatalf("list %q", got)
	}
	if got := render.List([]string{"hello"}, 0, 0, 4, nil); got != "> h…" {
		t.Fatalf("list width %q", got)
	}
	got := render.Table([]string{"name", "n"}, [][]string{{"ada", "10"}}, nil)
	want := "name  n \n────  ──\nada   10"
	if got != want {
		t.Fatalf("table %q", got)
	}
	if got := render.Progress(1, 2, 6); got != "[##--] 50%" {
		t.Fatalf("progress %q", got)
	}
	if render.SpinnerFrame(2) != "-" || render.SpinnerFrame(-1) != `\` {
		t.Fatalf("spinner %q %q", render.SpinnerFrame(2), render.SpinnerFrame(-1))
	}
}

func TestPlainThemeLeavesText(t *testing.T) {
	if render.Plain.Bold("x") != "x" || render.Default.Inverse("x") == "x" {
		t.Fatal("theme roles")
	}
}
