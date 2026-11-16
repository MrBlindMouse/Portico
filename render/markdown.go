package render

import (
	"strings"
	"unicode"
)

// FormatWriteup converts a small markdown subset into terminal-ready text.
// Headings use Theme.Bold with spacing below; lists support nesting via leading
// spaces (two per level). Inline **bold** and `code` use Bold and Muted.
// width <= 0 skips line wrapping; otherwise each output line is wrapped.
func FormatWriteup(markdown string, width int, theme Theme) string {
	if markdown == "" {
		return ""
	}
	if theme.Bold == nil {
		theme = Plain
	}
	body := formatWriteup(markdown, theme)
	if width <= 0 {
		return body
	}
	lines := strings.Split(body, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			out = append(out, "")
			continue
		}
		out = append(out, wrapLine(line, width)...)
	}
	return strings.Join(out, "\n")
}

func formatWriteup(md string, theme Theme) string {
	lines := strings.Split(md, "\n")
	var out []string
	var inFence bool
	var para []string

	flushPara := func() {
		if len(para) == 0 {
			return
		}
		text := applyInline(strings.Join(para, " "), theme)
		out = append(out, text)
		para = nil
	}

	blanks := func(n int) {
		for i := 0; i < n; i++ {
			out = append(out, "")
		}
	}

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			flushPara()
			inFence = !inFence
			continue
		}
		if inFence {
			code := line
			if theme.Muted != nil {
				code = theme.Muted(code)
			}
			out = append(out, code)
			continue
		}

		if trimmed == "" {
			flushPara()
			out = append(out, "")
			continue
		}

		if level, title, ok := headingLevel(trimmed); ok {
			flushPara()
			switch level {
			case 1:
				blanks(1)
				out = append(out, theme.Bold(applyInline(title, theme)))
				blanks(2)
			case 2:
				blanks(1)
				out = append(out, theme.Bold(applyInline(title, theme)))
				blanks(1)
			default:
				out = append(out, theme.Bold(applyInline(title, theme)))
				blanks(1)
			}
			continue
		}

		if depth, marker, body, ok := listLine(line); ok {
			flushPara()
			indent := strings.Repeat("  ", depth+1)
			text := applyInline(body, theme)
			out = append(out, indent+marker+text)
			continue
		}

		para = append(para, trimmed)
	}
	flushPara()
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

func headingLevel(line string) (level int, title string, ok bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return 0, "", false
	}
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	if i == 0 || i > 6 {
		return 0, "", false
	}
	if i >= len(line) {
		return 0, "", false
	}
	if line[i] != ' ' && line[i] != '\t' {
		return 0, "", false
	}
	title = strings.TrimSpace(line[i:])
	if title == "" {
		return 0, "", false
	}
	return i, title, true
}

func listLine(line string) (depth int, marker string, body string, ok bool) {
	leading := 0
	for leading < len(line) {
		switch line[leading] {
		case ' ':
			leading++
		case '\t':
			leading += 2
		default:
			goto doneLead
		}
	}
doneLead:
	depth = leading / 2
	if depth > 8 {
		depth = 8
	}
	rest := strings.TrimSpace(line[leading:])
	if rest == "" {
		return 0, "", "", false
	}

	if strings.HasPrefix(rest, "- ") || strings.HasPrefix(rest, "* ") || strings.HasPrefix(rest, "+ ") {
		return depth, "• ", strings.TrimSpace(rest[2:]), true
	}

	// Ordered list: "1. item"
	i := 0
	for i < len(rest) && unicode.IsDigit(rune(rest[i])) {
		i++
	}
	if i > 0 && i+1 < len(rest) && rest[i] == '.' && rest[i+1] == ' ' {
		num := rest[:i+1]
		return depth, num + " ", strings.TrimSpace(rest[i+2:]), true
	}
	return 0, "", "", false
}

func applyInline(s string, theme Theme) string {
	return applyCodeSpans(applyBoldSpans(s, theme.Bold), theme.Muted)
}

func applyBoldSpans(s string, bold func(string) string) string {
	parts := strings.Split(s, "**")
	if len(parts) == 1 {
		return s
	}
	var b strings.Builder
	for i, p := range parts {
		if i%2 == 1 && bold != nil && p != "" {
			b.WriteString(bold(p))
		} else {
			b.WriteString(p)
		}
	}
	return b.String()
}

func applyCodeSpans(s string, muted func(string) string) string {
	parts := strings.Split(s, "`")
	if len(parts) == 1 {
		return s
	}
	var b strings.Builder
	for i, p := range parts {
		if i%2 == 1 && muted != nil && p != "" {
			b.WriteString(muted(p))
		} else {
			b.WriteString(p)
		}
	}
	return b.String()
}