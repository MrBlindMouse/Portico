package render_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/MrBlindMouse/Portico/render"
)

type slotRec struct {
	in render.Input
	ok bool
}

func (r *slotRec) PlaceInput(in render.Input) {
	r.in = in
	r.ok = true
}

func TestScreenDrawAndInput(t *testing.T) {
	s := render.NewScreen(10, 3)
	s.Draw(0, 0, "left")
	s.Draw(0, 5, "right")
	s.Input(9, -1, 100, ">")
	rec := &slotRec{}
	got := s.Show(rec)
	if got != "left right\n\n" {
		t.Fatalf("grid %q", got)
	}
	if !rec.ok || rec.in.Row != 2 || rec.in.Col != 0 || rec.in.Width != 10 || rec.in.Prompt != ">" {
		t.Fatalf("slot %+v", rec.in)
	}
}

func TestDrawWithinClipsColumn(t *testing.T) {
	s := render.NewScreen(10, 2)
	pane := s.DrawWithin(0, 0, 4, 0, "abcdefghij\nzz")
	s.DrawWithin(0, 5, 5, 0, "right")
	if got := s.String(); got != "abcd right\nzz" {
		t.Fatalf("grid %q", got)
	}
	if pane.Row != 0 || pane.Col != 0 || pane.Width != 4 || pane.Height != 2 {
		t.Fatalf("pane %+v", pane)
	}
}

type themeRec struct {
	slotRec
	th render.Theme
}

func (r *themeRec) Theme() render.Theme { return r.th }

func TestDrawListPaneWidthAndTheme(t *testing.T) {
	const paneW = 12
	s := render.NewScreen(20, 4)
	items := []string{"xxxxxxxxxx", "yyyyyyyyyy", "zzzzzzzzzz"}
	s.DrawList(1, 0, paneW, 0, items, 1, 0)
	plain := s.String()
	if strings.Contains(plain, "\x1b[") {
		t.Fatalf("grid string must not contain CSI: %q", plain)
	}
	lines := strings.Split(plain, "\n")
	if len(lines) < 2 {
		t.Fatal(plain)
	}
	gotSel := strings.TrimRight(lines[2], " ")
	if utf8.RuneCountInString(gotSel) != paneW {
		t.Fatalf("selected row width %d want %d line %q", utf8.RuneCountInString(gotSel), paneW, gotSel)
	}
	rec := &themeRec{th: render.Default}
	themed := s.Show(rec)
	if !strings.Contains(themed, "\x1b[7m") || !strings.Contains(themed, "> yyyyyyyyyy") {
		t.Fatalf("themed %q", themed)
	}
	if strings.Contains(strings.Split(themed, "\n")[2], "\x1b[") && utf8.RuneCountInString(gotSel) != paneW {
		t.Fatal("grid width must not include escape runes")
	}
}

func TestDrawListInverseStaysInColumn(t *testing.T) {
	s := render.NewScreen(24, 1)
	s.Draw(0, 0, "Left")
	s.DrawList(0, 10, 8, 1, []string{"Menu"}, 0, 0)
	got := s.Show(&themeRec{th: render.Default})
	inv := strings.Index(got, "\x1b[7m")
	left := strings.Index(got, "Left")
	if left < 0 || inv < 0 || left > inv {
		t.Fatalf("%q", got)
	}
	if strings.Contains(got[:inv], "\x1b[") {
		t.Fatalf("left pane styled: %q", got)
	}
	if !strings.Contains(got, "\x1b[7m> Menu  \x1b[0m") {
		t.Fatalf("span %q", got)
	}
}

func TestMarkRoleDoesNotShiftBorder(t *testing.T) {
	s := render.NewScreen(20, 1)
	s.Draw(0, 0, "Hi")
	s.MarkRole(0, 0, 2, render.RoleBold)
	s.Draw(0, 10, "│")
	plain := s.String()
	got := s.Show(&themeRec{th: render.Default})
	if strings.Index(plain, "│") != 10 {
		t.Fatalf("plain %q", plain)
	}
	stripped := strings.ReplaceAll(got, "\x1b[1m", "")
	stripped = strings.ReplaceAll(stripped, "\x1b[0m", "")
	if strings.Index(stripped, "│") != 10 {
		t.Fatalf("styled %q stripped %q", got, stripped)
	}
	if !strings.Contains(got, "\x1b[1mHi\x1b[0m") {
		t.Fatalf("bold %q", got)
	}
}

func TestDrawListTwoPanesStableWidth(t *testing.T) {
	s := render.NewScreen(24, 3)
	items := []string{"one", "two", "three", "four"}
	for _, sel := range []int{0, 2, 3} {
		s = render.NewScreen(24, 3)
		s.DrawList(0, 0, 10, 0, items, sel, 0)
		s.DrawWithin(0, 12, 8, 0, "aside")
		line := strings.Split(s.String(), "\n")[0]
		if len(line) < 17 || line[12:17] != "aside" {
			t.Fatalf("sel %d line %q", sel, line)
		}
	}
}

func TestBlockRectangles(t *testing.T) {
	box := render.BoxBody(2, 4, "Title", "a\nb", 20)
	if box.Row != 5 || box.Col != 6 || box.Width != 16 || box.Height != 2 {
		t.Fatalf("box %+v", box)
	}
	list := render.ListRow(3, 1, []string{"a", "b", "c", "d", "e"}, 4, 2, 6)
	if list.Row != 4 || list.Col != 1 || list.Width != 6 || list.Height != 1 {
		t.Fatalf("list %+v", list)
	}
	cell := render.TableCell(2, 4, []int{4, 2}, 0, 1, true)
	if cell.Row != 4 || cell.Col != 10 || cell.Width != 2 || cell.Height != 1 {
		t.Fatalf("cell %+v", cell)
	}
}
