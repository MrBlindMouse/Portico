package render

import "strings"

// Rect is a cell region on a screen. Row and Col are the top left.
type Rect struct {
	Row, Col, Width, Height int
}

// Input is the one caret rectangle in a screen, in column and row offsets
// from the top left of that screen. Width is the rectangle in columns and
// includes Prompt. The line editor types inside it and nowhere else.
type Input struct {
	Row    int
	Col    int
	Width  int
	Prompt string
}

// SlotTarget receives the caret from Screen.Show.
// A render context implements it.
type SlotTarget interface {
	PlaceInput(Input)
}

// ThemeSource is optional on the Show target. Marked cells use Theme roles
// only in the terminal string.
type ThemeSource interface {
	Theme() Theme
}

// Screen is a grid of cells a scene draws, plus at most one input rectangle.
// role is parallel to cell. RolePlain is the zero value.
type Screen struct {
	w, h int
	cell []rune
	role []Role
	slot Input
	has  bool
}

// NewScreen returns a grid of spaces. Non-positive dimensions become 0.
func NewScreen(width, height int) *Screen {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	cell := make([]rune, width*height)
	for i := range cell {
		cell[i] = ' '
	}
	return &Screen{w: width, h: height, cell: cell, role: make([]Role, width*height)}
}

// Draw writes text at row, col. A newline returns to col on the next row.
// Runes past the grid are dropped. Draw does not stop at a column; use DrawWithin
// when two regions share a row.
func (s *Screen) Draw(row, col int, text string) {
	s.DrawWithin(row, col, 0, 0, text)
}

// DrawWithin writes text at row, col and drops every rune outside that rectangle.
// A width or height of 0 or less runs through the screen edge.
// The returned rectangle is the region that lies on the screen.
func (s *Screen) DrawWithin(row, col, width, height int, text string) Rect {
	if s == nil || s.w == 0 || s.h == 0 {
		return Rect{}
	}
	width = paneExtent(col, width, s.w)
	height = paneExtent(row, height, s.h)
	r, c := row, col
	home := col
	right, bottom := col+width, row+height
	for _, ch := range text {
		switch ch {
		case '\n':
			r++
			c = home
			continue
		case '\r':
			c = home
			continue
		}
		if r >= row && r < bottom && c >= col && c < right && r >= 0 && r < s.h && c >= 0 && c < s.w {
			s.cell[r*s.w+c] = ch
		}
		c++
	}
	return clipRect(row, col, width, height, s.w, s.h)
}

// MarkRole sets the theme role on one row from col for width cells.
// A width of 0 or less runs through the right edge. Last call wins per cell.
// The cells stay plain. Show applies the role.
func (s *Screen) MarkRole(row, col, width int, role Role) {
	if s == nil || s.w == 0 || s.h == 0 || row < 0 || row >= s.h {
		return
	}
	if col < 0 {
		width += col
		col = 0
	}
	if col >= s.w {
		return
	}
	if width <= 0 {
		width = s.w - col
	}
	end := col + width
	if end > s.w {
		end = s.w
	}
	for c := col; c < end; c++ {
		s.role[row*s.w+c] = role
	}
}

// DrawList writes a list window at row, col with plain cells.
// The selected line is marked inverse only on that column span.
// Pass Show a ThemeSource to apply the role. width and height clip like DrawWithin.
// Returns the same region as ListRow.
func (s *Screen) DrawList(row, col, width, height int, items []string, selected, visible int) Rect {
	if s == nil || s.w == 0 || s.h == 0 {
		return Rect{}
	}
	start, end, selected, ok := listWindow(len(items), selected, visible)
	if !ok {
		return Rect{Row: row, Col: col}
	}
	for i := start; i < end; i++ {
		lineRow := row + i - start
		if height > 0 && lineRow >= row+height {
			break
		}
		if lineRow < 0 || lineRow >= s.h {
			continue
		}
		prefix := "  "
		if i == selected {
			prefix = "> "
		}
		line := prefix + items[i]
		if width > 0 {
			line = Truncate(line, width)
		}
		s.DrawWithin(lineRow, col, width, 1, line)
		span := width
		if span <= 0 {
			span = runeLen(line)
		}
		role := RolePlain
		if i == selected {
			role = RoleInverse
		}
		s.MarkRole(lineRow, col, span, role)
	}
	return ListRow(row, col, items, selected, visible, width)
}

