package runtime

import (
	"bufio"
	"io"
	"strings"
	"sync"

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
	t.mu.Lock()
	defer t.mu.Unlock()
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	_, err := io.WriteString(t.sess, "\x1b[2J\x1b[H"+text)
	return err
}

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
