# Portico

Build interactive, multi-user terminal sites over SSH.

You write the experience (scenes, actions, spaces). Portico handles SSH, sessions, input, shared state, and persistence.

## Create a site

A Portico site is an ordinary Go program that imports Portico as a library. Either generate that program, or write it. Both end the same way: `go run .`, then SSH to port 2222.

### Generate it (`portico init`)

From an empty directory:

```bash
mkdir my-site && cd my-site
go run github.com/MrBlindMouse/Portico/cmd/portico@latest init -module github.com/yourname/my-site
go run .
```

`init` creates `go.mod` (if needed), runs `go get`, and writes `main.go`, `AGENTS.md`, and `README.md`.

| `-style` | `main.go` |
|----------|-----------|
| `community` (default) | `scaffold.RecipeCommunity()` — landing, guestbook, chat |
| `scratch` | One custom scene and a `greet` action (same program as below) |

Use `-force` to overwrite those three files. Doc URLs for the version in `go.mod`: `portico docs -tag vX.Y.Z`.

### Or write it

```bash
mkdir my-site && cd my-site
go mod init github.com/yourname/my-site
go get github.com/MrBlindMouse/Portico
```

```go
// main.go
package main

import (
	"log"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/experience"
)

func main() {
	exp, err := portico.New().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string {
				msg := ""
				if flash, ok := ctx.Local("flash"); ok {
					if s, ok := flash.(string); ok {
						msg = s + "\n\n"
					}
				}
				return msg + "Hello, " + ctx.Identity().String() + "\n\nType: greet\n"
			},
			Inputs: map[string]experience.ActionID{"greet": "greet"},
		}).
		Action("greet", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{
					experience.SetLocalEffect("flash", "you waved"),
				}
			},
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	if err := portico.ServeDev(exp, portico.Options{DataDir: ".portico-data"}); err != nil {
		log.Fatal(err)
	}
}
```

Runnable copy: [examples/from-scratch](examples/from-scratch). Scaffold equivalent: [examples/hello](examples/hello) (`RecipeCommunity`).

```bash
go run .
# in another terminal
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

The loop is always: compose an experience, call `ServeDev` (or `Serve`).

## Rules

1. **Actions return effects.** Use `SetLocalEffect`, `TransitionEffect`, `PersistEffect`, space effects, etc. Do not mutate session or space state yourself.
2. **Scenes only render.** `Render` receives a read-only context. State changes happen in actions.
3. **Use typed IDs.** `experience.SceneID`, `ActionID`, and `SpaceID`.
4. **`Build()` validates.** Missing scenes or actions fail before the server starts.

## Scaffolds

`portico init` (default) and [examples/hello](examples/hello) are this one-liner:

```go
import "github.com/MrBlindMouse/Portico/scaffold"

exp, err := portico.New().
	Merge(scaffold.RecipeCommunity()).
	Build()
