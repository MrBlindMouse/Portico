package runtime

import (
	"strings"
)

const logInputPreviewRunes = 48

// logInputPreview returns a short, log-safe summary of a PTY line.
func logInputPreview(line string) string {
	if line == "" {
		return `""`
	}
	r := []rune(line)
	if len(r) <= logInputPreviewRunes {
		return line
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
