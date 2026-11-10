// Terminal protocol filtering for line-oriented SSH sessions.
//
// # Input pipeline (read this top to bottom)
//
//  1. Byte stream (this file, readUserLine / readUserCommand): assemble logical lines,
//     strip inline escapes, and recognize line ends. A line ends on \n, \r, or CSI Enter
//     (key codes in csiEnterKeys with final u/~, SS3 ESC O M). SSHTerminal echoes typed
//     bytes, erases on backspace, and prints "> " via InputPrompter after each frame
//     (remote ECHO is often off). Enter echoes CR LF and the next "> " is on that
//     new line. A placed input rectangle is painted with cursor addressing inside
//     that rectangle instead, and Enter does not echo CR LF. Arrows, Home, and End move inside the line. Up and down recall
//     history. Tab completes the scene command. Ctrl-L redraws. Bracketed paste
//     is inserted as text. Ctrl-U clears the line, Ctrl-W deletes the previous word,
//     Ctrl-Z and Ctrl-\ are ignored, Ctrl-C ends the session, and Ctrl-D on an
//     empty line ends it too.
//  2. Transport line filter (IsTerminalProtocolLine): discard whole lines that are
//     unambiguous wire noise (KITTY_* markers, long base64, escape-shaped remnants,
//     iTerm's "-- BEGIN CONDUCTOR --"). Those lines are assembled without local echo
//     when the bytes are already buffered (a kitten write) or while a bootstrap window
//     is open. Kitty's ssh kitten streams a tar from KITTY_DATA_START through
//     KITTY_DATA_END; OK and every other line in that window are absorbed. A line with
//     a space ends the window early and is kept as a command. Short tokens such as
//     "OK" outside that window may still reach scene eligibility.
//  3. Scene eligibility (dispatcher InputEligible + engine RunLoop): dispatch only when
//     the line matches the scene Inputs map. Any other line prints "no such command".
//
// RunLoop repeats step 2 as a guard; step 1 never returns obvious protocol lines from
// readUserCommand, but callers other than SSHTerminal may feed HandleInput directly.
//
// # Non-goals (ponytail ceilings)
//
// Not a VT emulator. Submit codes live in csiEnterKeys and the SS3 (ESC O M) check.
// Add another only when a real client sends it; do not grow scene or dispatch logic
// for terminal quirks.
package runtime

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/MrBlindMouse/Portico/render"
)

// historyCap bounds per-session command recall.
// ponytail: fixed slice, no dedupe beyond the previous entry; raise the cap if sessions outgrow it.
const historyCap = 32

// inputReader consumes escapes on the PTY byte stream.
type inputReader struct {
	mu             sync.Mutex
	r              *bufio.Reader
	echo           func([]byte) // optional local echo for SSH PTY (ECHO often off during full-screen redraw)
	redraw         func()       // Ctrl-L: repaint the scene; the reader repaints the prompt line after
	complete       func(string) string
	keys           map[string]struct{} // non-nil: single-key scene
	buf            []byte
	pos            int // byte offset, rune-aligned
	pasting        bool
	history        []string
	hist           int // index into history; len(history) means the live draft
	draft          []byte
	lineBufLen     int    // set before consuming escapes (for skip diagnostics)
	lastLineEnd    string // debug: how the last readUserLine ended (lf, cr, csi_enter, ss3_enter)
	lastSubmitCSI  string // debug: CSI that submitted, e.g. "27;13;13~"
	lastSkippedCSI string // debug: CSI seen while a line had content but did not submit
	slot           *render.Input
	notice         string // shown in the slot until the next edit; not part of the line
	silentLine     bool   // this line is wire; do not echo bytes or CR LF
	lastSilent     bool   // last completed line was silent; prompt stays put
	crPending      bool   // saw CR; a following LF is the rest of CRLF, not a new line
	kittyBootstrap bool   // ssh kitten payload after KITTY_DATA_START, until END
	conductor      bool   // iTerm it2ssh script after "-- BEGIN CONDUCTOR --"
	keyWire        []byte // key scene: line-buffered kitty/base64 before dispatch
}

func newInputReader(r *bufio.Reader) *inputReader {
	return &inputReader{r: r}
}

// LineEndReason reports how the last completed line was terminated (SSH debug).
func (ir *inputReader) LineEndReason() string { return ir.lastLineEnd }

// SubmitCSI reports the CSI sequence that submitted the last line, if any.
func (ir *inputReader) SubmitCSI() string { return ir.lastSubmitCSI }

// SkippedCSI reports the last CSI that did not submit while the in-progress line was non-empty.
func (ir *inputReader) SkippedCSI() string { return ir.lastSkippedCSI }

