package installation

import (
	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/render"
)

// LocalListIndex is the session-local key MoveList and RenderList use when key is empty.
const LocalListIndex = "list_index"

// ListIndex reads a list selection stored as an int.
// A missing key, a negative value, or any other type reads as 0.
func ListIndex(ctx components.RenderContext, key string) int {
	if key == "" {
		key = LocalListIndex
	}
	v, ok := ctx.Local(key)
	if !ok || v == nil {
		return 0
	}
	n, ok := v.(int)
	if !ok || n < 0 {
		return 0
	}
	return n
}

// MoveList returns a components.SetLocal effect that moves the selection by delta and clamps
// it to 0..n-1. n <= 0 stores 0. key empty uses LocalListIndex.
// Call it from an action. The runtime applies the effect.
func MoveList(ctx components.ActionContext, key string, delta, n int) []components.Effect {
	if key == "" {
		key = LocalListIndex
	}
	next := ListIndex(ctx, key) + delta
	if n <= 0 {
		next = 0
	} else if next < 0 {
		next = 0
	} else if next >= n {
		next = n - 1
	}
	return []components.Effect{components.SetLocalEffect(key, next)}
}

// RenderList draws items with render.List.
// The index comes from key (or LocalListIndex). The row window is visible rows,
// and each row is fit to the terminal width from ctx.Size.
// The selected row uses the theme's inverse role.
func RenderList(ctx components.RenderContext, key string, items []string, visible int) string {
	w, _ := ctx.Size()
	return render.List(items, ListIndex(ctx, key), visible, w, ctx.Theme().Inverse)
}

// DrawList draws items on s with plain grid cells and inverse on the selected row at emit.
// Index key matches RenderList. width and height clip the list region.
func DrawList(s *render.Screen, ctx components.RenderContext, row, col, width, height int, key string, items []string, visible int) {
	s.DrawList(row, col, width, height, items, ListIndex(ctx, key), visible)
}
