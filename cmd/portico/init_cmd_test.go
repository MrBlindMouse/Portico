package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCmdStdoutHints(t *testing.T) {
	dir := t.TempDir()
	modRoot := filepath.Join(dir, "site")
	if err := os.MkdirAll(modRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	// go mod init + go get need a module; init creates it via -module.
	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = wOut
	os.Stderr = devNull
	defer func() {
		_ = devNull.Close()
		os.Stdout = oldOut
		os.Stderr = oldErr
	}()

	done := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		done <- buf.Bytes()
	}()

	initCmd([]string{"-dir", modRoot, "-module", "example.com/init-hint-test"})
	_ = wOut.Close()
	out := <-done

	if !bytes.Contains(out, []byte("go run github.com/MrBlindMouse/Portico/cmd/portico generate")) {
		t.Fatalf("stdout missing generate hint: %q", out)
	}
	if bytes.Contains(out, []byte("Next: portico generate")) {
		t.Fatalf("stdout still has bare portico generate: %q", out)
	}
}