func (ir *inputReader) readUserLine() (string, error) {
	for {
		b, err := ir.r.ReadByte()
		if err != nil {
			if err == io.EOF && len(ir.buf) > 0 {
				return string(ir.buf), nil
			}
			return "", err
		}
		if !ir.pasting && ir.crPending {
			ir.crPending = false
			if b == '\n' {
				continue
			}
		}
		if ir.pasting && (b == '\n' || b == '\r') {
			if b == '\r' && ir.r.Buffered() > 0 {
				if peek, _ := ir.r.Peek(1); len(peek) > 0 && peek[0] == '\n' {
					_, _ = ir.r.ReadByte()
				}
			}
			ir.insertSpace()
			continue
		}
		switch b {
		case '\n':
			if len(ir.buf) == 0 && ir.peekWireFollows() {
				ir.silentLine = true
			}
			return ir.endLine("lf", "")
		case '\r':
			// Peek blocks until another byte exists. A lone CR (Enter) must finish
			// the line now; only swallow LF when it is already buffered (CRLF).
			swallowed := false
			if ir.r.Buffered() > 0 {
				if peek, _ := ir.r.Peek(1); len(peek) > 0 && peek[0] == '\n' {
					_, _ = ir.r.ReadByte()
					swallowed = true
				}
			}
			ir.crPending = !swallowed
			if len(ir.buf) == 0 && ir.peekWireFollows() {
				ir.silentLine = true
				ir.crPending = false
			}
			return ir.endLine("cr", "")
		case 0x03: // Ctrl-C
			ir.echoBytes([]byte{'^', 'C', '\r', '\n'})
			return "", io.EOF
		case 0x04: // Ctrl-D
			if len(ir.buf) == 0 {
				return "", io.EOF
			}
		case 0x08, 0x7f: // BS, DEL
			if !ir.silentLine {
				ir.eraseBefore()
			}
		case 0x0c: // Ctrl-L
			if ir.silentLine {
				break
			}
			if ir.redraw != nil {
				ir.redraw()
			}
			ir.paintLine()
		case 0x15: // Ctrl-U
			if !ir.silentLine {
				ir.eraseAll()
			}
		case 0x17: // Ctrl-W
			if !ir.silentLine {
				ir.eraseWord()
			}
		case 0x1a, 0x1c: // Ctrl-Z, Ctrl-\
			// Swallow. These must not suspend or kill the server process.
		case '\t':
			if !ir.silentLine {
				ir.completeLine()
			}
		case 0x1b:
			if ir.silentLine {
				_, err = ir.consumeEscape()
				if err != nil {
					return "", err
				}
				break
			}
			ir.lineBufLen = len(ir.buf)
			submit, err := ir.consumeEscape()
			if err != nil {
				return "", err
			}
			if submit && !ir.pasting {
				return ir.endLine(ir.lastLineEnd, ir.lastSubmitCSI)
			}
		default:
			if userLineByteOK(b) {
				if ir.shouldSilence(b) {
					ir.silentLine = true
					ir.appendWireByte(b)
				} else {
					ir.insert([]byte{b})
				}
			}
		}
	}
}

// shouldSilence reports whether b belongs on a wire line.
// Peeks only bytes already buffered, so a typed character never blocks.
func (ir *inputReader) shouldSilence(b byte) bool {
	if ir.silentLine || ir.kittyBootstrap {
		return true
	}
	if len(ir.buf) != 0 {
		return false
	}
	if ir.conductor {
		// The next line is more script, or the user. Only the script stays silent.
		if ir.peekBase64Chunk(b) || ir.peekEscapeShaped(b) || (b == '-' && ir.peekHasPrefix("- BEGIN CONDUCTOR --")) {
			return true
		}
		ir.conductor = false
		return false
	}
	switch {
	case b == 'K' && ir.peekHasPrefix("ITTY_"):
		if ir.peekHasPrefix("ITTY_DATA_START") {
			ir.kittyBootstrap = true
		}
		return true
	case b == '-' && ir.peekHasPrefix("- BEGIN CONDUCTOR --"):
		ir.conductor = true
		return true
	case ir.peekEscapeShaped(b), ir.peekBase64Chunk(b):
		return true
	default:
		return false
	}
}

