package components

import (
	"strings"
	"unicode/utf8"
)

// ParseFields splits line into at most n fields. The last field is the remainder of the line.
func ParseFields(line string, n int) []string {
	line = strings.TrimSpace(line)
	if line == "" || n <= 0 {
		return nil
	}
	var out []string
	for len(out) < n-1 && line != "" {
		var field string
		field, line, _ = strings.Cut(line, " ")
		field = strings.TrimSpace(field)
		line = strings.TrimSpace(line)
		if field == "" {
			continue
		}
		out = append(out, field)
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

// Masked shows one star per rune (display only).
func Masked(secret string) string {
	n := utf8.RuneCountInString(secret)
	if n == 0 {
		return ""
	}
	return strings.Repeat("*", n)
}
