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

func TestMemoryStoreReReserveReleasesOldName(t *testing.T) {
	store := identity.NewMemoryStore()
	id := identity.ForFingerprint("testfp")
	if err := store.ReserveNickname(id, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := store.ReserveNickname(id, "ada2"); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store.NicknameOwner("alice"); err != nil || ok {
		t.Fatalf("old name still held ok=%v err=%v", ok, err)
	}
	owner, ok, err := store.NicknameOwner("ada2")
	if err != nil || !ok || owner.Fingerprint() != id.Fingerprint() {
		t.Fatalf("owner=%+v ok=%v err=%v", owner, ok, err)
	}
	if err := store.ReserveNickname(id, "ada2"); err != nil {
		t.Fatal(err)
	}
}

func TestAnonymousPersistRejected(t *testing.T) {
	store := identity.NewMemoryStore()
	if err := store.Set(identity.Anonymous(), "k", "v"); err == nil {
		t.Fatal("expected error")
	}
}
