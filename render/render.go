package render

import (
	"strings"
)

// BoxBody is the inside of a Box of the given width drawn at row, col.
// The rectangle starts under the top border, and under the title and rule when title is set.
// Height is the number of body lines. An empty body is one line, matching Box.
func BoxBody(row, col int, title, body string, width int) Rect {
	if width < 20 {
		width = 20
	}
	bodyRow := row + 1
	if title != "" {
		bodyRow = row + 3
	}
	lines := 1
	if body != "" {
		lines = strings.Count(body, "\n") + 1
	}
	inner := width - 4
	if inner < 1 {
		inner = 1
	}
	return Rect{Row: bodyRow, Col: col + 2, Width: inner, Height: lines}
}

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
		b.WriteString(Pad(title, inner-2))
		b.WriteString(" │\n")
		b.WriteString("├")
		b.WriteString(line)
		b.WriteString("┤\n")
	}
	for _, row := range strings.Split(body, "\n") {
		b.WriteString("│ ")
		b.WriteString(Pad(row, inner-2))
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
			if n := runeLen(ln); n > width {
				width = n
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
				out.WriteString(Pad(r[y], width))
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
	return Default.Bold(text)
}
