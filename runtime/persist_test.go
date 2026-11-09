package runtime_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
	gossh "golang.org/x/crypto/ssh"
)

func TestPersistEffect(t *testing.T) {
	store := identity.NewMemoryStore()
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Action("save", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{ctx.Persist("x", 42)}
			},
		}).
		Build()
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, store))
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := gossh.NewPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	idn := identity.FromPublicKey(pub)
	sess := runtime.NewSessionForTest("s", idn, "home")
	sess.SetIdentityForTest(idn)
	act, _ := exp.Action("save")
	ctx := runtime.NewActionContextForTest(sess, exp, eng)
	_, err = eng.Applier().Apply(sess, exp, act.Run(ctx))
	if err != nil {
		t.Fatal(err)
	}
	v, ok, err := store.Get(idn, "x")
	if err != nil || !ok || v != float64(42) && v != 42 {
		t.Fatalf("v=%v ok=%v err=%v", v, ok, err)
	}
}
