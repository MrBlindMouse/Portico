package asset

import (
	"os"
	"time"
)

// Poll watches a directory and calls onChange when mtime advances.
type Poll struct {
	Loader   *Loader
	Interval time.Duration
}

// Run blocks until ctx done, invoking onChange with Reload results.
func (p *Poll) Run(done <-chan struct{}, onChange func(map[string]string) error) {
	if p.Interval <= 0 {
		p.Interval = time.Second
	}
	var last time.Time
	if p.Loader != nil {
		if fi, err := os.Stat(p.Loader.dir); err == nil {
			last = fi.ModTime()
		}
	}
	for {
		select {
		case <-done:
			return
		default:
		}
		if fi, err := os.Stat(p.Loader.dir); err == nil && fi.ModTime().After(last) {
			last = fi.ModTime()
			assets, err := p.Loader.Reload()
			if err != nil {
				continue
			}
			_ = onChange(assets)
		}
		time.Sleep(p.Interval)
	}
}
