package runtime

import (
	"strings"
)

const logInputPreviewRunes = 48

// logInputPreview returns the first word of a PTY line, truncated for debug logs.
// The rest of the line stays out of the log (register passwords, and similar).
func logInputPreview(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return `""`
	}
	token, _, _ := strings.Cut(line, " ")
	r := []rune(token)
	if len(r) <= logInputPreviewRunes {
		return token
	}
	return string(r[:logInputPreviewRunes]) + "…"
}

func logInputAttrs(line string) (preview string, length int) {
	return logInputPreview(line), len(line)
}

// protocolLineKind classifies transport-level noise for debug logs.
func protocolLineKind(line string) string {
	switch {
	case isKittyWireLine(line):
		return "kitty"
	case looksLikeBase64Chunk(line):
		return "base64"
	case strings.ContainsRune(line, '\x1b'):
		return "escape"
	case looksLikeEscapeLine(line):
		return "escape"
	case mostlyNonPrintable(line):
		return "binary"
	default:
		return "protocol"
	}
}
