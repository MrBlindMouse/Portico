# Rendering

## Output model

Each scene returns one **string**. After every applied action batch (and on terminal resize), the runtime paints that string. Terminal size comes from `ctx.Size()` (columns, rows); default **80×24** until the client reports dimensions.

## Theme

- `opts.Theme` sets the session theme (`render.Theme`).
- When `Theme` is nil and `NO_COLOR` is non-empty, `render.Plain` is used.
- Otherwise `render.Default` ANSI roles apply: bold, muted, error, accent, ok, warn, inverse.

Use `ctx.Theme()` in `Render`; apply roles after fitting text (`render` does not strip ANSI for width). Theme output is for terminal strings. `Screen` grid cells must stay plain. `Screen.DrawList` marks the selected column. `Screen.MarkRole` marks bold, muted, error, accent, ok, warn, or inverse on plain cells. `Show` injects the theme. `Host.SubmitAll(nil)` repaints every live session. There is no `portico.Every` yet.

## Text helpers

| Function | Role |
|----------|------|
| `render.Wrap`, `Truncate`, `Pad` | Fit to width (rune counts) |
| `render.AlignLeft`, `AlignRight`, `AlignCenter` | Alignment after width known |

## Layout helpers

| Function | Role |
|----------|------|
| `render.Box`, `BoxBody` | Bordered region; `BoxBody` returns caret `Rect` |
| `render.Columns` | Side-by-side columns |
| `render.Table`, `TableCell` | Fixed-width table |
| `render.List`, `ListRow` | List with selection marker |
| `render.Progress` | Static bar `[##--] 50%` |
| `render.SpinnerFrame(i)` | Single glyph; you advance `i` in local state (no framework timer) |

## Lists in a scene

Store selected index in session-local state:

- Default key: `installation.LocalListIndex` (`list_index`).
- `installation.MoveList(ctx, key, delta, n)` → `SetLocal` effect, clamped.
- `installation.RenderList(ctx, key, items, visible)` draws the window as one string; selected row uses `>` and inverse role.
- `installation.DrawList` / `render.Screen.DrawList` for lists inside a `Screen` layout.

Bind `up`/`down` with `Scene.Keys: true`. Bind `enter` to confirm or open, `backspace` to go back, and `tab` to move to the next item. `home`, `end`, and `delete` are named keys too. Keys scenes use `WriteGrid` at paint time.

## Forms

Multi-step prompts (optional pattern in `installation` package):

- `StartForm`, `AdvanceFormValue`, `CancelForm`
- `FormText`, `RenderForm` for display

See the form plan (`go run …/cmd/portico generate form`) for an example graph.

## Screen grid and caret

```go
w, h := ctx.Size()
s := render.NewScreen(w, h)
s.DrawWithin(0, 0, w/2, h-1, leftPanel)
s.DrawWithin(0, w/2, w-w/2, h-1, rightPanel)
s.Input(h-1, 0, w/2, ">")
return s.Show(ctx)
```

- `DrawWithin` clips so regions do not overwrite neighbors. Do not draw ANSI into the grid.
- `DrawList` writes list rows as plain cells. Inverse on `Show(ctx)` covers the list column of the selected line.
- `MarkRole` stores a theme role on plain cells. `Show` wraps each contiguous run.
- One `Input` per frame; second call replaces the first.
- `Keys: true` + placed input → render error.

## PlaceInput on RenderContext

`ctx.PlaceInput(render.Input{...})` records caret metadata for the frame only (no session mutation). Prefer `Screen` for multi-region layouts.