```

`RecipeCommunity` merges guestbook, chat, and a landing menu. Merge other starters the same way: guestbook, chat, landing, registration, nickname, form, blog, lobby, `RecipeAccountsSite`. `portico scaffold <name>` prints the merge line. Later `Merge` wins when ids collide. Copy a scaffold into your repo when you need to change it.

## Options

| Field | Purpose |
|-------|---------|
| `DataDir` | Where identity data and the host key are stored |
| `IdleTimeout` | Disconnect idle sessions. `0` keeps the 30 minute default. Negative disables it |
| `MaxConns` / `MaxConnsPerIP` | Connection limits |
| `AllowKeys` / `DenyKeys` | Key fingerprint allow/deny lists. Setting either list requires a real SSH auth method |
| `ActionRate` / `ActionBurst` | Per-session action throttling. `0` is the default (5/s, burst 10). A negative `ActionRate` turns limiting off |
| `KeepAliveInterval` / `KeepAliveCountMax` | SSH keepalive probes. `0` leaves them off. After `KeepAliveCountMax` failed probes the session closes |
| `AssetDir` | Directory of assets to load and watch |
| `Accounts` | Enable password-based accounts (development). See account-store order below |
| `AccountBackend` | Account store that wins over the other account fields |
| `HashedAccounts` + `DataDir` | Bcrypt account files under the data directory, used when `AccountBackend` is unset |
| `AccountStore` | Dev plaintext store, used when the two rows above are unset |
| `RequireAccount` | Reject guests and keys that are not linked to an account |
| `ActionBefore` / `ActionAfter` | Hooks around action dispatch. They take `experience.SessionView`, so site code does not import `runtime` |
| `InspectAddr` / `StatusAddr` | Localhost HTTP endpoints for inspection and health |
| `Theme` | Session theme (`render.Default` when nil). A non-empty `NO_COLOR` selects `render.Plain` |
| `LogLevel` | `debug`, `info`, `warn`, or `error` — use **debug** to log Kitty/protocol line discards, ineligible PTY lines, command text, and per-command `read_ms` / `render_ms` (`input_cycle`) |

Account stores are chosen in this order: `AccountBackend`, then `HashedAccounts` with `DataDir`, then `AccountStore`, then `Accounts.Accounts`. If hashed accounts are requested and that directory cannot be opened, `Serve` returns the error.

With no account store and no key list, `ssh` with no key still connects. A key allow/deny list, password accounts, or `RequireAccount` turns that off, because SSH clients stop at the first successful method and a successful "none" login would skip the key and the password. A password login is the account. A key stays a key for persistence unless that fingerprint is linked to an account. `RequireAccount` drops guests and unlinked keys.

When the session has no account, the SSH username is the display label for that connection only (`johan` when OpenSSH fills in your local login). It is not stored and it is not unique. A reserved nickname replaces it for the rest of the session. With no username, a guest still shows as `guest` and an unlinked key as `key-` plus the start of the fingerprint.

You can also load options from a JSON config file (`log_level`). Timestamps use sub-second precision on the default logger.

## Going further

- **Hot reload** — Obtain a `*portico.Host` (via `NewHost` or the `OnReady` hook) and call `HotReload` with a new builder. Reloads that would strand live sessions are rejected.
- **Accounts** — Opt-in. Use the registration scaffold and `RegisterAccountEffect`. The built-in store is for development; production sites should use hashed passwords.
- **Input** — Scene inputs match exact lines or the first word of a line. Actions can read the full line with `ctx.InputLine()`.
- **Inspection** — When using `ServeDev`, an HTTP inspect endpoint is available on localhost (default `:9922`). Example: `curl http://127.0.0.1:9922/inspect/snapshot`.

## Rendering

A scene returns one string. The runtime paints that string after each applied action and when the terminal is resized. `ctx.Size()` is the window in columns and rows (80×24 until the client reports a size). `ctx.Theme()` is the session theme.

`render` fits and draws that string. Widths are rune counts. Fit the text, then apply a theme role. The helpers do not skip ANSI codes.

| Helper | Use |
|--------|-----|
| `Wrap`, `Truncate`, `Pad`, `AlignLeft`, `AlignRight`, `AlignCenter` | Fit text |
| `Box`, `Columns`, `Table` | Layout |
| `List` | A window of rows. The selected row starts with `>` |
| `Progress` | A static bar: `[##--] 50%` |
| `SpinnerFrame` | One glyph for an index the site stores |

`Theme` roles are bold, muted, error, accent, ok, warn, and inverse. `render.Plain` leaves text unchanged.

A selectable list stores an `int` in session-local state. `experience.MoveList` returns the `SetLocal` effect (clamped to the list). `experience.RenderList` only reads it. Bind `up` and `down` on a scene with `Keys: true`. The default key is `experience.LocalListIndex`.

`SpinnerFrame` does not animate. Change the index with `SetLocalEffect` from an action. Portico does not run a clock.

The caret is the full row under that string, prefixed with `>`. To put it somewhere else, draw a `render.Screen` and call `Input` once. `DrawWithin` keeps a region from painting over its neighbor. `BoxBody`, `ListRow`, and `TableCell` return the rectangle of the inside of a box, the selected list line, or a table cell, including when that block is not at the origin. `Show` returns the grid and places the caret. The prompt counts toward the width. One caret per scene. A `Keys` scene that places one fails when it renders.