// peekWireFollows reports a protocol dump already buffered after this newline.
func (ir *inputReader) peekWireFollows() bool {
	n := ir.r.Buffered()
	if n == 0 {
		return false
	}
	if n > 80 {
		n = 80
	}
	got, _ := ir.r.Peek(n)
	if bytes.HasPrefix(got, []byte("KITTY_")) || bytes.HasPrefix(got, []byte("-- BEGIN CONDUCTOR --")) {
		return true
	}
	line := got
	if i := bytesIndexAny(got); i >= 0 {
		line = got[:i]
	}
	s := string(line)
	return looksLikeEscapeLine(s) || looksLikeBase64Chunk(s)
}

func (ir *inputReader) peekHasPrefix(prefix string) bool {
	if ir.r.Buffered() < len(prefix) {
		return false
	}
	got, err := ir.r.Peek(len(prefix))
	return err == nil && string(got) == prefix
}

func (ir *inputReader) peekEscapeShaped(first byte) bool {
	if !strings.ContainsRune("]P_^X[", rune(first)) {
		return false
	}
	n := ir.r.Buffered()
	if n < 1 {
		return false
	}
	if n > 64 {
		n = 64
	}
	got, _ := ir.r.Peek(n)
	s := append([]byte{first}, got...)
	if i := bytesIndexAny(s); i >= 0 {
		s = s[:i]
	}
	return looksLikeEscapeLine(string(s))
}

func (ir *inputReader) peekBase64Chunk(first byte) bool {
	n := ir.r.Buffered()
	if n < 47 {
		return false
	}
	if n > 256 {
		n = 256
	}
	got, _ := ir.r.Peek(n)
	line := make([]byte, 0, n+1)
	line = append(line, first)
	for _, c := range got {
		if c == '\n' || c == '\r' {
			break
		}
		line = append(line, c)
	}
	return looksLikeBase64Chunk(string(line))
}

func bytesIndexAny(s []byte) int {
	for i, c := range s {
		if c == '\n' || c == '\r' {
			return i
		}
	}
	return -1
}

func (ir *inputReader) appendWireByte(b byte) {
	ir.mu.Lock()
	ir.buf = append(ir.buf, b)
	ir.pos = len(ir.buf)
	ir.mu.Unlock()
}

func (ir *inputReader) endLine(kind, csi string) (string, error) {
	ir.mu.Lock()
	line := string(ir.buf)
	ir.buf = ir.buf[:0]
	ir.pos = 0
	slotted := ir.slot != nil
	silent := ir.silentLine || ir.kittyBootstrap || ir.conductor
	ir.silentLine = false
	ir.lastSilent = silent
	ir.mu.Unlock()
	if !slotted && !silent {
		ir.echoBytes([]byte{'\r', '\n'})
	}
	ir.noteLineEnd(kind, csi)
	return line, nil
}

func (ir *inputReader) echoBytes(p []byte) {
	if ir.echo != nil && len(p) > 0 {
		ir.echo(p)
	}
}

func (ir *inputReader) snapshot() (buf []byte, pos int, slot *render.Input, notice string) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	var s *render.Input
	if ir.slot != nil {
		cp := *ir.slot
		s = &cp
	}
	return append([]byte(nil), ir.buf...), ir.pos, s, ir.notice
}

func (ir *inputReader) hasSlot() bool {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	return ir.slot != nil
}

func (ir *inputReader) setSlot(in *render.Input) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if in == nil {
		ir.slot = nil
		ir.notice = ""
		return
	}
	cp := *in
	ir.slot = &cp
}

func (ir *inputReader) rejectSlot(msg string) {
	ir.mu.Lock()
	ir.notice = msg
	ir.mu.Unlock()
	ir.paintLine()
}

func (ir *inputReader) clearNotice() {
	ir.notice = ""
}

// paintLine reprints the prompt line, or the placed caret rectangle.
// ponytail: one column per rune; wide glyphs can leave a cell. Upgrade path is wcwidth.
func (ir *inputReader) paintLine() {
	buf, pos, slot, notice := ir.snapshot()
	if slot != nil {
		ir.echoBytes(slotPaintBytes(*slot, buf, pos, notice))
		return
	}
	ir.echoBytes(paintLineBytes(buf, pos))
}

func paintLineBytes(buf []byte, pos int) []byte {
	out := make([]byte, 0, len(buf)+8)
	out = append(out, '\r', '>', ' ')
	out = append(out, buf...)
	out = append(out, '\x1b', '[', 'K')
	if pos < 0 || pos > len(buf) {
		pos = len(buf)
	}
	for i := 0; i < utf8.RuneCount(buf[pos:]); i++ {
		out = append(out, '\b')
	}
	return out
}

