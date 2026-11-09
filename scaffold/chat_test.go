package scaffold_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/scaffold"
)

func TestChatSendPostsMessage(t *testing.T) {
	exp, err := scaffold.Chat().Root("chat").Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), exp.Root())
	term := faketerm.New()
	eng.RegisterSession(sess)
	if err := eng.HandleInput(sess, "send"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if term.Contains("guest:") {
		t.Fatal("bare send should not post")
	}
	if err := eng.HandleInput(sess, "send hello there"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
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
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), exp.Root())
	term := faketerm.New()
	eng.RegisterSession(sess)
	if err := eng.HandleInput(sess, "sign ships ahoy"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("guest: ships ahoy") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
}
