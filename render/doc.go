// Package render draws terminal text for scene Render functions.
//
// Helpers return strings. They do not read sessions or apply effects.
// Widths are rune counts. Fit text first, then apply a Theme: these helpers
// do not skip ANSI codes, and they do not measure wide-glyph display columns.
//
// A scene reads the window from components.RenderContext.Size and the active
// theme from RenderContext.Theme.
//
// Screen is a grid plus one optional caret rectangle. Draw and DrawWithin store
// plain runes only—do not draw Theme role output or other ANSI into the grid.
// Use Screen.DrawList for a marked list. Show applies inverse only on that
// column span when the target implements ThemeSource (the runtime render
// context does). MarkRole does the same for bold, muted, and the other roles.
// SubmitAll(nil) repaints live sessions. portico.Every is not built.
// Fit text first, then theme, for standalone strings from List, Box, and Table.
//
// DrawWithin clips a region. BoxBody, ListRow, and TableCell report the
// rectangle of a block drawn at an origin. Show places the caret on the render
// context. With no rectangle, the runtime draws ">" on the row under the frame.
// Keys scenes paint with WriteGrid so a full-height grid does not scroll.
// See docs/site-agents.md.
package render
