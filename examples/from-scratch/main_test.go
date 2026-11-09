package main

import (
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestGreetShowsFlash(t *testing.T) {
	exp, err := build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), exp.Root())
	term := faketerm.New()
	eng.RegisterSession(sess)
	if err := eng.HandleInput(sess, "greet"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("you waved") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
}
