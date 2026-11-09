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

func TestReconnectPersist(t *testing.T) {
	store := identity.NewMemoryStore()
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string {
				v, ok := ctx.Persisted("token")
				if !ok {
					return "no"
				}
				if s, ok := v.(string); ok && s == "abc" {
					return "yes"
				}
				return "no"
			},
		}).
		Action("prime", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{ctx.Persist("token", "abc")}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub, _ := gossh.NewPublicKey(&key.PublicKey)
	idn := identity.FromPublicKey(pub)

	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, store))
	s1 := runtime.NewSessionForTest("1", idn, "home")
	act, _ := exp.Action("prime")
	ctx := runtime.NewActionContextForTest(s1, exp, eng)
	if _, err := eng.Applier().Apply(s1, exp, act.Run(ctx)); err != nil {
		t.Fatal(err)
	}
	s2 := runtime.NewSessionForTest("2", idn, "home")
	out, err := eng.RenderText(s2)
	if err != nil {
		t.Fatal(err)
	}
	if out != "yes" {
		t.Fatalf("render=%q want yes", out)
	}
}
