package main

import (
	"testing"

	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestCommunityChatAndGuestbook(t *testing.T) {
	exp, err := build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s", exp.Root())
	term := faketerm.New()
	h.Engine.RegisterSession(sess)

	if err := h.Engine.HandleInput(sess, "chat"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(sess, "send hello there"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: hello there") {
		t.Fatalf("chat frame=%q", term.LastFrame())
	}

	sess2 := h.Session("g", exp.Root())
	h.Engine.RegisterSession(sess2)
	if err := h.Engine.HandleInput(sess2, "guestbook"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(sess2, "sign ships ahoy"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(sess2, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: ships ahoy") {
		t.Fatalf("guestbook frame=%q", term.LastFrame())
	}
}
