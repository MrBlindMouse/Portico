package portico

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestEveryRunsUntilCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var n atomic.Int32
	done := make(chan struct{})
	go func() {
		Every(ctx, 10*time.Millisecond, func() error {
			n.Add(1)
			return nil
		})
		close(done)
	}()
	deadline := time.After(200 * time.Millisecond)
	for n.Load() < 1 {
		select {
		case <-deadline:
			t.Fatalf("no ticks, n=%d", n.Load())
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
	before := n.Load()
	cancel()
	<-done
	time.Sleep(30 * time.Millisecond)
	if n.Load()-before > 1 {
		t.Fatalf("ticks after cancel: before=%d after=%d", before, n.Load())
	}
}

func TestEveryPanicsNilContext(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	Every(nil, time.Millisecond, func() error { return nil })
}

func TestEveryPanicsNonPositiveInterval(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	Every(ctx, 0, func() error { return nil })
}

func TestEveryContinuesOnSubmitError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var n atomic.Int32
	go Every(ctx, 10*time.Millisecond, func() error {
		n.Add(1)
		if n.Load() == 1 {
			return errTestSubmit
		}
		return nil
	})
	deadline := time.After(200 * time.Millisecond)
	for n.Load() < 2 {
		select {
		case <-deadline:
			t.Fatalf("expected multiple ticks after error, n=%d", n.Load())
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
}

var errTestSubmit = errors.New("submit failed")
