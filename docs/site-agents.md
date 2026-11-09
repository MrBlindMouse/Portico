# Portico site agents

Contract for AI agents and operators building an **SSH site** that imports `github.com/MrBlindMouse/Portico`. You are in a **consumer repo**, not the framework tree.

## Version alignment

Read `require github.com/MrBlindMouse/Portico` in **this repo’s `go.mod`**. When unsure about an API, prefer that version over training data:

- Upstream contract: `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/site-agents.md`
- Human guide: `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/building-a-site.md`
- Offline API: `go doc github.com/MrBlindMouse/Portico/experience`

Use the **same git tag** as the module version (not `main` unless this repo tracks main).

**DESIGN.md** is vision; if something appears only in DESIGN and not in README, godoc, or these docs, treat it as **not shipped**.

## Hard rules

1. **Actions return effects only** (`[]experience.Effect`). Never mutate session or space state in site code.
2. **Scenes only render** with `experience.RenderContext`. No state changes in `Render`.
3. **Do not import** `github.com/MrBlindMouse/Portico/runtime`, `.../internal/...`, or `.../inspect` from site code.
4. **Typed IDs** — `experience.SceneID`, `ActionID`, `SpaceID`.
5. **`Build()` validates** — fix builder errors before chasing runtime bugs.
6. **Domain lives here** — scenes, actions, spaces, and assets in this repo; copy scaffolds to edit, do not fork framework packages.

## Layout

| Piece | Typical location |
|-------|------------------|
| `Serve` / `ServeDev`, options | `main.go` |
| Experience definition | `main.go` or `site/` package |
| Data + host key | `Options.DataDir` (e.g. `.portico-data`) |
| Static content | `Options.AssetDir` when using assets |

Entry pattern:

```go
exp, err := portico.New(). /* scenes, actions, Merge(scaffold.*) */ .Build()
portico.ServeDev(exp, portico.Options{DataDir: ".portico-data"})
```

## Iteration loop

```bash
go run .
# terminal 2
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
# terminal 3 (while server runs)
curl -s http://127.0.0.1:9922/inspect/snapshot | jq .
```

| Endpoint | Method | Use |
|----------|--------|-----|
| `/inspect/snapshot` | GET | Scenes, sessions, spaces summary |
| `/inspect/sessions` | GET | All sessions |
| `/inspect/session/{id}` | GET | One session render preview |
| `/inspect/spaces` | GET | Shared space keys |
| `/inspect/identity/{fingerprint}` | GET | Per-key persistence |
| `/inspect/validate` | POST | Validate a builder JSON body |
| `/inspect/reload-preview` | POST | Dry-run hot reload |

CLI helper (from dev machine with `go install` on Portico): `portico inspect -url http://127.0.0.1:9922 -resource snapshot`

Environment: `PORTICO_DATA`, `PORTICO_LOG_LEVEL` (when wired in `main`). Do not expose inspect/status HTTP on untrusted networks.

## Contexts and effects

| Context | Where | Read / do |
|---------|-------|-----------|
| `RenderContext` | `Scene.Render` | Identity, local, assets, space state, `Size()`, `Theme()` |
| `ActionContext` | `Action.Run` | Above + `InputLine()`, space helpers, persist effects |

Common effects: `SetLocalEffect`, `TransitionEffect`, `PersistEffect`, `JoinSpaceEffect`, `LeaveSpaceEffect`, `SpaceSetEffect`, `AppendSpaceListEffect`, `BroadcastEffect`, `ReserveNicknameEffect`, `RegisterAccountEffect` (requires account store in options).

Cross-session “events”: update space data, then `BroadcastEffect`. There is no in-process event bus.

## Input dispatch

Scene `Inputs` map **keys** to `ActionID`:

- Exact line match, or **first word** prefix (`register user pass` → key `register`).
- Multi-word commands: use `ctx.InputLine()` in the action.
- `Scene.Keys` dispatches one bound key (`up`, `down`, `left`, `right`, or one rune) with no line prompt. Use it for a list selection.

## Rendering

`Scene.Render` returns the whole frame. The runtime paints it after each applied action and on resize.

- `ctx.Size()` is columns and rows. It is 80 by 24 until the client reports a window.
- `ctx.Theme()` is `Options.Theme`, or `render.Plain` when `NO_COLOR` is set to a non-empty value, or `render.Default`.

Fit and draw with `render`: `Wrap`, `Truncate`, `Pad`, `AlignLeft`, `AlignRight`, `AlignCenter`, `Box`, `Columns`, `Table`, `List`, `Progress`, `SpinnerFrame`. Widths are rune counts. Style a string after you fit it. These helpers do not skip ANSI codes.

