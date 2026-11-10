// Package render draws terminal text for scene Render functions.
//
// Helpers return strings. They do not read sessions or apply effects.
// Widths are rune counts. Fit text first, then apply a Theme: these helpers
// do not skip ANSI codes, and they do not measure wide-glyph display columns.
//
// A scene reads the window from components.RenderContext.Size and the active
// theme from RenderContext.Theme.
//
// Screen is a grid plus one optional caret rectangle. DrawWithin clips a
// region. BoxBody, ListRow, and TableCell report the rectangle of a block
// drawn at an origin. Show places the caret on the render context. With no
// rectangle, the runtime draws ">" on the row under the frame.
// See docs/site-agents.md.
package render
