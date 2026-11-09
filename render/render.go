package render

import (
	"strings"
)

// Box draws a simple titled box around content.
func Box(title, body string, width int) string {
	if width < 20 {
		width = 20
	}
	inner := width - 2
	line := strings.Repeat("─", inner)
	var b strings.Builder
	b.WriteString("┌")
	b.WriteString(line)
	b.WriteString("┐\n")
	if title != "" {
		b.WriteString("│ ")
		b.WriteString(padTruncate(title, inner-2))
		b.WriteString(" │\n")
		b.WriteString("├")
		b.WriteString(line)
		b.WriteString("┤\n")
	}
	for _, row := range strings.Split(body, "\n") {
		b.WriteString("│ ")
		b.WriteString(padTruncate(row, inner-2))
		b.WriteString(" │\n")
	}
	b.WriteString("└")
	b.WriteString(line)
	b.WriteString("┘")
	return b.String()
}

// Columns lays out texts side by side with a minimum column width.
func Columns(cols ...string) string {
	if len(cols) == 0 {
		return ""
	}
	width := 0
	rows := make([][]string, len(cols))
	for i, c := range cols {
		lines := strings.Split(c, "\n")
		rows[i] = lines
		for _, ln := range lines {
			if len(ln) > width {
				width = len(ln)
			}
		}
	}
	if width < 10 {
		width = 10
	}
	height := 0
	for _, r := range rows {
		if len(r) > height {
			height = len(r)
		}
	}
	var out strings.Builder
	for y := 0; y < height; y++ {
		for i, r := range rows {
			if i > 0 {
				out.WriteString("  ")
			}
			if y < len(r) {
				out.WriteString(padTruncate(r[y], width))
			} else {
				out.WriteString(strings.Repeat(" ", width))
			}
		}
		out.WriteByte('\n')
	}
	return strings.TrimRight(out.String(), "\n")
}

// Emphasis wraps text with ANSI bold when enabled.
func Emphasis(text string, enableANSI bool) string {
	if !enableANSI {
		return text
	}
	return "\x1b[1m" + text + "\x1b[0m"
}

func padTruncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if len(s) > w {
		if w <= 1 {
			return s[:w]
		}
		return s[:w-1] + "…"
	}
	return s + strings.Repeat(" ", w-len(s))
}
