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
| `IdleTimeout` | Disconnect idle sessions |
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
| `LogLevel` | `debug`, `info`, `warn`, or `error` — use **debug** to log Kitty/protocol line discards, ineligible PTY lines, command text, and per-command `read_ms` / `render_ms` (`input_cycle`) |

Account stores are chosen in this order: `AccountBackend`, then `HashedAccounts` with `DataDir`, then `AccountStore`, then `Accounts.Accounts`.

With no account store and no key list, `ssh` with no key still connects as a guest. A key allow/deny list, password accounts, or `RequireAccount` turns that off, because SSH clients stop at the first successful method and a successful "none" login would skip the key and the password. A key login stays a key unless that fingerprint is linked to an account. A password login is the account. `RequireAccount` drops guests and unlinked keys.

You can also load options from a JSON config file (`log_level`). Timestamps use sub-second precision on the default logger.

## Going further

- **Hot reload** — Obtain a `*portico.Host` (via `NewHost` or the `OnReady` hook) and call `HotReload` with a new builder. Reloads that would strand live sessions are rejected.
- **Accounts** — Opt-in. Use the registration scaffold and `RegisterAccountEffect`. The built-in store is for development; production sites should use hashed passwords.
- **Input** — Scene inputs match exact lines or the first word of a line. Actions can read the full line with `ctx.InputLine()`.
- **Inspection** — When using `ServeDev`, an HTTP inspect endpoint is available on localhost (default `:9922`). Example: `curl http://127.0.0.1:9922/inspect/snapshot`.

## Packages

| Import | Use for |
|--------|---------|
| `github.com/MrBlindMouse/Portico` | `New`, `Serve`, `ServeDev`, `Options`, `NewHost` |
| `.../experience` | Scenes, actions, spaces, effects |
| `.../scaffold` | Ready-made starting definitions |
| `.../render` | Simple layout helpers (`Box`, `Columns`) |
| `.../asset` | Loading assets into a builder |

Do not import `internal/...` from site code.

## Production

Dev mode listens on port 2222. For port 22 on Linux:

```bash
go build -o my-site
sudo setcap 'cap_net_bind_service=+ep' ./my-site
./my-site
```

Do not expose the inspect or status HTTP endpoints without authentication.

## More

- **Building a site:** [docs/building-a-site.md](docs/building-a-site.md)
- **Agents in site repos:** [docs/site-agents.md](docs/site-agents.md)
- Examples: [from-scratch](examples/from-scratch) (write a scene), [hello](examples/hello) (merge a scaffold), [notes](examples/notes) (copy a scaffold and edit it)
- Design and architecture: [DESIGN.md](DESIGN.md)
- Framework contributor notes: [AGENTS.md](AGENTS.md) (not for consumer sites)