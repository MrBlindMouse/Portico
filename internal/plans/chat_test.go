package plans_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/internal/plans"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestChatBuildsWithItsOwnRoot(t *testing.T) {
	exp, err := plans.Chat().Build()
	if err != nil {
		t.Fatal(err)
	}
	if exp.Root() != "chat" {
		t.Fatalf("root %q", exp.Root())
	}
}

func TestChatSendPostsMessage(t *testing.T) {
	exp, err := plans.Chat().Root("chat").Build()
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
	exp, err := plans.Guestbook().Build()
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
