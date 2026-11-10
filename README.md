# Portico

Build interactive, multi-user terminal sites over SSH.

You define **components** (scenes, actions, spaces) and `Build()` an **installation**. Portico handles SSH, sessions, input, shared state, and persistence.

## Create a site

A Portico site is an ordinary Go program that imports Portico as a library. Bootstrap with `portico init`, or write `main.go` yourself. Both end the same way: `go run .`, then SSH to port 2222.

**`portico init`** creates the Go module and a minimal runnable site (one scene, one action). For starter patterns (guestbook, chat, landing, …), see **Plans** (next).

### Bootstrap (`portico init`)

From an empty directory:

```bash
mkdir my-site && cd my-site
go run github.com/MrBlindMouse/Portico/cmd/portico@latest init -module github.com/yourname/my-site
go run .
```

`init` creates `go.mod` (if needed), runs `go get`, and writes `main.go`, `AGENTS.md`, and `README.md`. That `main.go` is one `home` scene and a `greet` action (see below). Use `-force` to overwrite those three files. Doc URLs for the version in `go.mod`: `portico docs -tag vX.Y.Z`.

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
	"flag"
	"log"
	"os"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/components"
)

func main() {
	addr := flag.String("addr", ":2222", "SSH listen address")
	flag.Parse()

	inst, err := portico.New().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string {
				msg := ""
				if flash, ok := ctx.Local("flash"); ok {
					if s, ok := flash.(string); ok {
						msg = s + "\n\n"
					}
				}
				return msg + "Hello, " + ctx.Identity().String() + "\n\nType: greet\n"
			},
			Inputs: map[string]components.ActionID{"greet": "greet"},
		}).
		Action("greet", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{
					components.SetLocalEffect("flash", "you waved"),
				}
			},
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	dataDir := ".portico-data"
	if v := os.Getenv("PORTICO_DATA"); v != "" {
		dataDir = v
	}
	opts := portico.Options{
		DataDir:     dataDir,
		InspectAddr: "127.0.0.1:9922",
		StatusAddr:  "127.0.0.1:9923",
	}
	if err := portico.Serve(*addr, inst, opts); err != nil {
		log.Fatal(err)
	}
}
```

This matches what `portico init` writes: build the installation in code, set **server** `Options` in `main`, then `Serve(addr, inst, opts)`. `ServeDev` is a shorthand when you only need defaults (see **Options**).

```bash
go run .
# in another terminal
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

The loop is always: `Build()` an installation, configure `Options`, call `Serve` (or `ServeDev` for a minimal dev path).

## Plans

Starter patterns live in your repo as ordinary Go, not as an importable plan package. After `init` (or once you have `go.mod` and `main.go`), run `portico generate` in the site directory:

```bash
portico generate guestbook
portico generate -list    # landing, guestbook, chat, register, form, blog, lobby
```

Each command copies one file (and `rest.go` for `guestbook` or `chat`) and prints a merge line such as `portico.New().Merge(Guestbook()).Build()`. Edit the generated code, then wire that `Merge` into `main.go`.

Each plan is a small runnable installation with its own root scene. Combine several with `Merge`; later merges win when scene, action, or space ids collide. Existing files are kept unless you pass `-force` (needed when the second of `guestbook`/`chat` would overwrite an existing `rest.go`). See the manual for `-pkg` and per-plan notes.

`landing` takes a destinations map you fill in `main`. The `register` plan needs password accounts enabled in `Options` (below).

## Rules

1. **Actions return effects.** Use `SetLocalEffect`, `TransitionEffect`, `PersistEffect`, space effects, etc. Do not mutate session or space state yourself.
2. **Scenes only render.** `Render` receives a read-only context. State changes happen in actions.
3. **Use typed IDs.** `components.SceneID`, `ActionID`, and `SpaceID`.
4. **`Build()` validates.** Missing scenes or actions fail before the server starts.

## Options

`portico.Options` is the **server configuration** for `Serve`, `ServeDev`, and `NewHost`. It is separate from the installation you `Build()`: persistence, SSH auth, limits, optional inspect/status HTTP, assets on disk, logging, and action hooks all come from `Options`.

