package runtime

import (
	"bufio"
	"bytes"
	"github.com/MrBlindMouse/Portico/components"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/MrBlindMouse/Portico/installation"
)

func TestIsTerminalProtocolLine(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"KITTY_DATA_END", true},
		{"KITTY_DATA_START", true},
		{"KITTY_CHUNKED", true},
		{"-- BEGIN CONDUCTOR --", true},
		{"OK", false},
		{"greet", false},
		{"go elsewhere", false},
		{"register alice secret", false},
		{"\x1b[?1u", true},
		{"[?1u", true},
		{"]1337;File=", true},
		{"_Gi=1", true},
		{"hello", false},
		{"", false},
		{
			"n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA==",
			true,
		},
	}
	for _, tt := range tests {
		if got := IsTerminalProtocolLine(tt.line); got != tt.want {
			t.Errorf("IsTerminalProtocolLine(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestMostlyNonPrintable(t *testing.T) {
	binaryish := strings.Repeat("A", 10) + "\x00\x01\x02\x03\x04\x05"
	if !IsTerminalProtocolLine(binaryish) {
		t.Fatal("expected binary line to be protocol")
	}
}

func TestReadUserLineStripsEscape(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "greet\n", "greet"},
		{"csi only", "\x1b[?1u\n", ""},
		{"kitty key before text", "\x1b[97u\x1b[1;3Dhello\n", "hello"},
		{"osc", "\x1b]1337;SetKittyColors\x07world\n", "world"},
		{"apc kitty graphics", "\x1b_Ga=1,i=1,s=1,v=1;aGVsbG8=\x1b\\ok\n", "ok"},
		{"crlf", "line\r\n", "line"},
		{"inline esc", "before\x1b[2Jafter\n", "beforeafter"},
		{"kitty enter", "guestbook\x1b[13u", "guestbook"},
		{"kitty shift enter", "guestbook\x1b[13;2u", "guestbook"},
		{"xterm enter", "guestbook\x1b[13~", "guestbook"},
		{"modifyOtherKeys enter", "chat\x1b[27;13;13~", "chat"},
		{"ss3 enter", "go\x1bOM", "go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := newInputReader(bufio.NewReader(strings.NewReader(tt.in)))
			got, err := ir.readUserLine()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestCRReturnsWithoutWaitingForNextByte(t *testing.T) {
	pr, pw := io.Pipe()
	ir := newInputReader(bufio.NewReader(pr))
	go func() {
		_, _ = pw.Write([]byte("chat\r"))
	}()
	done := make(chan string, 1)
	errc := make(chan error, 1)
	go func() {
		line, err := ir.readUserLine()
		if err != nil {
			errc <- err
			return
		}
		done <- line
	}()
	select {
	case err := <-errc:
		t.Fatal(err)
	case line := <-done:
		if line != "chat" {
			t.Fatalf("got %q", line)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("CR blocked until another byte (Enter waits for the next key)")
	}
	go func() {
		_, _ = pw.Write([]byte("q"))
		_ = pw.Close()
	}()
	line, err := ir.readUserLine()
	if err != nil {
		t.Fatal(err)
	}
	if line != "q" {
		t.Fatalf("key after Enter got %q", line)
	}
}

func TestReadUserCommandKittyThenUser(t *testing.T) {
	stream := "KITTY_DATA_START\nOK\n\x1b[?1u\nregister u p\n"
	ir := newInputReader(bufio.NewReader(strings.NewReader(stream)))
	got, err := ir.readUserCommand()
	if err != nil {
		t.Fatal(err)
	}
	if got != "register u p" {
		t.Fatalf("bootstrap noise absorbed; got %q", got)
	}
}

func TestCSISubmitsLine(t *testing.T) {
	if !csiSubmitsLine('u', []byte("13")) || !csiSubmitsLine('~', []byte("13")) {
		t.Fatal("enter should submit")
	}
	if !csiSubmitsLine('~', []byte("27;13;13")) {
		t.Fatal("modifyOtherKeys enter should submit")
	}
	if !csiSubmitsLine('u', []byte("57414")) {
		t.Fatal("keypad enter code should submit")
	}
	if csiSubmitsLine('u', []byte("97;2")) || csiSubmitsLine('m', []byte("13")) {
		t.Fatal("non-enter must not submit")
	}
}

func TestReadUserCommandBase64ChunkWithoutAPC(t *testing.T) {
	const b64 = "n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA=="
	ir := newInputReader(bufio.NewReader(strings.NewReader(b64 + "\nhello\n")))
	got, err := ir.readUserCommand()
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestLeadingWireNewlineKeepsPrompt(t *testing.T) {
	const b64 = "n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA=="
	stream := "\nKITTY_DATA_START\nOK\n" + b64 + "\nKITTY_DATA_END\nhello\n"
	var echoed []byte
	ir := newInputReader(bufio.NewReader(strings.NewReader(stream)))
	ir.echo = func(p []byte) { echoed = append(echoed, p...) }
	got, err := ir.readUserCommand()
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
	if bytes.Contains(echoed, []byte(">")) {
		t.Fatalf("reprinted prompt: %q", echoed)
	}
}

func TestSplitCRLFDoesNotOpenBlankPrompt(t *testing.T) {
	pr, pw := io.Pipe()
	var echoed []byte
	ir := newInputReader(bufio.NewReader(pr))
	ir.echo = func(p []byte) { echoed = append(echoed, p...) }
	go func() {
		_, _ = pw.Write([]byte("chat\r"))
		time.Sleep(30 * time.Millisecond)
		_, _ = pw.Write([]byte("\ngo\n"))
		_ = pw.Close()
	}()
	got, err := ir.readUserCommand()
	if err != nil || got != "chat" {
		t.Fatalf("first %q %v", got, err)
	}
	got, err = ir.readUserCommand()
	if err != nil || got != "go" {
		t.Fatalf("second %q %v", got, err)
	}
	if bytes.Contains(echoed, []byte(">")) {
		t.Fatalf("blank enter reprinted prompt: %q", echoed)
	}
}

func TestKittyBootstrapSilentEcho(t *testing.T) {
	const b64 = "n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA=="
	stream := "KITTY_DATA_START\nOK\n" + b64 + "\nH4sI\nKITTY_DATA_END\nhello\n"
	var echoed []byte
	ir := newInputReader(bufio.NewReader(strings.NewReader(stream)))
	ir.echo = func(p []byte) { echoed = append(echoed, p...) }
	got, err := ir.readUserCommand()
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
	echo := string(echoed)
	if strings.Contains(echo, "KITTY_") || strings.Contains(echo, "OK") || strings.Contains(echo, "H4sI") || strings.Contains(echo, b64[:32]) {
		t.Fatalf("echoed kitten wire: %q", echo)
	}
	if !strings.Contains(echo, "hello") {
		t.Fatalf("user line not echoed: %q", echo)
	}
}

func TestBase64ChunkSilentWithoutMarkers(t *testing.T) {
	const b64 = "n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA=="
	var echoed []byte
	ir := newInputReader(bufio.NewReader(strings.NewReader(b64 + "\nping\n")))
	ir.echo = func(p []byte) { echoed = append(echoed, p...) }
	got, err := ir.readUserCommand()
	if err != nil {
		t.Fatal(err)
	}
	if got != "ping" {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(string(echoed), b64[:24]) {
		t.Fatalf("echoed base64: %q", echoed)
	}
}

func TestITermConductorSilent(t *testing.T) {
	const b64 = "n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA=="
	stream := "-- BEGIN CONDUCTOR --\n" + b64 + "\nguestbook\n"
	var echoed []byte
	ir := newInputReader(bufio.NewReader(strings.NewReader(stream)))
	ir.echo = func(p []byte) { echoed = append(echoed, p...) }
	got, err := ir.readUserCommand()
	if err != nil {
		t.Fatal(err)
	}
	if got != "guestbook" {
		t.Fatalf("got %q", got)
	}
	echo := string(echoed)
	if strings.Contains(echo, "CONDUCTOR") || strings.Contains(echo, b64[:24]) {
		t.Fatalf("echoed conductor: %q", echo)
	}
}

func TestLineEdit(t *testing.T) {
	var echoed []byte
	ir := newInputReader(bufio.NewReader(strings.NewReader("ab\x7fc\x08\r")))
	ir.echo = func(p []byte) { echoed = append(echoed, p...) }
	got, err := ir.readUserLine()
	if err != nil {
		t.Fatal(err)
	}
	if got != "a" {
		t.Fatalf("got %q", got)
	}
	if !bytes.HasSuffix(echoed, []byte("\r\n")) {
		t.Fatalf("submit did not echo CR LF: %q", echoed)
	}

	ir = newInputReader(bufio.NewReader(strings.NewReader("abc\x15x\n")))
	got, err = ir.readUserLine()
	if err != nil || got != "x" {
		t.Fatalf("ctrl-u got %q err %v", got, err)
	}

	ir = newInputReader(bufio.NewReader(strings.NewReader("hello world\x17\n")))
	got, err = ir.readUserLine()
	if err != nil || got != "hello " {
		t.Fatalf("ctrl-w got %q err %v", got, err)
	}

	ir = newInputReader(bufio.NewReader(strings.NewReader("é\x7fz\n")))
	got, err = ir.readUserLine()
	if err != nil || got != "z" {
		t.Fatalf("rune erase got %q err %v", got, err)
	}

	echoed = nil
	ir = newInputReader(bufio.NewReader(strings.NewReader("\rchat\n")))
	ir.echo = func(p []byte) { echoed = append(echoed, p...) }
	got, err = ir.readUserCommand()
	if err != nil || got != "chat" {
		t.Fatalf("enter prompt got %q err %v", got, err)
	}
	if !bytes.Equal(echoed, []byte("\r\n\r> \x1b[Kchat\r\n")) {
		t.Fatalf("enter prompt echo %q", echoed)
	}

	ir = newInputReader(bufio.NewReader(strings.NewReader("\x04")))
	if _, err = ir.readUserLine(); err != io.EOF {
		t.Fatalf("ctrl-d err %v", err)
	}

	echoed = nil
	ir = newInputReader(bufio.NewReader(strings.NewReader("nope\x03")))
	ir.echo = func(p []byte) { echoed = append(echoed, p...) }
	if _, err = ir.readUserLine(); err != io.EOF {
		t.Fatalf("ctrl-c err %v", err)
	}
	if !bytes.Equal(echoed, []byte("nope^C\r\n")) {
		t.Fatalf("ctrl-c echo %q", echoed)
	}
}

func TestLineEditorKeys(t *testing.T) {
	line := func(in string) string {
		t.Helper()
		ir := newInputReader(bufio.NewReader(strings.NewReader(in)))
		got, err := ir.readUserLine()
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	if got := line("ab\x1b[Dc\n"); got != "acb" {
		t.Fatalf("left insert %q", got)
	}
	if got := line("ab\x1b[Hc\n"); got != "cab" {
		t.Fatalf("home insert %q", got)
	}
	if got := line("ab\x1b[D\x1b[3~\n"); got != "a" {
		t.Fatalf("delete forward %q", got)
	}
	if got := line("a\x1az\x1c\n"); got != "az" {
		t.Fatalf("swallowed controls %q", got)
	}
	if got := line("ab\x1b[200~c\nd\x1b[201~e\n"); got != "abc de" {
		t.Fatalf("paste %q", got)
	}

	ir := newInputReader(bufio.NewReader(strings.NewReader("one\n\x1b[A\n")))
	got, err := ir.readUserCommand()
	if err != nil || got != "one" {
		t.Fatalf("history seed %q %v", got, err)
	}
	got, err = ir.readUserCommand()
	if err != nil || got != "one" {
		t.Fatalf("history up %q %v", got, err)
	}

	ir = newInputReader(bufio.NewReader(strings.NewReader("ch\t\n")))
	ir.complete = func(prefix string) string {
		if prefix == "ch" {
			return "chat"
		}
		return prefix
	}
	got, err = ir.readUserLine()
	if err != nil || got != "chat" {
		t.Fatalf("tab %q %v", got, err)
	}

	redraws := 0
	ir = newInputReader(bufio.NewReader(strings.NewReader("\x0chi\n")))
	ir.redraw = func() { redraws++ }
	got, err = ir.readUserLine()
	if err != nil || got != "hi" || redraws != 1 {
		t.Fatalf("ctrl-l got %q redraws %d err %v", got, redraws, err)
	}
}

func TestReadKeyIgnoresUnbound(t *testing.T) {
	ir := newInputReader(bufio.NewReader(strings.NewReader("xn\x1b[A")))
	ir.setKeys([]string{"n", "up"})
	got, err := ir.readUserCommand()
	if err != nil || got != "n" {
		t.Fatalf("got %q err %v", got, err)
	}
	got, err = ir.readUserCommand()
	if err != nil || got != "up" {
		t.Fatalf("arrow got %q err %v", got, err)
	}
}

func TestReadKeyDrainsKittyBase64BeforeMenuKey(t *testing.T) {
	const b64 = "n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA=="
	stream := "KITTY_DATA_START\nOK\n" + b64 + "\nKITTY_DATA_END\ng"
	ir := newInputReader(bufio.NewReader(strings.NewReader(stream)))
	ir.setKeys([]string{"b", "c", "f", "g", "l", "r"})
	got, err := ir.readUserCommand()
	if err != nil {
		t.Fatal(err)
	}
	if got != "g" {
		t.Fatalf("got %q, want g after wire drain", got)
	}
}

func TestReadKeyDrainsBareBase64BeforeMenuKey(t *testing.T) {
	const b64 = "n0v4uazhRRJ+tvBzET+X8HNZw3sg/Gzh5yJ+LuHnsnb8fAf+L8BHCz4W4WMJPpY1EIlqKAvVQFSoCbGb+O5rIFd6J4l59/fu793fu78f9u//HwAzKeTyAH4BAA=="
	ir := newInputReader(bufio.NewReader(strings.NewReader(b64 + "\nc")))
	ir.setKeys([]string{"b", "c", "f", "g", "l", "r"})
	got, err := ir.readUserCommand()
	if err != nil {
		t.Fatal(err)
	}
	if got != "c" {
		t.Fatalf("got %q, want c after bare base64 line", got)
	}
}

func TestCompleteSceneInput(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"chat": "go", "changelog": "go", "guestbook": "go"},
		}).
		Action("go", components.Action{Run: func(components.ActionContext) []components.Effect { return nil }}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if got := completeSceneInput(exp, "home", "c"); got != "cha" {
		t.Fatalf("shared prefix %q", got)
	}
	if got := completeSceneInput(exp, "home", "gue"); got != "guestbook" {
		t.Fatalf("unique %q", got)
	}
	if got := completeSceneInput(exp, "home", ""); got != "" {
		t.Fatalf("empty %q", got)
	}
}

func TestPaintLineRestoresCursor(t *testing.T) {
	got := paintLineBytes([]byte("chat"), 2)
	if !bytes.Equal(got, []byte("\r> chat\x1b[K\b\b")) {
		t.Fatalf("%q", got)
	}
}

func TestReadUserLineNonEnterKittyKeyWaitsForNewline(t *testing.T) {
	ir := newInputReader(bufio.NewReader(strings.NewReader("hello\x1b[97u\n")))
	got, err := ir.readUserLine()
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
}
