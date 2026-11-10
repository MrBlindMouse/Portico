package portico

import (
	"context"
	"time"
)

// Every runs submit on each tick until ctx is cancelled. The site owns submit:
// build effects and call Host.Submit, SubmitToSpace, or SubmitAll.
//
// ctx must be non-nil. interval must be positive; both are programming errors and panic.
// submit errors are ignored; handle them inside submit. A tick in flight may run once after cancel.
func Every(ctx context.Context, interval time.Duration, submit func() error) {
	if ctx == nil {
		panic("portico: Every: nil context")
	}
	if interval <= 0 {
		panic("portico: Every: interval must be positive")
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = submit()
		}
	}
}