// slotPaintBytes fills the caret rectangle and parks the cursor inside it.
// notice is shown only while the edit buffer is empty.
// ponytail: the window is a rune slice of the editable columns. Wide glyphs can spill the cell.
func slotPaintBytes(in render.Input, buf []byte, pos int, notice string) []byte {
	width := in.Width
	if width < 1 {
		width = 1
	}
	prompt := []rune(in.Prompt)
	if len(prompt) > width {
		prompt = prompt[:width]
	}
	editable := width - len(prompt)
	var vis []rune
	cursor := len(prompt)
	if notice != "" && len(buf) == 0 {
		vis = []rune(notice)
		if len(vis) > editable {
			vis = vis[:editable]
		}
		cursor = len(prompt) + len(vis)
	} else if editable > 0 {
		runes := []rune(string(buf))
		if pos < 0 {
			pos = 0
		}
		if pos > len(buf) {
			pos = len(buf)
		}
		posR := utf8.RuneCount(buf[:pos])
		if posR > len(runes) {
			posR = len(runes)
		}
		start := 0
		if posR > editable {
			start = posR - editable
		}
		end := start + editable
		if end > len(runes) {
			end = len(runes)
			start = end - editable
			if start < 0 {
				start = 0
			}
		}
		vis = runes[start:end]
		cursor = len(prompt) + posR - start
	}
	if cursor > width {
		cursor = width
	}
	if cursor < 0 {
		cursor = 0
	}
	cells := make([]rune, width)
	for i := range cells {
		cells[i] = ' '
	}
	copy(cells, prompt)
	if len(prompt) < width {
		copy(cells[len(prompt):], vis)
	}
	out := make([]byte, 0, 32+width*2)
	out = append(out, cup(in.Row, in.Col)...)
	out = append(out, string(cells)...)
	out = append(out, cup(in.Row, in.Col+cursor)...)
	return out
}

func cup(row, col int) []byte {
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	return fmt.Appendf(nil, "\x1b[%d;%dH", row+1, col+1)
}

func (ir *inputReader) insert(p []byte) {
	if len(p) == 0 {
		return
	}
	ir.mu.Lock()
	ir.clearNotice()
	atEnd := ir.pos == len(ir.buf)
	slotted := ir.slot != nil
	if atEnd {
		ir.buf = append(ir.buf, p...)
		ir.pos = len(ir.buf)
		ir.mu.Unlock()
		if slotted {
			ir.paintLine()
			return
		}
		ir.echoBytes(p)
		return
	}
	n := make([]byte, 0, len(ir.buf)+len(p))
	n = append(n, ir.buf[:ir.pos]...)
	n = append(n, p...)
	n = append(n, ir.buf[ir.pos:]...)
	ir.buf = n
	ir.pos += len(p)
	ir.mu.Unlock()
	ir.paintLine()
}

func (ir *inputReader) insertSpace() {
	ir.mu.Lock()
	if ir.pos > 0 && ir.buf[ir.pos-1] == ' ' {
		ir.mu.Unlock()
		return
	}
	ir.mu.Unlock()
	ir.insert([]byte{' '})
}

func (ir *inputReader) setLine(s string) {
	ir.mu.Lock()
	ir.clearNotice()
	ir.buf = append(ir.buf[:0], s...)
	ir.pos = len(ir.buf)
	ir.mu.Unlock()
	ir.paintLine()
}

func (ir *inputReader) eraseBefore() {
	ir.mu.Lock()
	if ir.pos == 0 {
		ir.mu.Unlock()
		return
	}
	_, size := utf8.DecodeLastRune(ir.buf[:ir.pos])
	if size < 1 {
		size = 1
	}
	atEnd := ir.pos == len(ir.buf)
	slotted := ir.slot != nil
	ir.clearNotice()
	ir.buf = append(ir.buf[:ir.pos-size], ir.buf[ir.pos:]...)
	ir.pos -= size
	ir.mu.Unlock()
	if slotted || !atEnd {
		ir.paintLine()
		return
	}
	ir.echoBytes([]byte{'\b', ' ', '\b'})
}

func (ir *inputReader) eraseAll() {
	ir.mu.Lock()
	if len(ir.buf) == 0 {
		ir.mu.Unlock()
		return
	}
	atEnd := ir.pos == len(ir.buf)
	slotted := ir.slot != nil
	n := utf8.RuneCount(ir.buf)
	ir.clearNotice()
	ir.buf = ir.buf[:0]
	ir.pos = 0
	ir.mu.Unlock()
	if slotted || !atEnd {
		ir.paintLine()
		return
	}
	seq := make([]byte, 0, n*3)
	for i := 0; i < n; i++ {
		seq = append(seq, '\b', ' ', '\b')
	}
	ir.echoBytes(seq)
}

