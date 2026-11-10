package plans_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
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
	h.Register(sess)
	h.MustDo(t, sess, "sign ships ahoy")
	h.AssertRenderedContains(t, sess, "guest: ships ahoy")
}

func TestChatNickAndLeave(t *testing.T) {
	exp, err := plans.Chat().Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	guest := h.Session("g", exp.Root())
	term := faketerm.New()
	h.Engine.RegisterSession(guest)
	if err := h.Engine.HandleInput(guest, "nick ada"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(guest, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("nickname needs a key or account") {
		t.Fatalf("frame=%q", term.LastFrame())
	}

	idn := identity.ForFingerprint("abcdef0123456789")
	sess := h.SessionWithIdentity("s", idn, exp.Root())
	h.Engine.RegisterSession(sess)
	if err := h.Engine.HandleInput(sess, "nick ada"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(sess, "send hello"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("ada: hello") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := h.Engine.HandleInput(sess, "leave"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("left chat") || term.Contains("- ada") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
}

func TestGuestbookPersistsLastSignature(t *testing.T) {
	exp, err := plans.Guestbook().Build()
	if err != nil {
		t.Fatal(err)
	}
	h := testkit.New(exp)
	idn := identity.ForFingerprint("abcdef0123456789")
	sess := h.SessionWithIdentity("s", idn, exp.Root())
	term := faketerm.New()
	h.Engine.RegisterSession(sess)
	if err := h.Engine.HandleInput(sess, "sign ships ahoy"); err != nil {
		t.Fatal(err)
	}
	v, ok, err := h.Store.Get(idn, "last_signature")
	if err != nil || !ok || v != "key-abcdef01: ships ahoy" {
		t.Fatalf("stored=%v ok=%v err=%v", v, ok, err)
	}
	if err := h.Engine.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("Your last: key-abcdef01: ships ahoy") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
}
