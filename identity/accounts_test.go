package identity_test

import (
	"path/filepath"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
)

func TestFileAccountStoreRegisterVerify(t *testing.T) {
	dir := t.TempDir()
	store, err := identity.OpenFileAccountStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Register("alice", "secret"); err != nil {
		t.Fatal(err)
	}
	ok, err := store.Verify("alice", "secret")
	if err != nil || !ok {
		t.Fatalf("verify: ok=%v err=%v", ok, err)
	}
	ok, err = store.Verify("alice", "wrong")
	if err != nil || ok {
		t.Fatalf("bad pass: ok=%v", ok)
	}
}

func TestFileAccountStorePersist(t *testing.T) {
	dir := t.TempDir()
	store, err := identity.OpenFileAccountStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	_ = store.Register("bob", "pass")
	if err := store.LinkKey("bob", "abcd"); err != nil {
		t.Fatal(err)
	}
	store2, err := identity.OpenFileAccountStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	rec, ok, err := store2.LookupByFingerprint("abcd")
	if err != nil || !ok || rec.Username != "bob" {
		t.Fatalf("%+v ok=%v err=%v", rec, ok, err)
	}
}

func TestFileAccountStoreExportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := identity.OpenFileAccountStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	_ = store.Register("u", "p")
	records, err := store.Export()
	if err != nil {
		t.Fatal(err)
	}
	store2, err := identity.OpenFileAccountStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store2.Import(records); err != nil {
		t.Fatal(err)
	}
	ok, _ := store2.Verify("u", "p")
	if !ok {
		t.Fatal("imported account should verify")
	}
}