func (ir *inputReader) eraseWord() {
	for {
		ir.mu.Lock()
		if ir.pos == 0 {
			ir.mu.Unlock()
			return
		}
		r, _ := utf8.DecodeLastRune(ir.buf[:ir.pos])
		ir.mu.Unlock()
		if !unicode.IsSpace(r) {
			break
		}
		ir.eraseBefore()
	}
	for {
		ir.mu.Lock()
		if ir.pos == 0 {
			ir.mu.Unlock()
			return
		}
		r, _ := utf8.DecodeLastRune(ir.buf[:ir.pos])
		ir.mu.Unlock()
		if unicode.IsSpace(r) {
			break
		}
		ir.eraseBefore()
	}
}

func (ir *inputReader) deleteForward() {
	ir.mu.Lock()
	if ir.pos >= len(ir.buf) {
		ir.mu.Unlock()
		return
	}
	_, size := utf8.DecodeRune(ir.buf[ir.pos:])
	if size < 1 {
		size = 1
	}
	ir.clearNotice()
	ir.buf = append(ir.buf[:ir.pos], ir.buf[ir.pos+size:]...)
	ir.mu.Unlock()
	ir.paintLine()
}

func (ir *inputReader) moveLeft() {
	ir.mu.Lock()
	if ir.pos == 0 {
		ir.mu.Unlock()
		return
	}
	_, size := utf8.DecodeLastRune(ir.buf[:ir.pos])
	if size < 1 {
		size = 1
	}
	ir.pos -= size
	slotted := ir.slot != nil
	ir.mu.Unlock()
	if slotted {
		ir.paintLine()
		return
	}
	ir.echoBytes([]byte{'\b'})
}

func (ir *inputReader) moveRight() {
	ir.mu.Lock()
	if ir.pos >= len(ir.buf) {
		ir.mu.Unlock()
		return
	}
	_, size := utf8.DecodeRune(ir.buf[ir.pos:])
	if size < 1 {
		size = 1
	}
	chunk := append([]byte(nil), ir.buf[ir.pos:ir.pos+size]...)
	ir.pos += size
	slotted := ir.slot != nil
	ir.mu.Unlock()
	if slotted {
		ir.paintLine()
		return
	}
	ir.echoBytes(chunk)
}

func (ir *inputReader) moveHome() {
	ir.mu.Lock()
	if ir.pos == 0 {
		ir.mu.Unlock()
		return
	}
	ir.pos = 0
	ir.mu.Unlock()
	ir.paintLine()
}

func (ir *inputReader) moveEnd() {
	ir.mu.Lock()
	if ir.pos == len(ir.buf) {
		ir.mu.Unlock()
		return
	}
	ir.pos = len(ir.buf)
	ir.mu.Unlock()
	ir.paintLine()
}

func (ir *inputReader) pushHistory(line string) {
	if line == "" {
		return
	}
	if n := len(ir.history); n > 0 && ir.history[n-1] == line {
		ir.hist = n
		return
	}
	ir.history = append(ir.history, line)
	if len(ir.history) > historyCap {
		ir.history = append([]string(nil), ir.history[len(ir.history)-historyCap:]...)
	}
	ir.hist = len(ir.history)
	ir.draft = ir.draft[:0]
}

func (ir *inputReader) clearRecall() {
	ir.history = nil
	ir.hist = 0
	ir.draft = ir.draft[:0]
}

func (ir *inputReader) historyOlder() {
	if len(ir.history) == 0 || ir.hist == 0 {
		return
	}
	if ir.hist == len(ir.history) {
		ir.mu.Lock()
		ir.draft = append(ir.draft[:0], ir.buf...)
		ir.mu.Unlock()
	}
	ir.hist--
	ir.setLine(ir.history[ir.hist])
}

func (ir *inputReader) historyNewer() {
	if ir.hist >= len(ir.history) {
		return
	}
	ir.hist++
	if ir.hist == len(ir.history) {
		ir.setLine(string(ir.draft))
		return
	}
	ir.setLine(ir.history[ir.hist])
}

func (ir *inputReader) completeLine() {
	if ir.complete == nil || ir.pasting {
		return
	}
	ir.mu.Lock()
	if ir.pos != len(ir.buf) {
		ir.mu.Unlock()
		return
	}
	cur := string(ir.buf)
	ir.mu.Unlock()
	next := ir.complete(cur)
	if next != "" && next != cur {
		ir.setLine(next)
	}
}

