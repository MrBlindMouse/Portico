package experience

import "strings"

// ParseFields splits line into at most n whitespace-separated fields.
func ParseFields(line string, n int) []string {
	line = strings.TrimSpace(line)
	if line == "" || n <= 0 {
		return nil
	}
	var out []string
	for len(out) < n && line != "" {
		var field string
		field, line, _ = strings.Cut(line, " ")
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		out = append(out, field)
		line = strings.TrimSpace(line)
	}
	if len(out) == n-1 && line != "" {
		out = append(out, line)
	}
	return out
}

// Masked shows stars for each rune (display only).
func Masked(secret string) string {
	if secret == "" {
		return ""
	}
	return strings.Repeat("*", len(secret))
}
