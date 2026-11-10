# Portico site agents (extended)

Extended contract for AI agents building an **SSH site** that imports `github.com/MrBlindMouse/Portico`. You are in a **consumer repo**, not the framework tree.

**Start in the site repo:** `portico init` writes **AGENTS.md** there—that file is the primary entry. Use this document when you need more detail at the same module tag.

## Version alignment

Read `require github.com/MrBlindMouse/Portico` in **this repo’s `go.mod`**. When unsure about an API, prefer that version over training data:

| Need | URL or command |
|------|----------------|
| Site **AGENTS.md** | Primary (in your repo, from `portico init`) |
| Manual | `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/manual/README.md` |
| This file | `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/site-agents.md` |
| Human how-to | `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/building-a-site.md` |
| Offline API | `go doc github.com/MrBlindMouse/Portico/components`, `…/installation` |

Use the **same git tag** as the module version (not `main` unless this repo tracks main).

**DESIGN.md** is vision; if something appears only in DESIGN and not in README, godoc, or these docs, treat it as **not shipped**.

## Plans — convenient starting points

`portico generate <name>` copies runnable starter Go into your repo. Each plan has its own root scene. **Merge** into `portico.New()` before `Build()`; edit the generated files.

```bash
portico generate guestbook
portico generate -list    # landing, guestbook, chat, register, form, blog, lobby
```

| Name | Job |
|------|-----|
| `landing` | Entrance menu — fill destinations in `main` |
| `guestbook`, `chat` | Shared spaces; also writes `rest.go` |
| `register` | Account signup — enable accounts in `Options` |
| `form`, `blog`, `lobby` | Form demo, assets, ready room |

Do not import plan sources from the framework module; only the copies in your repo.

`Guestbook` and `Chat` read the rest of the line (`sign hello`, `send hello`) via `InputLine()`. Later `Merge` calls win on id collision. Session start/end hook slices append.

## Hard rules

1. **Actions return effects only** (`[]components.Effect`). Never mutate session or space state in site code.
2. **Scenes only render** with `components.RenderContext`. No state changes in `Render`.
3. **Typed IDs** — `components.SceneID`, `ActionID`, `SpaceID`.
4. **`Build()` validates** — fix builder errors before chasing runtime bugs.
5. **Domain lives here** — scenes, actions, spaces, and assets in this repo.
6. **Imports (site code):** `portico`, `components`, `installation`, `render`, `asset`, `identity`, `testkit` (tests). Do not import `runtime` or `inspect`.

## Layout

| Piece | Typical location |
|-------|------------------|
| Installation (`Build()`) | `main.go` or `site/` package |
| Server config | `portico.Options` in `main.go` |
| Data + host key | `Options.DataDir` (e.g. `.portico-data`) |
| Static content | `Options.AssetDir` when using assets |

Entry pattern (matches `portico init` `main.go`):

```go
addr := flag.String("addr", ":2222", "SSH listen address")
flag.Parse()
inst, err := portico.New(). /* Merge(Guestbook()) */ .Build()
opts := portico.Options{
    DataDir:     dataDir,
    InspectAddr: "127.0.0.1:9922",
    StatusAddr:  "127.0.0.1:9923",
}
portico.Serve(*addr, inst, opts)
```

`portico.ServeDev(inst, opts)` is a dev shortcut (SSH `:2222`, loopback inspect/status when those fields are empty).

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
| `/inspect/validate` | POST | Validate an `installation.Wire` JSON body |
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

- `installation.LocalListIndex` (`list_index`) is the key when you pass `""`.
- `installation.MoveList(ctx, key, delta, n)` returns a `SetLocal` effect clamped to `0..n-1`.
- `installation.RenderList(ctx, key, items, visible)` draws that window. The selected row starts with `>` and uses the inverse role. `visible <= 0` shows every row. Rows are cut to `Size` width.
- Bind `up` and `down` with `Keys: true`.

```go
Render: func(ctx components.RenderContext) string {
    _, h := ctx.Size()
    return installation.RenderList(ctx, "", []string{"one", "two", "three"}, h-2)
},
Keys: true,
Inputs: map[string]components.ActionID{"up": "menu-up", "down": "menu-down"},
```

```go
Run: func(ctx components.ActionContext) []components.Effect {
    return installation.MoveList(ctx, "", -1, 3)
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

## Identity

A session that connects with no credentials is a guest. That path stays open when the site has no key allow/deny list, no password accounts, and `RequireAccount` is false, so `ssh` with no key still works. The SSH username (OpenSSH sends your local login when you omit `user@`) is the display label for that connection. It is not persisted and not unique. `Identity.String()` uses a reserved nickname, then the account name, then that login.

SSH clients stop when "none" succeeds, so a key or password is only used once that path is off. Turning on accounts, a key list, or `RequireAccount` does that. A public key stays a key for persistence unless the fingerprint is linked to an account. A password login is the account named by the SSH user. `RequireAccount` rejects guests and keys that are not linked.

## Testing

Site integration tests can use **`testkit`** (`testkit.New(exp)`) for an in-memory engine and store. Prefer testing pure helpers and `Build()` validation when that is enough; full SSH e2e is optional.

## Production notes

- Dev: **`ServeDev`** → SSH **:2222**. Production: **`Serve(addr, …)`** with `-addr` or config in `main` (see README **Production**).
- **Two SSH daemons on :22:** second public IP + DNS hostname; bind Portico to that IP (`203.0.113.10:22`); restrict **sshd** `ListenAddress` to the primary IP so both can use port 22.
- Dev `Accounts` store is plaintext; production should use `identity.FileAccountStore` (bcrypt), or `HashedAccounts` with `DataDir`. Account store precedence is `AccountBackend`, then hashed files, then `AccountStore`, then `Accounts.Accounts`.
- Inspect/status HTTP: localhost or disabled in prod unless protected.

## Framework vs site docs

| Doc | Audience |
|-----|----------|
| **`AGENTS.md` in the site repo** (from `portico init`) | Primary agent entry — start here |
| This file | Same contract, extended (inspect API, rendering detail, production) at your module tag |
| `docs/building-a-site.md` | Humans + agents (how-to) |
| Framework root `AGENTS.md` | Portico contributors only |
