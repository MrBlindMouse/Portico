package main

import (
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestCommunityChatAndGuestbook(t *testing.T) {
	exp, err := build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), exp.Root())
	term := faketerm.New()
	eng.RegisterSession(sess)

	if err := eng.HandleInput(sess, "chat"); err != nil {
		t.Fatal(err)
	}
	if err := eng.HandleInput(sess, "send hello there"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: hello there") {
		t.Fatalf("chat frame=%q", term.LastFrame())
	}

	sess2 := runtime.NewSessionForTest("g", identity.Anonymous(), exp.Root())
	eng.RegisterSession(sess2)
	if err := eng.HandleInput(sess2, "guestbook"); err != nil {
		t.Fatal(err)
	}
	if err := eng.HandleInput(sess2, "sign ships ahoy"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess2, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: ships ahoy") {
		t.Fatalf("guestbook frame=%q", term.LastFrame())
	}
}
