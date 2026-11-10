package plans

import "strings"

// restAfter returns the text after a prefix command ("send hello" → "hello").
func restAfter(line, cmd string) string {
	line = strings.TrimSpace(line)
	head, rest, found := strings.Cut(line, " ")
	if head != cmd || !found {
		return ""
	}
	return strings.TrimSpace(rest)
}
