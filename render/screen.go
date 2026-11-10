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

// Screen is a grid of cells a scene draws, plus at most one input rectangle.
type Screen struct {
	w, h int
	cell []rune
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
	return &Screen{w: width, h: height, cell: cell}
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

// Show returns the grid. Trailing spaces on each row are removed.
// When Input was called and target is non-nil, the caret is placed on target.
// The prompt is not part of the grid. The runtime paints it in the rectangle.
func (s *Screen) Show(target SlotTarget) string {
	if s == nil {
		return ""
	}
	if target != nil && s.has {
		target.PlaceInput(s.slot)
	}
	return s.String()
}

// String returns the grid without a caret. Trailing spaces on each row are removed.
func (s *Screen) String() string {
	if s == nil || s.h == 0 {
		return ""
	}
	var b strings.Builder
	for r := 0; r < s.h; r++ {
		if r > 0 {
			b.WriteByte('\n')
		}
		line := string(s.cell[r*s.w : (r+1)*s.w])
		b.WriteString(strings.TrimRight(line, " "))
	}
	return b.String()
}
