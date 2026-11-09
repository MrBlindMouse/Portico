package runtime

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"github.com/MrBlindMouse/Portico/render"
	"github.com/gliderlabs/ssh"
)

// SSHTerminal adapts a gliderlabs SSH session to Terminal.
type SSHTerminal struct {
	sess   ssh.Session
	mu     sync.Mutex
	input  *inputReader
	resize func(func(int, int))
}

// NewSSHTerminal wraps an SSH session.
func NewSSHTerminal(sess ssh.Session) *SSHTerminal {
	t := &SSHTerminal{sess: sess}
	t.input = newInputReader(bufio.NewReader(sess))
	t.input.echo = t.echo
	return t
}

func (t *SSHTerminal) WriteFrame(text string) error {
	return t.writeCleared(text, true)
}

// WriteGrid clears the screen and writes text without adding a trailing newline,
// so a full-height grid does not scroll before the caret is placed.
// ponytail: autowrap is off for the write. The bottom-right cell can still wrap on
// some terminals if the last row is exactly the width. Upgrade path is CUP for that cell.
func (t *SSHTerminal) WriteGrid(text string) error {
	return t.writeCleared(text, false)
}

func (t *SSHTerminal) writeCleared(text string, ensureNewline bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if ensureNewline && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	prefix := "\x1b[2J\x1b[H"
	if !ensureNewline {
		prefix = "\x1b[?7l" + prefix
		text += "\x1b[?7h"
	}
	_, err := io.WriteString(t.sess, prefix+text)
	return err
}

// SetSlot installs the caret rectangle. Nil restores the prompt row under the frame.
func (t *SSHTerminal) SetSlot(in *render.Input) { t.input.setSlot(in) }

// PaintSlot redraws the caret rectangle from the line buffer.
func (t *SSHTerminal) PaintSlot() { t.input.paintLine() }

// SlotActive reports whether a caret rectangle is installed.
func (t *SSHTerminal) SlotActive() bool { return t.input.hasSlot() }

// RejectSlot shows msg inside the caret rectangle until the next edit.
func (t *SSHTerminal) RejectSlot(msg string) { t.input.rejectSlot(msg) }

func (t *SSHTerminal) ReadInput() (string, error) {
	return t.input.readUserCommand()
}

func (t *SSHTerminal) LineEndReason() string { return t.input.LineEndReason() }

func (t *SSHTerminal) SubmitCSI() string { return t.input.SubmitCSI() }

func (t *SSHTerminal) SkippedCSI() string { return t.input.SkippedCSI() }

func (t *SSHTerminal) echo(p []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, _ = t.sess.Write(p)
}

// BeginInputLine prints the command prompt on the line after a scene frame.
func (t *SSHTerminal) BeginInputLine() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := io.WriteString(t.sess, "> ")
	return err
}

// CommitInputLine keeps the submitted command and moves ">" onto the next line.
func (t *SSHTerminal) CommitInputLine(line string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := io.WriteString(t.sess, "> "+line+"\r\n> ")
	return err
}

// RejectInput prints why a line was not run, then a new prompt.
func (t *SSHTerminal) RejectInput(msg string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := io.WriteString(t.sess, msg+"\r\n> ")
	return err
}

// SetRedraw installs the scene repaint used by Ctrl-L.
func (t *SSHTerminal) SetRedraw(fn func()) { t.input.redraw = fn }

// SetComplete installs scene-key completion used by Tab.
func (t *SSHTerminal) SetComplete(fn func(string) string) { t.input.complete = fn }

// RestoreInputLine reprints the prompt and any half-typed line.
// A key scene has no prompt, so a resize only keeps the scene that was just drawn.
func (t *SSHTerminal) RestoreInputLine() {
	if len(t.input.keys) > 0 {
		return
	}
	t.input.paintLine()
}

// SetKeys selects single-key reads. An empty list returns to line entry.
func (t *SSHTerminal) SetKeys(keys []string) {
	t.input.setKeys(keys)
}

// ClearRecall drops up/down lines. A scene change calls this so the next scene starts empty.
func (t *SSHTerminal) ClearRecall() { t.input.clearRecall() }

// WindowSize returns the PTY size reported at session start.
func (t *SSHTerminal) WindowSize() (width, height int, ok bool) {
	pty, _, has := t.sess.Pty()
	if !has || pty.Window.Width <= 0 || pty.Window.Height <= 0 {
		return 0, 0, false
	}
	return pty.Window.Width, pty.Window.Height, true
}

func (t *SSHTerminal) OnResize(fn func(int, int)) {
	if fn == nil {
		return
	}
	_, winCh, ok := t.sess.Pty()
	if !ok {
		return
	}
	// ponytail: RunLoop already renders once; an eager fn() here doubled ESC [2J on
	// connect and prompted terminals to flood DA/mode replies into ReadInput.
	go func() {
		for win := range winCh {
			fn(win.Width, win.Height)
		}
	}()
}

func (t *SSHTerminal) Close() error { return t.sess.Close() }
