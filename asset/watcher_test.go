package asset

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPollSeesInPlaceEdit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := make(chan string, 1)
	p := &Poll{Loader: NewLoader(dir), Interval: 15 * time.Millisecond}
	done := make(chan struct{})
	defer close(done)
	go p.Run(done, func(m map[string]string) error {
		got <- m["a.txt"]
		return nil
	})

	time.Sleep(60 * time.Millisecond)
	select {
	case s := <-got:
		t.Fatalf("reloaded unchanged files: %q", s)
	default:
	}

	if err := os.WriteFile(path, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case s := <-got:
		if s != "two" {
			t.Fatalf("asset=%q", s)
		}
	case <-time.After(time.Second):
		t.Fatal("in-place edit was not reloaded")
	}
}