`ServeDev(inst, opts)` is for local work. It listens on **:2222** and, when `InspectAddr` / `StatusAddr` are empty, binds inspect and status to loopback defaults. `Serve(addr, inst, opts)` uses your SSH listen address; inspect and status run only if you set those fields (empty means off).

Sites usually assemble `opts` in `main` from flags and environment variables. You can overlay a JSON config file with `opts.MergeConfig` (same shape as `portico serve -demo -config`). `Listen` in that file applies to `Serve(addr, …)`, not to a field on `Options`.

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
| `ActionBefore` / `ActionAfter` | Hooks around action dispatch (`components.SessionView`) |
| `InspectAddr` / `StatusAddr` | Localhost HTTP endpoints for inspection and health |
| `Theme` | Session theme (`render.Default` when nil). A non-empty `NO_COLOR` selects `render.Plain` |
| `LogLevel` | `debug`, `info`, `warn`, or `error` — use **debug** to log Kitty/protocol line discards, ineligible PTY lines, command text, and per-command `read_ms` / `render_ms` (`input_cycle`) |

Account stores are chosen in this order: `AccountBackend`, then `HashedAccounts` with `DataDir`, then `AccountStore`, then `Accounts.Accounts`. If hashed accounts are requested and that directory cannot be opened, `Serve` returns the error.

With no account store and no key list, `ssh` with no key still connects. A key allow/deny list, password accounts, or `RequireAccount` turns that off, because SSH clients stop at the first successful method and a successful "none" login would skip the key and the password. A password login is the account. A key stays a key for persistence unless that fingerprint is linked to an account. `RequireAccount` drops guests and unlinked keys.

When the session has no account, the SSH username is the display label for that connection only (`johan` when OpenSSH fills in your local login). It is not stored and it is not unique. A reserved nickname replaces it for the rest of the session. With no username, a guest still shows as `guest` and an unlinked key as `key-` plus the start of the fingerprint.

## Going further

| Topic | Summary |
|-------|---------|
| [Authoring](docs/manual/02-authoring.md) | Scenes, actions, `Merge`, prefix input, `InputLine()` |
| [Identity & accounts](docs/manual/05-identity-accounts.md) | Guests, keys, nicknames, `register` plan, hashed accounts in production |
| [Testing](docs/manual/08-testing.md) | `testkit` exercises real apply logic without opening SSH |
| [Hot reload](docs/manual/appendices/hot-reload.md) | `NewHost` + `HotReload`; preview with `POST /inspect/reload-preview` before swapping |
| [Manual](docs/manual/README.md) | Full CLI, effects, wire, inspect paths |

Inspect HTTP has no authentication. Keep it on loopback in production unless something else protects the port.

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

A selectable list stores an `int` in session-local state. `installation.MoveList` returns the `SetLocal` effect (clamped to the list). `installation.RenderList` only reads it. Bind `up` and `down` on a scene with `Keys: true`. The default key is `installation.LocalListIndex`.

`SpinnerFrame` does not animate. Change the index with `SetLocalEffect` from an action. Portico does not run a clock.

The caret is the full row under that string, prefixed with `>`. A successful command redraws the scene and leaves a fresh `>`. `ctx.RecentCommands` is the transcript if a scene wants to draw one; it is not printed for you. To put the caret somewhere else, draw a `render.Screen` and call `Input` once. `DrawWithin` keeps a region from painting over its neighbor. `BoxBody`, `ListRow`, and `TableCell` return the rectangle of the inside of a box, the selected list line, or a table cell, including when that block is not at the origin. `Show` returns the grid and places the caret. The prompt counts toward the width. One caret per scene. A `Keys` scene that places one fails when it renders.

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
| `.../components` | Scenes, actions, spaces, effects, contexts |
| `.../installation` | Builder, `Installation`, wire, form/list helpers |
| `.../render` | Layout, theme, list, table, progress (`Box`, `Wrap`, `Table`, `List`) |
| `.../asset` | Loading assets into a builder |
| `.../identity` | Custom stores when you implement `identity.Store` / `AccountStore` |
| `.../testkit` | In-memory engine fixtures for site tests |

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
- Design and architecture: [DESIGN.md](DESIGN.md)
- Framework contributor notes: [AGENTS.md](AGENTS.md) (not for consumer sites)