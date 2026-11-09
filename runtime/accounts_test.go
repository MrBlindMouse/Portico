package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/runtime"
)

func TestAccountVerify(t *testing.T) {
	s := runtime.NewAccountStore()
	_ = s.Register("alice", "secret")
	ok, err := s.Verify("alice", "secret")
	if err != nil || !ok {
		t.Fatal("expected valid password")
	}
	ok, _ = s.Verify("alice", "wrong")
	if ok {
		t.Fatal("expected invalid password")
	}
}
