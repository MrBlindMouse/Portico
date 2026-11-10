package main

import (
	"testing"

	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestEveryPlanFromLanding(t *testing.T) {
	exp, err := build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	h.Applier().SetAccountStore(runtime.NewAccountStore())
	term := faketerm.New()

	chat := h.Session("chat", exp.Root())
	h.Engine.RegisterSession(chat)
	if err := h.Engine.HandleInput(chat, "chat"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(chat, "send hello there"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(chat, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: hello there") {
		t.Fatalf("chat frame=%q", term.LastFrame())
	}

	book := h.Session("book", exp.Root())
	h.Engine.RegisterSession(book)
	if err := h.Engine.HandleInput(book, "guestbook"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(book, "sign ships ahoy"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(book, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: ships ahoy") {
		t.Fatalf("guestbook frame=%q", term.LastFrame())
	}

	reg := h.Session("reg", exp.Root())
	h.Engine.RegisterSession(reg)
	if err := h.Engine.HandleInput(reg, "register"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(reg, "register alice secret"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(reg, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("reconnect to sign in") {
		t.Fatalf("register frame=%q", term.LastFrame())
	}

	form := h.Session("form", exp.Root())
	h.Engine.RegisterSession(form)
	if err := h.Engine.HandleInput(form, "form"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(form, "start"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(form, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("Enter name:") {
		t.Fatalf("form frame=%q", term.LastFrame())
	}

	lobby := h.Session("lobby", exp.Root())
	h.Engine.RegisterSession(lobby)
	if err := h.Engine.HandleInput(lobby, "lobby"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(lobby, "join"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(lobby, "ready"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(lobby, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("Ready: guest") {
		t.Fatalf("lobby frame=%q", term.LastFrame())
	}

	blog := h.Session("blog", exp.Root())
	h.Engine.RegisterSession(blog)
	if err := h.Engine.HandleInput(blog, "blog"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(blog, "open 1.md"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(blog, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("hello post") {
		t.Fatalf("blog frame=%q", term.LastFrame())
	}
}
