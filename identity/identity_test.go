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
