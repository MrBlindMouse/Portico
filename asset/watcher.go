package asset

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Poll watches a directory and calls onChange when file names or contents change.
type Poll struct {
	Loader   *Loader
	Interval time.Duration
}

// Run blocks until done is closed, invoking onChange when file names or contents change.
// ponytail: each pass hashes every file in the directory. Switch to a filesystem watch if the directory gets large.
func (p *Poll) Run(done <-chan struct{}, onChange func(map[string]string) error) {
	if p == nil || p.Loader == nil {
		return
	}
	if p.Interval <= 0 {
		p.Interval = time.Second
	}
	last, _ := p.Loader.fingerprint()
	for {
		select {
		case <-done:
			return
		default:
		}
		cur, err := p.Loader.fingerprint()
		if err == nil && cur != last {
			last = cur
			assets, err := p.Loader.Reload()
			if err == nil {
				_ = onChange(assets)
			}
		}
		time.Sleep(p.Interval)
	}
}

// fingerprint is a content hash of every regular file. Directory mtime is not enough:
// writing an existing file does not change it.
func (l *Loader) fingerprint() (string, error) {
	if l == nil {
		return "", fmt.Errorf("portico: nil asset loader")
	}
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(l.dir, e.Name()))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\n%d\n", e.Name(), len(b))
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
