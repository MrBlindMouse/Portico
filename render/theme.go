package render

// Theme provides ANSI styling helpers for terminal output.
type Theme struct {
	Bold  func(string) string
	Muted func(string) string
	Error func(string) string
}

// Default is the package-wide terminal theme (ANSI on).
var Default = Theme{
	Bold:  func(s string) string { return "\x1b[1m" + s + "\x1b[0m" },
	Muted: func(s string) string { return "\x1b[90m" + s + "\x1b[0m" },
	Error: func(s string) string { return "\x1b[31m" + s + "\x1b[0m" },
}
