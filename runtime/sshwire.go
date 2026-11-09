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
	reader *bufio.Reader
	resize func(func(int, int))
}

// NewSSHTerminal wraps an SSH session.
func NewSSHTerminal(sess ssh.Session) *SSHTerminal {
	return &SSHTerminal{
		sess:   sess,
		reader: bufio.NewReader(sess),
	}
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
	line, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (t *SSHTerminal) OnResize(fn func(int, int)) {
	if fn == nil {
		return
	}
	pty, winCh, ok := t.sess.Pty()
	if !ok {
		return
	}
	fn(pty.Window.Width, pty.Window.Height)
	go func() {
		for win := range winCh {
			fn(win.Width, win.Height)
		}
	}()
}

func (t *SSHTerminal) Close() error { return t.sess.Close() }
