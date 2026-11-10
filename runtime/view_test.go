package runtime

import (
	"fmt"
	"github.com/MrBlindMouse/Portico/components"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

func TestRenderSizeAndTheme(t *testing.T) {
	exp, err := installation.NewBuilder().Root("home").Scene("home", components.Scene{
		Render: func(ctx components.RenderContext) string {
			w, h := ctx.Size()
			return fmt.Sprintf("%d %d %s", w, h, ctx.Theme().Bold("x"))
		},
	}).Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, nil)
	sess := newSession("s", identity.Anonymous(), "home")
	out, err := eng.RenderText(sess)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%d %d %s", render.DefaultWidth, render.DefaultHeight, render.Default.Bold("x"))
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
	sess.setSize(100, 40)
	eng.SetTheme(render.Plain)
	out, err = eng.RenderText(sess)
	if err != nil {
		t.Fatal(err)
	}
	if out != "100 40 x" {
		t.Fatalf("got %q", out)
	}
}
