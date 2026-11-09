package identity_test

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	gossh "golang.org/x/crypto/ssh"
)

func TestFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := identity.OpenFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := gossh.NewPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	id := identity.FromPublicKey(pub)
	if err := store.Set(id, "note", "persisted"); err != nil {
		t.Fatal(err)
	}
	store2, err := identity.OpenFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	v, ok, err := store2.Get(id, "note")
	if err != nil || !ok || v != "persisted" {
		t.Fatalf("v=%v ok=%v err=%v", v, ok, err)
	}
	if err := store.ReserveNickname(id, "alice"); err != nil {
		t.Fatal(err)
	}
	owner, ok, err := store.NicknameOwner("alice")
	if err != nil || !ok || owner.Fingerprint() != id.Fingerprint() {
		t.Fatalf("owner=%+v ok=%v err=%v", owner, ok, err)
	}
	path := filepath.Join(dir, "identities", id.Fingerprint()+".json")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
