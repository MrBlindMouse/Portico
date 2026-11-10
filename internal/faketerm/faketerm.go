package faketerm

import (
	"fmt"
	"strings"
	"sync"
)

// Terminal is an in-memory Terminal for tests.
type Terminal struct {
	mu       sync.Mutex
	frames   []string
	inputs   []string
	idx      int
	onResize func(int, int)
	closed   bool
}

// New creates a fake terminal with scripted inputs.
func New(inputs ...string) *Terminal {
	return &Terminal{inputs: inputs}
}

// Frames returns recorded frames.
func (t *Terminal) Frames() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	cp := make([]string, len(t.frames))
	copy(cp, t.frames)
	return cp
}

// LastFrame returns the most recent frame.
func (t *Terminal) LastFrame() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.frames) == 0 {
		return ""
	}
	return t.frames[len(t.frames)-1]
}

func (t *Terminal) WriteFrame(text string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return fmt.Errorf("terminal closed")
	}
	t.frames = append(t.frames, text)
	return nil
}

func (t *Terminal) ReadInput() (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return "", fmt.Errorf("terminal closed")
	}
	if t.idx >= len(t.inputs) {
		return "", fmt.Errorf("no more inputs")
	}
	v := t.inputs[t.idx]
	t.idx++
	return v, nil
}

func (t *Terminal) OnResize(fn func(int, int)) {
	t.onResize = fn
}

func (t *Terminal) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

// Contains checks the last frame for a substring.
func (t *Terminal) Contains(sub string) bool {
	return strings.Contains(t.LastFrame(), sub)
}
