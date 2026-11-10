package render

import (
	"fmt"
	"strings"
)

// List draws items in a window of visible rows.
// selected is clamped to the list. visible <= 0 shows every row.
// Once the selection moves past the first page, the window ends on the selected row.
// width <= 0 leaves each row uncut. Otherwise the row, including the "> " marker,
// is truncated before mark runs, so mark can add ANSI without being cut.
// The selected row is passed through mark when mark is non-nil.
// For a Screen layout, use Screen.DrawList and Show with a ThemeSource instead of
// passing mark into DrawWithin (CSI must not be stored in grid cells).
func List(items []string, selected, visible, width int, mark func(string) string) string {
	start, end, selected, ok := listWindow(len(items), selected, visible)
	if !ok {
		return ""
	}
	var b strings.Builder
	for i := start; i < end; i++ {
		if i > start {
			b.WriteByte('\n')
		}
		prefix := "  "
		if i == selected {
			prefix = "> "
		}
		line := prefix + items[i]
		if width > 0 {
			line = Truncate(line, width)
		}
		if i == selected && mark != nil {
			line = mark(line)
		}
		b.WriteString(line)
	}
	return b.String()
}

// ListRow is the selected line when that same List is drawn at row, col.
// Width is the list width. When width is 0 the rectangle is as wide as the marked line.
func ListRow(row, col int, items []string, selected, visible, width int) Rect {
	start, _, selected, ok := listWindow(len(items), selected, visible)
	if !ok {
		return Rect{Row: row, Col: col}
	}
	w := width
	if w <= 0 {
		w = runeLen("> " + items[selected])
	}
	return Rect{Row: row + selected - start, Col: col, Width: w, Height: 1}
}

func listWindow(n, selected, visible int) (start, end, sel int, ok bool) {
	if n == 0 {
		return 0, 0, 0, false
	}
	if selected < 0 {
		selected = 0
	}
	if selected >= n {
		selected = n - 1
	}
	start, end = 0, n
	if visible > 0 && visible < n {
		start = selected - visible + 1
		if start < 0 {
			start = 0
		}
		end = start + visible
		if end > n {
			end = n
		}
	}
	return start, end, selected, true
}

// Table lays out headers and rows in padded columns.
// widths sets a column's rune width. A missing or non-positive width uses the
// widest cell in that column. Cells are separated by two spaces.
func Table(headers []string, rows [][]string, widths []int) string {
	cols := len(headers)
	for _, row := range rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return ""
	}
	w := make([]int, cols)
	for i := 0; i < cols; i++ {
		if i < len(widths) && widths[i] > 0 {
			w[i] = widths[i]
			continue
		}
		n := 1
		if i < len(headers) && runeLen(headers[i]) > n {
			n = runeLen(headers[i])
		}
		for _, row := range rows {
			if i < len(row) && runeLen(row[i]) > n {
				n = runeLen(row[i])
			}
		}
		w[i] = n
	}
	var b strings.Builder
	if len(headers) > 0 {
		writeCells(&b, headers, w)
		b.WriteByte('\n')
		for i, cw := range w {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(strings.Repeat("─", cw))
		}
		if len(rows) > 0 {
			b.WriteByte('\n')
		}
	}
	for i, row := range rows {
		if i > 0 {
			b.WriteByte('\n')
		}
		writeCells(&b, row, w)
	}
	return b.String()
}

// TableCell is one body cell of a Table drawn at originRow, originCol
// with the same widths. header counts the header row and the rule row.
// Columns are separated by two spaces, matching Table.
// Width is the cell's column count. Pass it to Screen.Input to keep the caret there.
func TableCell(originRow, originCol int, widths []int, bodyRow, col int, header bool) Rect {
	row := originRow + bodyRow
	if bodyRow < 0 {
		row = originRow
	}
	if header {
		row += 2
	}
	if col < 0 {
		col = 0
	}
	colX := originCol
	for i := 0; i < col && i < len(widths); i++ {
		w := widths[i]
		if w < 1 {
			w = 1
		}
		colX += w + 2
	}
	width := 0
	if col < len(widths) {
		width = widths[col]
		if width < 1 {
			width = 1
		}
	}
	return Rect{Row: row, Col: colX, Width: width, Height: 1}
}

func writeCells(b *strings.Builder, cells []string, widths []int) {
	for i, cw := range widths {
		if i > 0 {
			b.WriteString("  ")
		}
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteString(Pad(cell, cw))
	}
}

// Progress returns a bracketed bar of width columns and a percent, as in "[##--] 50%".
// width includes the brackets. Values below 4 become 4.
// done is clamped to 0..total. total <= 0 draws an empty bar at 0%.
func Progress(done, total, width int) string {
	if width < 4 {
		width = 4
	}
	inner := width - 2
	filled := 0
	pct := 0
	if total > 0 {
		if done < 0 {
			done = 0
		}
		if done > total {
			done = total
		}
		filled = inner * done / total
		pct = 100 * done / total
	}
	bar := strings.Repeat("#", filled) + strings.Repeat("-", inner-filled)
	return fmt.Sprintf("[%s] %d%%", bar, pct)
}

// spinnerFrames is one ASCII cycle. The index is chosen by the site.
var spinnerFrames = []string{"|", "/", "-", `\`}

// SpinnerFrame returns one glyph for index i.
// It does not advance. Store i in session-local state and change it with
// SetLocalEffect from an action, or from Host.Submit, which also repaints.
// Portico does not run a clock.
func SpinnerFrame(i int) string {
	n := len(spinnerFrames)
	i %= n
	if i < 0 {
		i += n
	}
	return spinnerFrames[i]
}
