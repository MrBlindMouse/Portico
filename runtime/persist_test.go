package runtime_test

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/MrBlindMouse/Portico/components"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
	gossh "golang.org/x/crypto/ssh"
)

func TestPersistEffect(t *testing.T) {
	store := identity.NewMemoryStore()
	exp, _ := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{Render: func(components.RenderContext) string { return "" }}).
		Action("save", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{ctx.Persist("x", 42)}
			},
		}).
		Build()
	h := testkit.NewWithStore(exp, store)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := gossh.NewPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	idn := identity.FromPublicKey(pub)
	sess := h.SessionWithIdentity("s", idn, "home")
	sess.SetIdentityForTest(idn)
	act, _ := exp.Action("save")
	ctx := runtime.NewActionContextForTest(sess, exp, h.Engine)
	_, err = h.Applier().Apply(sess, exp, act.Run(ctx))
	if err != nil {
		t.Fatal(err)
	}
	v, ok, err := store.Get(idn, "x")
	if err != nil || !ok || v != float64(42) && v != 42 {
		t.Fatalf("v=%v ok=%v err=%v", v, ok, err)
	}
}