```go
w, h := ctx.Size()
s := render.NewScreen(w, h)
s.DrawWithin(0, 0, w/2, h-1, history)
s.DrawWithin(0, w/2, w-w/2, h-1, replies)
s.Input(h-1, 0, w/2, ">")
return s.Show(ctx)
```

## Packages

| Import | Use for |
|--------|---------|
| `github.com/MrBlindMouse/Portico` | `New`, `Serve`, `ServeDev`, `Options`, `NewHost` |
| `.../experience` | Scenes, actions, spaces, effects |
| `.../scaffold` | Ready-made starting definitions |
| `.../render` | Layout, theme, list, table, progress (`Box`, `Wrap`, `Table`, `List`) |
| `.../asset` | Loading assets into a builder |

Do not import `internal/...` from site code.

## Production

`ServeDev` is for local work: SSH on **:2222** plus localhost inspect/status defaults. Production sites call **`Serve(addr, exp, opts)`** with the listen address you want. The address is a normal Go `host:port` string (`:22`, `203.0.113.5:22`, `127.0.0.1:2222`, …). Inspect and status bind separately via `Options.InspectAddr` and `Options.StatusAddr` (empty disables that HTTP server).

Wire the SSH address in your site's `main` (flags, env, or config). The library does not add CLI flags to generated sites; a typical pattern:

```go
addr := flag.String("addr", ":2222", "SSH listen address")
flag.Parse()
opts := portico.Options{DataDir: ".portico-data" /* + HashedAccounts, limits, … */}
if err := portico.Serve(*addr, exp, opts); err != nil {
    log.Fatal(err)
}
```

The framework CLI smoke test (`portico serve -demo`) accepts `-addr`, `-inspect`, `-status`, and an optional JSON `-config` file (`listen`, `inspect_addr`, `status_addr`, `data_dir`, …).

### Two SSH servers on port 22

Most hosts already run **OpenSSH** on port 22. You cannot bind two processes to the same **IP:port**. The straightforward layout is:

1. Add a **second public IP** on the machine (or VM).
2. Point a **hostname** at that IP (`play.example.com` → the new address).
3. Keep **sshd** on the primary IP only (`ListenAddress` in `sshd_config`, not `0.0.0.0` if Portico also uses :22).
4. Run Portico on the second IP, still on port 22:

```go
portico.Serve("203.0.113.10:22", exp, opts)
```

Users run `ssh play.example.com` (port 22) for the site and unchanged SSH admin/login on the original host. No alternate client port and no proxy in front of SSH unless you want one.

Use **bcrypt account files** in production (`HashedAccounts` + `DataDir`, or `identity.FileAccountStore` via `AccountBackend`), not the dev plaintext `AccountStore`.

### Privileged port 22 on Linux

Ports below 1024 need extra permission. Prefer the **second-IP** setup above when both services use :22. If Portico is the only listener on :22 (dedicated box or high port elsewhere for admin SSH), build a binary and grant bind capability once:

```bash
go build -o my-site
sudo setcap 'cap_net_bind_service=+ep' ./my-site
./my-site -addr :22   # when main parses -addr and calls Serve
```

Running the binary as root works but is discouraged.

### Operations

- Turn off or keep inspect/status on loopback only unless you add authentication in front.
- Set connection limits, key/account allow lists, and `LogLevel` as needed.
- Use `ShutdownGrace` when a supervisor sends SIGTERM.

See [docs/site-agents.md](docs/site-agents.md) for account-store precedence and agent-facing notes.

## More

- **Manual (full reference):** [docs/manual/README.md](docs/manual/README.md)
- **Building a site:** [docs/building-a-site.md](docs/building-a-site.md)
- **Agents in site repos:** [docs/site-agents.md](docs/site-agents.md)
- Examples: [from-scratch](examples/from-scratch) (write a scene), [hello](examples/hello) (merge a scaffold), [notes](examples/notes) (copy a scaffold and edit it)
- Design and architecture: [DESIGN.md](DESIGN.md)
- Framework contributor notes: [AGENTS.md](AGENTS.md) (not for consumer sites)