Theme roles: bold, muted, error, accent, ok, warn, inverse. Copy `render.Default` or `render.Plain` and replace a field to change one role. Do not nest role calls.

A list selection is an `int` in session-local state:

- `experience.LocalListIndex` (`list_index`) is the key when you pass `""`.
- `experience.MoveList(ctx, key, delta, n)` returns a `SetLocal` effect clamped to `0..n-1`.
- `experience.RenderList(ctx, key, items, visible)` draws that window. The selected row starts with `>` and uses the inverse role. `visible <= 0` shows every row. Rows are cut to `Size` width.
- Bind `up` and `down` with `Keys: true`.

```go
Render: func(ctx experience.RenderContext) string {
    _, h := ctx.Size()
    return experience.RenderList(ctx, "", []string{"one", "two", "three"}, h-2)
},
Keys: true,
Inputs: map[string]experience.ActionID{"up": "menu-up", "down": "menu-down"},
```

```go
Run: func(ctx experience.ActionContext) []experience.Effect {
    return experience.MoveList(ctx, "", -1, 3)
},
```

`render.Progress(done, total, width)` is a static bar. `render.SpinnerFrame(i)` is one glyph. Store `i` yourself and change it with `SetLocalEffect`. The framework does not tick a clock.

## Caret

With no placed input, and without `Keys: true`, the caret is `>` on the full row under the frame. Typing stays on that row.

`render.Screen` is a grid. `Draw` writes text to the screen edge. `DrawWithin(row, col, width, height, text)` drops runes outside that rectangle, so a left column cannot paint over the right. A width or height of `0` runs through the screen edge. `Show(ctx)` returns the grid and places the caret. The prompt is painted in the rectangle and counts toward `width`. A second `Input` replaces the first.

These return a `render.Rect` for a block drawn at an origin, so the caret can sit on it:

| Function | Rectangle |
|----------|-----------|
| `BoxBody(row, col, title, body, width)` | Inside the border, under the title when there is one |
| `ListRow(row, col, items, selected, visible, width)` | The selected line of that `List` |
| `TableCell(row, col, widths, bodyRow, colIndex, header)` | One body cell. `header` counts the header and the rule |

```go
s := render.NewScreen(w, h)
s.DrawWithin(2, 4, 40, 0, render.Table(headers, rows, widths))
cell := render.TableCell(2, 4, widths, bodyRow, 0, true)
s.Input(cell.Row, cell.Col, cell.Width, ">")
return s.Show(ctx)
```

A scene with `Keys: true` that also places an input fails when it renders. `Build` does not run `Render`, so it cannot see the rectangle.

## Scaffolds

Merge with `portico.New().Merge(scaffold.Guestbook())` etc. Later merges win on ID collisions. Session start and end hooks append.

| Function | Adds |
|----------|------|
| `Landing(destinations)` | Root menu, `go-*` transitions |
| `Guestbook`, `Chat` | Shared spaces |
| `Registration` | Needs `Options.Accounts` |
| `NicknameCommands` | Bind `nick` on your scenes. `RecipeAccountsSite` does this |
| `Form`, `Blog`, `Lobby` | Demos / assets |
| `RecipeCommunity`, `RecipeAccountsSite` | Pre-wired graphs |

Hints: `portico scaffold <name>` prints merge snippets. See `scaffold/AGENTS.md` in the module at your tag.

`Guestbook` and `Chat` read the rest of the line (`sign hello`, `send hello`) via `InputLine()`. To change a scaffold, copy it into the site and edit that copy (`examples/notes`). Do not patch the module.

## Identity

A session that connects with no credentials is a guest. That path stays open when the site has no key allow/deny list, no password accounts, and `RequireAccount` is false, so `ssh` with no key still works.

SSH clients stop when "none" succeeds, so a key or password is only used once that path is off. Turning on accounts, a key list, or `RequireAccount` does that. A public key stays a key identity unless the fingerprint is linked to an account. A password login is the account named by the SSH user. `RequireAccount` rejects guests and keys that are not linked.

## Testing

Site integration tests can use **`testkit`** (`testkit.New(exp)`) for an in-memory engine and store. Prefer testing pure helpers and `Build()` validation when that is enough; full SSH e2e is optional.

## Production notes

- Dev: SSH **:2222**. Production may use `:22` + `setcap` (see README).
- Dev `Accounts` store is plaintext; production should use `identity.FileAccountStore` (bcrypt), or `HashedAccounts` with `DataDir`. Account store precedence is `AccountBackend`, then hashed files, then `AccountStore`, then `Accounts.Accounts`.

## Framework vs site docs

| Doc | Audience |
|-----|----------|
| This file | Site repo agents |
| `docs/building-a-site.md` | Humans + agents (how-to) |
| Framework `AGENTS.md` | Portico contributors only |