// readUserCommand returns the next line that is not transport-level protocol noise.
func (ir *inputReader) readUserCommand() (string, error) {
	if len(ir.keys) > 0 {
		return ir.readKey()
	}
	for {
		line, err := ir.readUserLine()
		if err != nil {
			return "", err
		}
		line = normalizeUserLine(line)
		ir.noteWireWindow(line)
		if ir.discardTransportLine(line) {
			ir.promptAfterTransportDiscard()
			continue
		}
		if ir.lastSilent && line != "" && !ir.hasSlot() {
			// Typed through a wire window: the bytes were not echoed as they arrived.
			ir.echoBytes([]byte(line))
			ir.echoBytes([]byte("\r\n"))
		}
		ir.pushHistory(line)
		return line, nil
	}
}

const itermConductorBegin = "-- BEGIN CONDUCTOR --"

func (ir *inputReader) noteWireWindow(line string) {
	switch line {
	case "KITTY_DATA_START":
		ir.kittyBootstrap = true
	case "KITTY_DATA_END":
		ir.kittyBootstrap = false
	case itermConductorBegin:
		ir.conductor = true
	}
}

func (ir *inputReader) discardTransportLine(line string) bool {
	if line == "" {
		return true
	}
	if ir.kittyBootstrap {
		// A spaced line is someone typing through a stuck window. Keep it.
		if strings.ContainsRune(line, ' ') && !isKittyWireLine(line) {
			ir.kittyBootstrap = false
			return false
		}
		return true
	}
	if ir.conductor {
		if line == itermConductorBegin || IsTerminalProtocolLine(line) {
			return true
		}
		ir.conductor = false
		return false
	}
	return IsTerminalProtocolLine(line)
}

// promptAfterTransportDiscard reprints the prompt only when the discarded line moved the cursor.
// The repaint starts with CR, so a second prompt replaces the first instead of stacking "> >".
func (ir *inputReader) promptAfterTransportDiscard() {
	if ir.lastSilent {
		return
	}
	ir.paintLine()
}

func (ir *inputReader) setKeys(keys []string) {
	if len(keys) == 0 {
		ir.keys = nil
		return
	}
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	ir.keys = m
}

func (ir *inputReader) hasKey(s string) bool {
	_, ok := ir.keys[s]
	return ok
}

// readKey returns the next bound key. Unbound input is discarded.
// Kitty/iTerm bootstrap lines are buffered and dropped so wire bytes that
// match menu keys (common in base64) do not dispatch before the user acts.
func (ir *inputReader) readKey() (string, error) {
	for {
		b, err := ir.r.ReadByte()
		if err != nil {
			return "", err
		}
		switch b {
		case 0x03:
			ir.echoBytes([]byte{'^', 'C', '\r', '\n'})
			return "", io.EOF
		case 0x04:
			return "", io.EOF
		case 0x0c:
			if ir.redraw != nil {
				ir.redraw()
			}
		case '\n', '\r':
			if b == '\r' && ir.r.Buffered() > 0 {
				if peek, _ := ir.r.Peek(1); len(peek) > 0 && peek[0] == '\n' {
					_, _ = ir.r.ReadByte()
				}
			}
			if len(ir.keyWire) > 0 || ir.kittyBootstrap || ir.conductor {
				ir.finishKeyWireLine(string(ir.keyWire))
			}
		case 0x1b:
			if len(ir.keyWire) > 0 || ir.kittyBootstrap || ir.conductor {
				ir.keyWire = append(ir.keyWire, b)
				continue
			}
			name, err := ir.readKeyEscape()
			if err != nil {
				return "", err
			}
			if name != "" && !ir.pasting && ir.hasKey(name) {
				return name, nil
			}
		default:
			if b < 0x20 || b == 0x7f {
				continue
			}
			s, err := ir.readKeyRune(b)
			if err != nil {
				return "", err
			}
			if len(ir.keyWire) > 0 || ir.kittyBootstrap || ir.conductor {
				ir.keyWire = append(ir.keyWire, s...)
				continue
			}
			if ir.keyByteOpensWire(b) {
				ir.keyWire = append(ir.keyWire, s...)
				continue
			}
			if ir.pasting || !ir.hasKey(s) {
				continue
			}
			return s, nil
		}
	}
}

func (ir *inputReader) peekWireLineFrom(first byte) string {
	n := ir.r.Buffered()
	if n > 512 {
		n = 512
	}
	got, _ := ir.r.Peek(n)
	line := append([]byte{first}, got...)
	if i := bytesIndexAny(line); i >= 0 {
		line = line[:i]
	}
	return string(line)
}

