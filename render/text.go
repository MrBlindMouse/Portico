package render

import (
	"strings"
	"unicode/utf8"
)

// DefaultWidth and DefaultHeight are the window used before a client reports a size.
const (
	DefaultWidth  = 80
	DefaultHeight = 24
)

// ponytail: width is a rune count, not terminal display columns. Wide glyphs
// and ANSI sequences count wrong. Upgrade path is one display-width function
// used by Truncate, Pad, Wrap, and the align helpers.

func runeLen(s string) int { return utf8.RuneCountInString(s) }

// Truncate fits s to width runes. Shorter strings are unchanged.
// A cut longer than one rune ends with an ellipsis.
func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width == 1 {
		return string(r[:1])
	}
	return string(r[:width-1]) + "…"
}

// Pad fits s to width runes, truncating with an ellipsis or padding with spaces
// on the right. Newlines count as runes; use AlignLeft to pad each line.
func Pad(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = Truncate(s, width)
	n := runeLen(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

// Wrap breaks s on spaces so each line is at most width runes.
// Existing newlines are kept. Spaces inside a line are collapsed.
// A word longer than width is split. width <= 0 or an empty s returns "".
func Wrap(s string, width int) string {
	if s == "" || width <= 0 {
		return ""
	}
	parts := strings.Split(s, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, wrapLine(part, width)...)
	}
	return strings.Join(out, "\n")
}

func wrapLine(line string, width int) []string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return []string{""}
	}
	var lines []string
	var cur []rune
	flush := func() {
		lines = append(lines, string(cur))
		cur = nil
	}
	for _, word := range fields {
		wr := []rune(word)
		for len(wr) > width {
			if len(cur) > 0 {
				flush()
			}
			lines = append(lines, string(wr[:width]))
			wr = wr[width:]
		}
		if len(wr) == 0 {
			continue
		}
		if len(cur) == 0 {
			cur = append(cur, wr...)
			continue
		}
		if len(cur)+1+len(wr) <= width {
			cur = append(cur, ' ')
			cur = append(cur, wr...)
			continue
		}
		flush()
		cur = append(cur, wr...)
	}
	if len(cur) > 0 {
		flush()
	}
	return lines
}

// AlignLeft pads each line on the right to width runes.
func AlignLeft(s string, width int) string { return mapLines(s, width, Pad) }

// AlignRight pads each line on the left to width runes.
func AlignRight(s string, width int) string {
	return mapLines(s, width, func(line string, width int) string {
		if width <= 0 {
			return ""
		}
		r := []rune(line)
		if len(r) > width {
			return Truncate(line, width)
		}
		return strings.Repeat(" ", width-len(r)) + line
	})
}

// AlignCenter pads each line on both sides to width runes.
// An odd gap leaves the extra space on the right.
func AlignCenter(s string, width int) string {
	return mapLines(s, width, func(line string, width int) string {
		if width <= 0 {
			return ""
		}
		r := []rune(line)
		if len(r) > width {
			return Truncate(line, width)
		}
		gap := width - len(r)
		left := gap / 2
		return strings.Repeat(" ", left) + line + strings.Repeat(" ", gap-left)
	})
}

func mapLines(s string, width int, fn func(string, int) string) string {
	if width <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = fn(line, width)
	}
	return strings.Join(lines, "\n")
}
