# Input and dispatch

## Line mode (default)

When `Scene.Keys` is **false** (default), the runtime reads a line of text and matches `Scene.Inputs`:

1. **Exact match** on the full line (trimmed).
2. Else **first word** prefix match (`register user pass` → input key `register`).

Unmatched lines print `no such command`.

Use `ctx.InputLine()` in the action for the full line. Helpers: `experience.ParseFields`, `experience.Masked` (display only).

## Key mode

When `Scene.Keys` is **true**:

- Input names must be a **single rune** or `up`, `down`, `left`, `right` (`experience.SingleKeyInput`).
- Keys dispatch immediately; there is no `>` line editor on that scene.
- Typical pattern: list navigation bound to `up`/`down` actions that return `experience.MoveList`.

A scene with `Keys: true` **cannot** place a custom caret via `PlaceInput` during render (render fails).

## Caret (line mode)

Without `PlaceInput`, the prompt is `>` on the row below the scene output. After a successful command the runtime redraws the scene and prints a fresh `>`. It does not leave the command on that row. `Options.ScrollbackEcho` restores the old `> command` line under the frame.

`ctx.RecentCommands(n)` is the session transcript of lines that ran (oldest first; `n <= 0` is the whole list, capped at 32). A scene change clears it. `Options.Transcript` set to `runtime.TranscriptKeep` retains it, including the line that changed scenes. `runtime.TranscriptOff` records nothing. Drawing the list is the scene's job (`render.Screen`); chat messages and guestbook signatures stay in spaces.

Up and down still recall earlier lines in the line editor. That recall clears when the scene changes.

With `render.Screen` + `Input` + `Show`, you place the caret in one rectangle. See [Rendering](04-rendering.md).

## Rate limiting

`portico.Options`:

| Field | Default | Meaning |
|-------|---------|---------|
| `ActionRate` | `0` → 5/s | Actions per second per session |
| `ActionBurst` | `0` → 10 | Burst when limiting is on |
| `ActionRate` | negative | Disables limiting |

Limiting runs after successful input resolution.

## Logging

`LogLevel: "debug"` logs input protocol details (PTY line timing, ineligible lines, per-command `read_ms` / `render_ms`). See [Options](reference/options.md).

## Contributor note

The runtime implements a line editor (history, tab completion for scene input keys, resize). Site authors rely on the contract above; implementation lives under `runtime/` (not imported by sites).
