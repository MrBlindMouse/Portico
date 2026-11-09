package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestApplyReserveNickname(t *testing.T) {
	store := identity.NewMemoryStore()
	ap := runtime.NewApplier(nil, store)
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	idn := identity.ForFingerprint("testfp")
	sess := runtime.NewSessionForTest("s1", idn, "home")
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
	ap := runtime.NewApplier(nil, identity.NewMemoryStore())
	ap.SetAccountStore(accounts)
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	sess := runtime.NewSessionForTest("s1", identity.Anonymous(), "home")
	_, err := ap.Apply(sess, exp, []experience.Effect{experience.RegisterAccountEffect("bob", "secret")})
	if err != nil {
		t.Fatal(err)
	}
	if !accounts.Verify("bob", "secret") {
		t.Fatal("expected registered user")
	}
}
