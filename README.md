# Portico

Go framework for building your own interactive, multi-user terminal sites over SSH. You own the experience definition (scenes, actions, spaces); Portico owns SSH, sessions, dispatch, shared state, and persistence.

Vision and layer model: [DESIGN.md](DESIGN.md). Maintainer/agent ops: [AGENTS.md](AGENTS.md).

## Build a site

A Portico site is a normal Go program: compose an `experience`, then call `Serve`.

```go
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
				return "Hello, " + ctx.Identity().String() + "\n\nType: greet"
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

Reference implementation with scaffolds: [examples/hello](examples/hello).

### Rules site code must follow

1. **Actions return effects** — use `SetLocalEffect`, `TransitionEffect`, `PersistEffect`, space effects, etc. Do not mutate session or space state yourself.
2. **Scenes render only** — `Render` receives `RenderContext` (read-only). Use actions for state changes.
3. **Typed IDs** — `experience.SceneID`, `ActionID`, `SpaceID` in builder and maps.
4. **Validate at `Build()`** — dangling action/scene refs fail before listen.

### Compose scaffolds

```go
portico.New().
	Merge(scaffold.Guestbook()).
	Merge(scaffold.Landing(map[string]experience.SceneID{"guestbook": "guestbook"})).
	Build()
```

Hints: `go run ./cmd/portico scaffold guestbook`

### Serve options

| Field | Purpose |
|-------|---------|
| `DataDir` | File-backed identity KV + host key |
| `InspectAddr` / `StatusAddr` | Localhost HTTP for agents (`127.0.0.1` defaults in `ServeDev`) |
| `IdleTimeout` | Disconnect after idle (resets on input) |
| `MaxConns` / `MaxConnsPerIP` | Connection limits |
| `AllowKeys` / `DenyKeys` | SSH key allow/deny (fingerprints) |
| `ActionRate` / `ActionBurst` | Per-session action throttle |
| `AssetDir` | Load files into live assets at startup; poll for hot swap |
| `AssetPollEvery` | Poll interval (default 1s) |
| `OnReady` | Hook with `*portico.Host` before `Serve` listens (hot reload, admin hooks) |

JSON config: `config.Load` + `opts.MergeConfig(cfg)` — see `config.File`.

### Hot reload

`Serve` blocks, so obtain a `*portico.Host` with `NewHost`, or use `OnReady`:

```go
host, err := portico.NewHost(":2222", exp, portico.Options{
    OnReady: func(h *portico.Host) {
        go func() {
            // e.g. on SIGHUP or admin HTTP: h.HotReload(patch)
        }()
    },
})
if err != nil { log.Fatal(err) }
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()
return host.Server.ListenAndServe(ctx)
```

Or: `portico.HotReload(host.Server, patchedBuilder)`. Reload rejects definitions that would leave live sessions on removed scenes. Wire JSON via inspect only **validates** wiring (`POST /inspect/validate`); full reload needs a `Builder` with your `Render`/`Run` funcs.

### Accounts (development)

Password SSH auth is opt-in (`Options.Accounts` / `accounts_enabled` in config). Sessions get `identity.FromAccount(username)`. Registration scaffold: `scaffold.Registration()` plus `RegisterAccountEffect`. **Dev store is plaintext** — hash passwords in production sites.

### Input commands

Scene `Inputs` map **exact lines** or **prefix** (first word): `register alice secret` matches input key `register`. Actions can read the full line with `ctx.InputLine()`.

## Quick start (demo)

```bash
go run ./examples/hello
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

Inspect: `curl http://127.0.0.1:9922/inspect/snapshot`

The `portico` CLI is for operators (`inspect`, `identity-export`, scaffold hints). To run the bundled demo only: `go run ./cmd/portico serve -demo`.

## Package map (for site authors)

| Import | Use |
|--------|-----|
| `github.com/.../Portico` (`portico`) | `New()`, `Serve()`, `ServeDev()`, `NewHost()`, `Options`, `HotReload` |
| `.../experience` | Scenes, actions, spaces, effects, builder |
| `.../scaffold` | Optional starter definitions to `Merge` |
| `.../render` | `Box`, `Columns` layout helpers |
| `.../asset` | `MergeAssetsIntoBuilder` |
| `.../runtime` | Advanced: custom `Serve` wiring; most sites only need `HotReload` |

Do **not** import `internal/*` from site code.

## Test

```bash
go test ./...
go vet ./...
```

## Production SSH (:22)

Dev and CI use port **2222**. On Linux:

```bash
sudo setcap 'cap_net_bind_service=+ep' ./your-site-binary
./your-site-binary   # listen with portico.Serve(":22", ...)
```

Do not expose inspect/status HTTP without auth.
