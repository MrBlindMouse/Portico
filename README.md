# Portico

Build interactive, multi-user terminal sites over SSH.

You write the experience (scenes, actions, spaces). Portico handles SSH, sessions, input, shared state, and persistence.

## Create a site

A Portico site is an ordinary Go program that imports Portico as a library.

```bash
mkdir my-site
cd my-site
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
				return "Hello, " + ctx.Identity().String() + "\n\nType: greet\n"
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

```bash
go run .
# in another terminal
ssh -p 2222 localhost
```

That’s the whole loop: `go get` the framework, compose an experience, call `ServeDev` (or `Serve`).

## Rules

1. **Actions return effects.** Use `SetLocalEffect`, `TransitionEffect`, `PersistEffect`, space effects, etc. Do not mutate session or space state yourself.
2. **Scenes only render.** `Render` receives a read-only context. State changes happen in actions.
3. **Use typed IDs.** `experience.SceneID`, `ActionID`, and `SpaceID`.
4. **`Build()` validates.** Missing scenes or actions fail before the server starts.

## Scaffolds

Start from a working pattern instead of a blank file:

```go
import "github.com/MrBlindMouse/Portico/scaffold"

exp, err := portico.New().
	Merge(scaffold.Guestbook()).
	Merge(scaffold.Landing(map[string]experience.SceneID{
		"guestbook": "guestbook",
	})).
	Build()
```

Available scaffolds include guestbook, chat, landing, registration, and nickname.

## Options

| Field | Purpose |
|-------|---------|
| `DataDir` | Where identity data and the host key are stored |
| `IdleTimeout` | Disconnect idle sessions |
| `MaxConns` / `MaxConnsPerIP` | Connection limits |
| `AllowKeys` / `DenyKeys` | Key fingerprint allow/deny lists |
| `ActionRate` / `ActionBurst` | Per-session action throttling |
| `AssetDir` | Directory of assets to load and watch |
| `Accounts` | Enable password-based accounts (development) |
| `InspectAddr` / `StatusAddr` | Localhost HTTP endpoints for inspection and health |
| `LogLevel` | `debug`, `info`, `warn`, or `error` — use **debug** to log Kitty/protocol line discards, ineligible PTY lines, and per-command `read_ms` / `render_ms` (`input_cycle`) |

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

- Design and architecture: [DESIGN.md](DESIGN.md)
- Framework contributor / agent notes: [AGENTS.md](AGENTS.md)
- Reference site: [examples/hello](examples/hello)