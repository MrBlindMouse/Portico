package experience_test

import (
	"reflect"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
)

func TestParseFieldsKeepsRemainder(t *testing.T) {
	got := experience.ParseFields("a b c", 2)
	want := []string{"a", "b c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
	if experience.ParseFields("only", 2)[0] != "only" {
		t.Fatal(experience.ParseFields("only", 2))
	}
	if experience.ParseFields("", 2) != nil {
		t.Fatal("empty")
	}
}

func TestMaskedCountsRunes(t *testing.T) {
	if got := experience.Masked("é"); got != "*" {
		t.Fatalf("got %q", got)
	}
	if got := experience.Masked(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
