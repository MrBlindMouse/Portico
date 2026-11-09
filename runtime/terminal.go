package runtime

// Terminal abstracts SSH or fake terminal I/O for the session loop.
type Terminal interface {
	WriteFrame(text string) error
	ReadInput() (string, error)
	OnResize(func(width, height int))
	Close() error
}

// LineEndTracer is optional. SSHTerminal implements it for debug logs after ReadInput.
type LineEndTracer interface {
	LineEndReason() string
	SubmitCSI() string
	SkippedCSI() string
}

// InputPrompter is optional. SSHTerminal writes a command line after each scene frame.
// CommitInputLine reprints the submitted line and a fresh prompt on the next line.
// RejectInput reports a line the current scene does not accept, then reprints the prompt.
type InputPrompter interface {
	BeginInputLine() error
	CommitInputLine(line string) error
	RejectInput(msg string) error
}

// LineEditor is optional. SSHTerminal edits the current line and restores it after a redraw.
type LineEditor interface {
	SetRedraw(func())
	SetComplete(func(prefix string) string)
	RestoreInputLine()
}

// KeyMode is optional. A nil or empty key list is line entry.
// A non-empty list reads one matching key and does not wait for Enter.
type KeyMode interface {
	SetKeys(keys []string)
}