func (ir *inputReader) keyByteOpensWire(first byte) bool {
	if ir.shouldSilence(first) {
		return true
	}
	return looksLikeBase64Chunk(ir.peekWireLineFrom(first))
}

func (ir *inputReader) finishKeyWireLine(raw string) {
	ir.keyWire = ir.keyWire[:0]
	line := normalizeUserLine(raw)
	ir.noteWireWindow(line)
	_ = ir.discardTransportLine(line)
}

func (ir *inputReader) readKeyRune(first byte) (string, error) {
	n := 1
	switch {
	case first&0xE0 == 0xC0:
		n = 2
	case first&0xF0 == 0xE0:
		n = 3
	case first&0xF8 == 0xF0:
		n = 4
	}
	buf := make([]byte, n)
	buf[0] = first
	for i := 1; i < n; i++ {
		b, err := ir.r.ReadByte()
		if err != nil {
			return "", err
		}
		buf[i] = b
	}
	return string(buf), nil
}

func (ir *inputReader) readKeyEscape() (string, error) {
	b, err := ir.r.ReadByte()
	if err != nil {
		return "", err
	}
	switch b {
	case '[':
		return ir.readKeyCSI()
	case 'O':
		b, err = ir.r.ReadByte()
		if err != nil {
			return "", err
		}
		return arrowName(b), nil
	default:
		if b >= 0x20 && b <= 0x2f {
			_, err = ir.r.ReadByte()
		}
		return "", err
	}
}

func (ir *inputReader) readKeyCSI() (string, error) {
	var params []byte
	for {
		b, err := ir.r.ReadByte()
		if err != nil {
			return "", err
		}
		if b >= 0x40 && b <= 0x7e {
			key := string(params)
			if b == '~' && key == "200" {
				ir.pasting = true
				return "", nil
			}
			if b == '~' && key == "201" {
				ir.pasting = false
				return "", nil
			}
			if ir.pasting {
				return "", nil
			}
			if b == 'A' || b == 'B' || b == 'C' || b == 'D' {
				return arrowName(b), nil
			}
			return "", nil
		}
		if len(params) < 32 {
			params = append(params, b)
		}
	}
}

func arrowName(b byte) string {
	switch b {
	case 'A':
		return "up"
	case 'B':
		return "down"
	case 'C':
		return "right"
	case 'D':
		return "left"
	default:
		return ""
	}
}

// IsTerminalProtocolLine reports transport-level noise (not scene dispatch policy).
func IsTerminalProtocolLine(s string) bool {
	if s == "" {
		return false
	}
	if isKittyWireLine(s) || s == itermConductorBegin {
		return true
	}
	if strings.ContainsRune(s, '\x1b') {
		return true
	}
	if looksLikeEscapeLine(s) {
		return true
	}
	if looksLikeBase64Chunk(s) {
		return true
	}
	if mostlyNonPrintable(s) {
		return true
	}
	return false
}

func isKittyWireLine(s string) bool {
	return strings.HasPrefix(s, "KITTY_")
}

func looksLikeBase64Chunk(s string) bool {
	const minLen = 48
	if len(s) < minLen {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		case r == '+', r == '/', r == '=':
		default:
			return false
		}
	}
	return true
}

func looksLikeEscapeLine(s string) bool {
	switch {
	case strings.HasPrefix(s, "]"), strings.HasPrefix(s, "P"), strings.HasPrefix(s, "_"),
		strings.HasPrefix(s, "^"), strings.HasPrefix(s, "X"):
		return len(s) > 1
	case strings.HasPrefix(s, "["):
		if len(s) < 2 {
			return false
		}
		for _, r := range s {
			if (r >= '0' && r <= '9') || r == ';' || r == ':' || r == '?' || r == '<' || r == '=' || r == '>' {
				continue
			}
			if r >= '@' && r <= '~' {
				return true
			}
		}
	}
	return false
}

func mostlyNonPrintable(s string) bool {
	var print, other int
	for _, r := range s {
		if r == '\t' || unicode.IsPrint(r) {
			print++
		} else {
			other++
		}
	}
	if other == 0 {
		return false
	}
	return other >= 4 || other*2 > print
}

func userLineByteOK(b byte) bool {
	if b == '\t' {
		return true
	}
	if b < 0x20 || b == 0x7f {
		return false
	}
	return true
}

// CSI key codes treated as Return (Kitty / xterm / common keypad reports).
var csiEnterKeys = []int{
	13,    // Return
	57414, // XK_KP_Enter (common in keyboard-protocol terminals)
	65293, // alternate keypad enter
}

