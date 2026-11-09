package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/runtime"
)

func TestAccountVerify(t *testing.T) {
	s := runtime.NewAccountStore()
	s.Register("alice", "secret")
	if !s.Verify("alice", "secret") {
		t.Fatal("expected valid password")
	}
	if s.Verify("alice", "wrong") {
		t.Fatal("expected invalid password")
	}
}
