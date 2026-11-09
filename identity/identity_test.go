package identity_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
)

func TestFromAccountDisplay(t *testing.T) {
	idn := identity.FromAccount("alice")
	if idn.Account() != "alice" || idn.String() != "alice" {
		t.Fatalf("%+v", idn)
	}
	if idn.Fingerprint() == "" {
		t.Fatal("expected persistence fingerprint")
	}
}

func TestLoginAliasIsDisplayOnly(t *testing.T) {
	guest := identity.Anonymous().WithLogin("johan")
	if guest.String() != "johan" || guest.Nickname() != "" || guest.Fingerprint() != "" || !guest.Anonymous() {
		t.Fatalf("guest=%+v label=%q", guest, guest.String())
	}
	key := identity.ForFingerprint("abcdef0123456789").WithLogin("johan")
	if key.String() != "johan" || key.Anonymous() || key.Fingerprint() == "" {
		t.Fatalf("key=%+v label=%q", key, key.String())
	}
	if identity.FromAccount("alice").WithLogin("johan").String() != "alice" {
		t.Fatal("login overrode account")
	}
	if guest.WithNickname("ada").String() != "ada" {
		t.Fatal("login overrode nickname")
	}
	if identity.Anonymous().String() != "guest" {
		t.Fatal("missing login should stay guest")
	}
}
