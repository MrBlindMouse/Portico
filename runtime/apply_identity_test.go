package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
)

func TestApplyReserveNickname(t *testing.T) {
	store := identity.NewMemoryStore()
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	h := testkit.NewWithStore(exp, store)
	ap := h.Applier()
	idn := identity.ForFingerprint("testfp")
	sess := h.SessionWithIdentity("s1", idn, "home")
	_, err := ap.Apply(sess, exp, []experience.Effect{experience.ReserveNicknameEffect("alice")})
	if err != nil {
		t.Fatal(err)
	}
	if sess.Identity().Nickname() != "alice" {
		t.Fatalf("nick=%q", sess.Identity().Nickname())
	}
}

func TestApplyRegisterAccount(t *testing.T) {
	accounts := runtime.NewAccountStore()
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	h := testkit.New(exp)
	ap := h.Applier()
	ap.SetAccountStore(accounts)
	sess := h.Session("s1", "home")
	_, err := ap.Apply(sess, exp, []experience.Effect{experience.RegisterAccountEffect("bob", "secret")})
	if err != nil {
		t.Fatal(err)
	}
	ok, err := accounts.Verify("bob", "secret")
	if err != nil || !ok {
		t.Fatal("expected registered user")
	}
}
