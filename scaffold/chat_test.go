package scaffold_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/scaffold"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestChatSendPostsMessage(t *testing.T) {
	exp, err := scaffold.Chat().Root("chat").Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s", exp.Root())
	term := faketerm.New()
	h.Engine.RegisterSession(sess)
	if err := h.Engine.HandleInput(sess, "send"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if term.Contains("guest:") {
		t.Fatal("bare send should not post")
	}
	if err := h.Engine.HandleInput(sess, "send hello there"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: hello there") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
}

func TestGuestbookSignPostsMessage(t *testing.T) {
	exp, err := scaffold.Guestbook().Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	sess := h.Session("s", exp.Root())
	term := faketerm.New()
	h.Engine.RegisterSession(sess)
	if err := h.Engine.HandleInput(sess, "sign ships ahoy"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: ships ahoy") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
}