func paneExtent(origin, size, limit int) int {
	if size > 0 {
		return size
	}
	if origin >= limit {
		return 0
	}
	if origin < 0 {
		return limit - origin
	}
	return limit - origin
}

func clipRect(row, col, width, height, screenW, screenH int) Rect {
	r0, c0 := row, col
	r1, c1 := row+height, col+width
	if r0 < 0 {
		r0 = 0
	}
	if c0 < 0 {
		c0 = 0
	}
	if r1 > screenH {
		r1 = screenH
	}
	if c1 > screenW {
		c1 = screenW
	}
	if r1 < r0 || c1 < c0 {
		return Rect{}
	}
	return Rect{Row: r0, Col: c0, Width: c1 - c0, Height: r1 - r0}
}

// Input records the single caret rectangle. A later call replaces it.
// The rectangle is clamped to the grid. Width includes prompt.
func (s *Screen) Input(row, col, width int, prompt string) {
	if s == nil || s.w == 0 || s.h == 0 {
		return
	}
	if row < 0 {
		row = 0
	}
	if row >= s.h {
		row = s.h - 1
	}
	if col < 0 {
		col = 0
	}
	if col >= s.w {
		col = s.w - 1
	}
	maxW := s.w - col
	if width <= 0 || width > maxW {
		width = maxW
	}
	s.slot = Input{Row: row, Col: col, Width: width, Prompt: prompt}
	s.has = true
}

// Show returns the grid. Trailing plain spaces on each row are removed.
// A non-plain role keeps its cells, including spaces, so a selection bar stays wide.
// When Input was called and target is non-nil, the caret is placed on target.
// The prompt is not part of the grid. The runtime paints it in the rectangle.
// Roles are applied only when target implements ThemeSource.
func (s *Screen) Show(target SlotTarget) string {
	if s == nil {
		return ""
	}
	if target != nil && s.has {
		target.PlaceInput(s.slot)
	}
	var th *Theme
	if ts, ok := target.(ThemeSource); ok {
		t := ts.Theme()
		th = &t
	}
	return s.emit(th)
}

// String returns the grid without a caret or roles.
// Trailing spaces on each row are removed.
func (s *Screen) String() string {
	return s.emit(nil)
}

func (s *Screen) emit(th *Theme) string {
	if s == nil || s.h == 0 {
		return ""
	}
	var b strings.Builder
	for r := 0; r < s.h; r++ {
		if r > 0 {
			b.WriteByte('\n')
		}
		end := s.rowEnd(r, th != nil)
		if end == 0 {
			continue
		}
		if th == nil {
			b.WriteString(string(s.cell[r*s.w : r*s.w+end]))
			continue
		}
		c := 0
		for c < end {
			role := s.roleAt(r, c)
			n := c + 1
			for n < end && s.roleAt(r, n) == role {
				n++
			}
			run := string(s.cell[r*s.w+c : r*s.w+n])
			if role == RolePlain {
				b.WriteString(run)
			} else {
				b.WriteString(th.wrap(role, run))
			}
			c = n
		}
	}
	return b.String()
}

// rowEnd is the exclusive end column. Styled emit keeps a trailing non-plain
// role so inverse padding is not trimmed. Plain emit stops at the last glyph.
func (s *Screen) rowEnd(r int, keepRole bool) int {
	end := 0
	base := r * s.w
	for c := 0; c < s.w; c++ {
		if s.cell[base+c] != ' ' || (keepRole && s.roleAt(r, c) != RolePlain) {
			end = c + 1
		}
	}
	return end
}

func (s *Screen) roleAt(row, col int) Role {
	if s == nil || len(s.role) != s.w*s.h || row < 0 || col < 0 || row >= s.h || col >= s.w {
		return RolePlain
	}
	return s.role[row*s.w+col]
}
