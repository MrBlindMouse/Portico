package render

func ansi(code, s string) string { return "\x1b[" + code + "m" + s + "\x1b[0m" }

func plain(s string) string { return s }

// Theme is a fixed set of emphasis roles. Layout stays in the other helpers.
// Copy Default or Plain and replace a field to change one role.
// Do not nest calls: each role resets SGR at the end of the string.
type Theme struct {
	Bold    func(string) string
	Muted   func(string) string
	Error   func(string) string
	Accent  func(string) string
	OK      func(string) string
	Warn    func(string) string
	Inverse func(string) string
}

// Default is the ANSI theme used when a site does not set Options.Theme.
var Default = Theme{
	Bold:    func(s string) string { return ansi("1", s) },
	Muted:   func(s string) string { return ansi("90", s) },
	Error:   func(s string) string { return ansi("31", s) },
	Accent:  func(s string) string { return ansi("36", s) },
	OK:      func(s string) string { return ansi("32", s) },
	Warn:    func(s string) string { return ansi("33", s) },
	Inverse: func(s string) string { return ansi("7", s) },
}

// Plain leaves every string unchanged. The runtime uses it when NO_COLOR is
// set to a non-empty value and Options.Theme is nil.
var Plain = Theme{
	Bold:    plain,
	Muted:   plain,
	Error:   plain,
	Accent:  plain,
	OK:      plain,
	Warn:    plain,
	Inverse: plain,
}
