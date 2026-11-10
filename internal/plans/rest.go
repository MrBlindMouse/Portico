package plans

import "strings"

// How to change this plan
//
// go run github.com/MrBlindMouse/Portico/cmd/portico generate writes this
// file next to guestbook.go and chat.go.
// Generating the second of those two needs -force when rest.go is already there.
// Other plans do not use this helper. Prefix commands ("sign hello", "nick ada")
// take the text after the first token. The first token must be the command;
// "sign" alone and "other hello" return empty.

// restAfter returns the text after a prefix command ("send hello there" → "hello there").
func restAfter(line, cmd string) string {
	line = strings.TrimSpace(line)
	head, rest, found := strings.Cut(line, " ")
	if head != cmd || !found {
		return ""
	}
	return strings.TrimSpace(rest)
}
