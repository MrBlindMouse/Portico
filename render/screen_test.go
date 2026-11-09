package render_test

import (
	"testing"

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
