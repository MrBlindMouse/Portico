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

func TestFileStoreReReserveReleasesOldName(t *testing.T) {
	dir := t.TempDir()
	store, err := identity.OpenFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
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
	store2, err := identity.OpenFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := store2.NicknameOwner("alice"); err != nil || ok {
		t.Fatalf("reopen still holds alice ok=%v err=%v", ok, err)
	}
}

func TestFileStoreRejectsPathFingerprint(t *testing.T) {
	store, err := identity.OpenFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set(identity.ForFingerprint(".."), "k", "v"); err == nil {
		t.Fatal("expected invalid fingerprint")
	}
	if err := store.Set(identity.ForFingerprint("../outside"), "k", "v"); err == nil {
		t.Fatal("expected invalid fingerprint")
	}
}
