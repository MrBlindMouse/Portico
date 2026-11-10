# Portico site — agent contract

You are in a **consumer repo**: a Go module that imports `github.com/MrBlindMouse/Portico`. All domain logic (scenes, actions, spaces, wiring) lives **here**. This file is the starting point for building and changing the site.

## Plans — convenient starting points

`init` gives one scene. **Plans** are the usual next step: `portico generate` copies runnable starter Go into this repo (you own and edit it). Each plan is a small whole site with its own root. Wire it with `Merge` before `Build()`.

```bash
portico generate guestbook
portico generate -list    # landing, guestbook, chat, register, form, blog, lobby
```

The CLI prints a merge line, e.g. `portico.New().Merge(Guestbook()).Build()`. Add that to your builder in `main.go`, then customize the generated file. Later `Merge` calls win when ids collide. `guestbook` and `chat` also write `rest.go` (use `-force` for the second if the file exists).

| Name | Good for |
|------|----------|
| `landing` | Entrance menu — fill the destinations map in `main` |
| `guestbook` | Shared signature board |
| `chat` | Shared chat room |
| `register` | Password signup (turn on accounts in `Options`) |
| `form` | Two-step form demo |
| `blog` | List and open post assets (`AssetDir`) |
| `lobby` | Ready / start multiplayer room |

Prefer adapting a plan over building a feature from scratch. Each generated file starts with **How to change this plan** — read that before editing. Combine plans (e.g. `guestbook` + `chat` + `landing`) when the site needs more than one surface. Chat shows `nick` and `leave`; lobby shows `SpaceCAS`; guestbook, form, and blog show `Persist` for a key or account.

## Pin the framework version

Read `require github.com/MrBlindMouse/Portico` in **go.mod** and use that **same tag** for every upstream link and for `go doc`. Do not rely on training data or on `DESIGN.md` (vision only; not shipped until it appears in godoc or the manual at your tag).

Replace `<tag>` below with your module version (e.g. `v0.2.0`).

| Need | Upstream (at `<tag>`) |
|------|------------------------|
| Full reference | `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/manual/README.md` |
| Extended agent notes | `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/site-agents.md` |
| Human how-to | `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/building-a-site.md` |
| Production / SSH on :22 | `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/manual/07-production.md` |
| Offline API | `go doc github.com/MrBlindMouse/Portico`, `…/components`, `…/installation` |

## Mental model

```text
components     scenes, actions, spaces, effects, contexts (you write these)
installation     portico.New(), Merge, Build() → validated Installation
portico.Options  server config in main (data dir, auth, limits, inspect) — not the graph
Serve(addr, inst, opts)   runs SSH until interrupt
```

- **`Build()`** — graph only: root scene, actions, inputs, spaces. Fails fast on bad references.
- **`Options`** — persistence, accounts, assets dir, rate limits, inspect/status addresses, hooks. Set in **`main.go`** alongside the installation (see below).
- **State changes** — actions return `[]components.Effect` only. Scenes **render** with `components.RenderContext` and do not mutate session or space state.

## What to edit in this repo

| Piece | Where |
|-------|--------|
| Installation (scenes, actions, merges) | `main.go` or a `site/` package |
| Server config | `portico.Options` in `main.go` (`-addr`, `PORTICO_DATA`, …) |
| Starter patterns | Generated plan files (see **Plans** above) |
| Tests | This module; use `testkit` for effect paths without SSH |

Entry pattern (matches `main.go` from `portico init`):

```go
addr := flag.String("addr", ":2222", "SSH listen address")
// … portico.New().Merge(Guestbook()). … .Build() → inst
opts := portico.Options{
    DataDir:     dataDir, // default ".portico-data"; honor PORTICO_DATA
    InspectAddr: "127.0.0.1:9922",
    StatusAddr:  "127.0.0.1:9923",
}
portico.Serve(*addr, inst, opts)
```

`portico.ServeDev(inst, opts)` is a shortcut for local work (SSH `:2222`, loopback inspect/status defaults). Production still builds `Options` explicitly and calls `Serve`.

## Hard rules

1. Actions return **effects** only (`SetLocalEffect`, `TransitionEffect`, space effects, `PersistEffect`, …). Never touch session or space state outside that model.
2. Scenes **only render**; use `ActionContext` / effects for writes.
3. Use typed IDs: `components.SceneID`, `ActionID`, `SpaceID`.
4. Fix **`Build()`** errors before debugging SSH behavior.
5. **Imports for site code:** `portico`, `components`, `installation`, `render`, `asset`, `identity`, `testkit` (tests). Do not import framework `runtime` or `inspect` packages.
6. Generated plan Go stays in **this repo** — merge it; do not import plan sources from the framework module.

## Input and effects (essentials)

- Scene `Inputs`: exact line or **first word** (`register alice pass` → key `register`). Multi-word payloads: `ctx.InputLine()` in the action.
- Common effects: `SetLocalEffect`, `TransitionEffect`, `PersistEffect`, `JoinSpaceEffect`, `BroadcastEffect`, `SpaceSetEffect`, `AppendSpaceListEffect`, `RegisterAccountEffect` (needs account store in `Options`), `ReserveNicknameEffect` (identity; no separate plan).
- Cross-user updates: write space data, then `BroadcastEffect(spaceID)`.

Rendering: optional `render` helpers (`Box`, `Wrap`, `Table`, `List`, …). See manual **Rendering** at `<tag>`.

## Verify locally

```bash
go test ./...
go run .
# other terminal
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
# optional (loopback only; no auth)
curl -s http://127.0.0.1:9922/inspect/snapshot
```

`PORTICO_DATA` overrides the data directory when `main` reads it. `PORTICO_LOG_LEVEL=debug` when wired in `main`.

Inspect HTTP is **unauthenticated**. Keep inspect/status on loopback in production or disable them (`InspectAddr` / `StatusAddr` empty with `Serve`).

## Testing

Prefer `Build()` validation and pure helpers first. Integration without SSH: `github.com/MrBlindMouse/Portico/testkit` (`testkit.New(inst)`). Details: manual **Testing** at `<tag>`.

## When you need more

Use the upstream links in **Pin the framework version**. **`docs/site-agents.md`** at your tag adds inspect endpoints, list/caret rendering, identity/auth edge cases, and production account stores — read it when this file is not enough.
