package identity_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	gossh "golang.org/x/crypto/ssh"
)

func TestMemoryStoreRoundTrip(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := gossh.NewPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	id := identity.FromPublicKey(pub)
	store := identity.NewMemoryStore()
	if err := store.Set(id, "note", "hello"); err != nil {
		t.Fatal(err)
	}
	v, ok, err := store.Get(id, "note")
	if err != nil || !ok || v != "hello" {
		t.Fatalf("v=%v ok=%v err=%v", v, ok, err)
	}
}

func TestAnonymousPersistRejected(t *testing.T) {
	store := identity.NewMemoryStore()
	if err := store.Set(identity.Anonymous(), "k", "v"); err == nil {
		t.Fatal("expected error")
	}
}
