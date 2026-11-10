package runtime

import (
	"bufio"
	"github.com/MrBlindMouse/Portico/components"
	"io"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

func TestSlotPaintScrollsInsideCell(t *testing.T) {
	in := render.Input{Row: 1, Col: 2, Width: 4, Prompt: ">"}
	got := string(slotPaintBytes(in, []byte("abcd"), 4, ""))
	want := "\x1b[2;3H>bcd\x1b[2;7H"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	empty := string(slotPaintBytes(in, nil, 0, ""))
	if empty != "\x1b[2;3H>   \x1b[2;4H" {
		t.Fatalf("empty %q", empty)
	}
}

func TestSlotInsertDoesNotUsePromptRow(t *testing.T) {
	var echoed strings.Builder
	ir := newInputReader(bufio.NewReader(strings.NewReader("")))
	ir.echo = func(p []byte) { echoed.Write(p) }
	ir.setSlot(&render.Input{Row: 0, Col: 0, Width: 4, Prompt: ">"})
	ir.insert([]byte("abcd"))
	got := echoed.String()
	if strings.Contains(got, "\r>") || !strings.Contains(got, "\x1b[1;1H>bcd") {
		t.Fatalf("%q", got)
	}
}

type slotTerm struct {
	grids   []string
	frames  []string
	slot    *render.Input
	painted int
}

func (t *slotTerm) WriteFrame(s string) error {
	t.frames = append(t.frames, s)
	return nil
}
func (t *slotTerm) WriteGrid(s string) error {
	t.grids = append(t.grids, s)
	return nil
}
func (t *slotTerm) SetSlot(in *render.Input) {
	if in == nil {
		t.slot = nil
		return
	}
	cp := *in
	t.slot = &cp
}
func (t *slotTerm) PaintSlot()                 { t.painted++ }
func (t *slotTerm) ReadInput() (string, error) { return "", io.EOF }
func (t *slotTerm) OnResize(func(int, int))    {}
func (t *slotTerm) Close() error               { return nil }

func TestRenderPlacesInput(t *testing.T) {
	exp, err := installation.NewBuilder().Root("home").Scene("home", components.Scene{
		Render: func(ctx components.RenderContext) string {
			s := render.NewScreen(10, 3)
			s.Draw(0, 0, "left")
			s.Draw(0, 5, "right")
			s.Input(2, 0, 4, ">")
			return s.Show(ctx)
		},
	}).Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, nil)
	sess := newSession("s", identity.Anonymous(), "home")
	term := &slotTerm{}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if len(term.grids) != 1 || !strings.Contains(term.grids[0], "left right") {
		t.Fatalf("grids %#v", term.grids)
	}
	if len(term.frames) != 0 {
		t.Fatalf("frames %#v", term.frames)
	}
	if term.slot == nil || term.slot.Row != 2 || term.slot.Col != 0 || term.slot.Width != 4 || term.painted != 1 {
		t.Fatalf("slot %+v painted %d", term.slot, term.painted)
	}
	in, ok := sess.inputSlot()
	if !ok || in.Width != 4 {
		t.Fatalf("session slot %+v %v", in, ok)
	}
}

func TestKeysSceneRejectsPlacedInput(t *testing.T) {
	exp, err := installation.NewBuilder().Root("home").Scene("home", components.Scene{
		Keys: true,
		Inputs: map[string]components.ActionID{
			"j": "down",
		},
		Render: func(ctx components.RenderContext) string {
			s := render.NewScreen(8, 2)
			s.Input(0, 0, 4, ">")
			return s.Show(ctx)
		},
	}).Action("down", components.Action{
		Run: func(components.ActionContext) []components.Effect { return nil },
	}).Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, nil)
	sess := newSession("s", identity.Anonymous(), "home")
	_, err = eng.RenderText(sess)
	if err == nil || !strings.Contains(err.Error(), "Keys") {
		t.Fatalf("err=%v", err)
	}
}
