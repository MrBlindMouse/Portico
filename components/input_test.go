package components_test

import (
	"reflect"
	"testing"

	"github.com/MrBlindMouse/Portico/components"
)

func TestParseFieldsKeepsRemainder(t *testing.T) {
	got := components.ParseFields("a b c", 2)
	want := []string{"a", "b c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
	if components.ParseFields("only", 2)[0] != "only" {
		t.Fatal(components.ParseFields("only", 2))
	}
	if components.ParseFields("", 2) != nil {
		t.Fatal("empty")
	}
}

func TestMaskedCountsRunes(t *testing.T) {
	if got := components.Masked("é"); got != "*" {
		t.Fatalf("got %q", got)
	}
	if got := components.Masked(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
