package plans_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/faketerm"
	"github.com/MrBlindMouse/Portico/internal/plans"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestLandingCommandsAreSorted(t *testing.T) {
	exp, err := plans.Landing(map[string]components.SceneID{
		"zoo":   "zoo",
		"alpha": "alpha",
	}).Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), exp.Root())
	term := faketerm.New()
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	frame := term.LastFrame()
	if strings.Index(frame, "alpha") > strings.Index(frame, "zoo") {
		t.Fatalf("commands out of order: %q", frame)
	}
}

func TestGuestbookSpectatorSeesSign(t *testing.T) {
	exp, err := plans.Guestbook().Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	s1 := runtime.NewSessionForTest("s1", identity.Anonymous(), exp.Root())
	s2 := runtime.NewSessionForTest("s2", identity.Anonymous(), exp.Root())
	t2 := faketerm.New()
	eng.RegisterSession(s1)
	eng.RegisterSession(s2)
	runtime.SetSessionTerminalForTest(s2, t2)
	if _, err := eng.Applier().Apply(s1, exp, exp.OnSessionStart()); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.Applier().Apply(s2, exp, exp.OnSessionStart()); err != nil {
		t.Fatal(err)
	}
	if err := eng.HandleInput(s1, "sign hello"); err != nil {
		t.Fatal(err)
	}
	if !t2.Contains("guest: hello") {
		t.Fatalf("spectator frame=%q", t2.LastFrame())
	}
}

func TestChatJoinBroadcastsPresence(t *testing.T) {
	exp, err := plans.Chat().Root("chat").Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	s1 := runtime.NewSessionForTest("s1", identity.Anonymous(), exp.Root())
	s2 := runtime.NewSessionForTest("s2", identity.Anonymous(), exp.Root())
	t1 := faketerm.New()
	eng.RegisterSession(s1)
	eng.RegisterSession(s2)
	runtime.SetSessionTerminalForTest(s1, t1)
	if err := eng.HandleInput(s1, "join-chat"); err != nil {
		t.Fatal(err)
	}
	if err := eng.HandleInput(s2, "join-chat"); err != nil {
		t.Fatal(err)
	}
	if strings.Count(t1.LastFrame(), "- guest") < 2 {
		t.Fatalf("frame=%q", t1.LastFrame())
	}
}

func TestLobbyReadyAndStart(t *testing.T) {
	exp, err := plans.Lobby().Root("lobby").Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), exp.Root())
	term := faketerm.New()
	eng.RegisterSession(sess)
	if err := eng.HandleInput(sess, "ready"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("join the lobby first") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "join"); err != nil {
		t.Fatal(err)
	}
	if err := eng.HandleInput(sess, "start"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("not everyone is ready") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "ready"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("Ready: guest") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "start"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("Started.") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "leave"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("left the lobby") || !term.Contains("No one in lobby") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if term.Contains("Ready: guest") {
		t.Fatalf("ready survived leave: %q", term.LastFrame())
	}
}

func TestFormStepsAndMenu(t *testing.T) {
	exp, err := plans.Form().Root("form").Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), "form")
	term := faketerm.New()
	eng.RegisterSession(sess)
	if err := eng.HandleInput(sess, "next hi"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("start the form first") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "start"); err != nil {
		t.Fatal(err)
	}
	if err := eng.HandleInput(sess, "next Ada"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("Enter message:") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "next hi there"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	frame := term.LastFrame()
	if !strings.Contains(frame, "name: Ada") || !strings.Contains(frame, "message: hi there") {
		t.Fatalf("frame=%q", frame)
	}
	menu := runtime.NewSessionForTest("m", identity.Anonymous(), "form-menu")
	menuTerm := faketerm.New()
	if err := eng.Render(menu, menuTerm); err != nil {
		t.Fatal(err)
	}
	if !menuTerm.Contains("go-form") || menuTerm.Contains("start —") {
		t.Fatalf("menu=%q", menuTerm.LastFrame())
	}
}

func TestBlogListsAndOpens(t *testing.T) {
	exp, err := plans.Blog().Root("blog-list").Asset("post-1.md", "hello post").Asset("notes.txt", "nope").Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), exp.Root())
	term := faketerm.New()
	eng.RegisterSession(sess)
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("1.md") || term.Contains("notes.txt") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "open"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("usage: open <post-id>") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "open 1.md"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("hello post") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
}

func TestEachPlanBuilds(t *testing.T) {
	cases := []struct {
		name string
		b    *installation.Builder
		root components.SceneID
	}{
		{"landing", plans.Landing(nil), "landing"},
		{"guestbook", plans.Guestbook(), "guestbook"},
		{"chat", plans.Chat(), "chat"},
		{"register", plans.Registration(), "register"},
		{"form", plans.Form(), "form"},
		{"blog", plans.Blog(), "blog-list"},
		{"lobby", plans.Lobby(), "lobby"},
	}
	for _, tc := range cases {
		exp, err := tc.b.Build()
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if exp.Root() != tc.root {
			t.Fatalf("%s root %q", tc.name, exp.Root())
		}
	}
}

func TestRegistrationCreatesAccount(t *testing.T) {
	exp, err := plans.Registration().Build()
	if err != nil {
		t.Fatal(err)
	}
	ap := runtime.NewApplier(nil, identity.NewMemoryStore())
	ap.SetAccountStore(runtime.NewAccountStore())
	eng := runtime.NewEngine(exp, ap)
	sess := runtime.NewSessionForTest("s", identity.Anonymous(), exp.Root())
	term := faketerm.New()
	eng.RegisterSession(sess)
	if err := eng.HandleInput(sess, "register"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("usage: register <username> <password>") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "register ../x secret"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("username:") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "register alice secret"); err != nil {
		t.Fatal(err)
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if !term.Contains("reconnect to sign in") {
		t.Fatalf("frame=%q", term.LastFrame())
	}
	if err := eng.HandleInput(sess, "register alice secret"); err == nil {
		t.Fatal("duplicate register should fail")
	}
	if err := eng.Render(sess, term); err != nil {
		t.Fatal(err)
	}
	if term.Contains("account created") {
		t.Fatalf("stale success flash: %q", term.LastFrame())
	}
}
