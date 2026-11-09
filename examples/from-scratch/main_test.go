package main

import (
	"testing"

	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestGreetShowsFlash(t *testing.T) {
	exp, err := build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s", exp.Root())
	term := faketerm.New()
	h.Engine.RegisterSession(sess)
	if err := h.Engine.HandleInput(sess, "greet"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("you waved") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
}