func (ir *inputReader) noteLineEnd(kind, csi string) {
	ir.lastLineEnd = kind
	ir.lastSubmitCSI = csi
	ir.lastSkippedCSI = ""
}

func csiSubmitsLine(final byte, params []byte) bool {
	switch final {
	case 'u', '~':
		for _, key := range csiEnterKeys {
			if csiParamsContain(params, key) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// csiParamsContain reports whether any semicolon-separated CSI parameter equals key.
func csiParamsContain(params []byte, key int) bool {
	n := 0
	ok := false
	check := func() bool {
		return ok && n == key
	}
	for _, b := range params {
		if b == ';' || b == ':' {
			if check() {
				return true
			}
			n = 0
			ok = false
			continue
		}
		if b < '0' || b > '9' {
			n = 0
			ok = false
			continue
		}
		ok = true
		n = n*10 + int(b-'0')
	}
	return check()
}

func (ir *inputReader) consumeEscape() (submit bool, err error) {
	next, err := ir.r.ReadByte()
	if err != nil {
		return false, err
	}
	switch next {
	case '[':
		return ir.consumeCSI()
	case 'O':
		b, err := ir.r.ReadByte()
		if err != nil {
			return false, err
		}
		switch b {
		case 'M':
			ir.noteLineEnd("ss3_enter", "OM")
			return true, nil
		case 'A':
			ir.historyOlder()
		case 'B':
			ir.historyNewer()
		case 'C':
			ir.moveRight()
		case 'D':
			ir.moveLeft()
		case 'H':
			ir.moveHome()
		case 'F':
			ir.moveEnd()
		}
		return false, nil
	case ']':
		return false, ir.consumeUntilST(true)
	case 'P', '_', '^', 'X':
		return false, ir.consumeUntilST(false)
	case '\\':
		return false, nil
	default:
		if next >= 0x20 && next <= 0x2f {
			err = ir.consumeTwoCharEscape(next)
		}
		return false, err
	}
}

func (ir *inputReader) consumeCSI() (bool, error) {
	var params []byte
	for {
		b, err := ir.r.ReadByte()
		if err != nil {
			return false, err
		}
		if b >= 0x40 && b <= 0x7e {
			return ir.handleCSI(b, params), nil
		}
		if len(params) < 32 {
			params = append(params, b)
		}
	}
}

func (ir *inputReader) handleCSI(final byte, params []byte) bool {
	key := string(params)
	if final == '~' && key == "200" {
		ir.pasting = true
		return false
	}
	if final == '~' && key == "201" {
		ir.pasting = false
		return false
	}
	if ir.pasting {
		return false
	}
	if csiSubmitsLine(final, params) {
		ir.noteLineEnd("csi_enter", key+string(final))
		return true
	}
	switch final {
	case 'A':
		ir.historyOlder()
	case 'B':
		ir.historyNewer()
	case 'C':
		ir.moveRight()
	case 'D':
		ir.moveLeft()
	case 'H':
		ir.moveHome()
	case 'F':
		ir.moveEnd()
	case '~':
		switch csiFirstParam(params) {
		case 1, 7:
			ir.moveHome()
		case 3:
			ir.deleteForward()
		case 4, 8:
			ir.moveEnd()
		default:
			if ir.lineBufLen > 0 {
				ir.lastSkippedCSI = key + string(final)
			}
		}
	default:
		if (final == 'u' || final == '~') && ir.lineBufLen > 0 {
			ir.lastSkippedCSI = key + string(final)
		}
	}
	return false
}

func csiFirstParam(params []byte) int {
	n := 0
	ok := false
	for _, b := range params {
		if b == ';' || b == ':' {
			break
		}
		if b < '0' || b > '9' {
			return -1
		}
		ok = true
		n = n*10 + int(b-'0')
	}
	if !ok {
		return -1
	}
	return n
}

func (ir *inputReader) consumeTwoCharEscape(intermediate byte) error {
	_ = intermediate
	_, err := ir.r.ReadByte()
	return err
}

func (ir *inputReader) consumeUntilST(allowBEL bool) error {
	for {
		b, err := ir.r.ReadByte()
		if err != nil {
			return err
		}
		if allowBEL && b == 0x07 {
			return nil
		}
		if b == 0x1b {
			n, err := ir.r.Peek(1)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			if len(n) > 0 && n[0] == '\\' {
				_, _ = ir.r.ReadByte()
				return nil
			}
		}
	}
}

// normalizeUserLine trims space and validates UTF-8; invalid sequences are rejected
// as protocol noise.
func normalizeUserLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	if !utf8.ValidString(line) {
		return ""
	}
	return line
}
