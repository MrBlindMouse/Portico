# Building a Portico SSH site

A Portico site is a normal Go program: compose an **installation** from **components** (scenes, actions, optional spaces), set **`portico.Options`** in `main`, then call **`Serve(addr, inst, opts)`**. The map is [the manual overview](manual/00-overview.md).

## Quick start

```bash
mkdir my-site && cd my-site
go run github.com/MrBlindMouse/Portico/cmd/portico@latest init -module github.com/you/my-site
go run .
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

`init` writes `main.go` (one scene, `greet`, `Serve` + `-addr` + loopback inspect), **AGENTS.md**, and **README.md**. Add plans from the site directory:

```bash
go run github.com/MrBlindMouse/Portico/cmd/portico generate <name>
go run github.com/MrBlindMouse/Portico/cmd/portico generate -list
```

## Mental model

```
SSH connect → session in root scene → user types line
    → dispatcher matches scene Inputs → action Run() → effects
    → runtime apply → re-render scene
```

Session and space mutation happen only inside the runtime **apply** step. Your actions describe intent via effects.

## Project layout

```
my-site/
  main.go          # Build(), Options, Serve (from portico init)
  *.go             # plans from go run …/cmd/portico generate (you edit)
  .portico-data/   # host key, identity KV (gitignore)
  AGENTS.md        # primary agent contract (from init)
  README.md        # human ops notes (from init)
```

Keep `main` thin; grow a `site` package when definitions get large.

## Authoring checklist

1. Pick **root scene** (`Root("home")`, or the root a generated plan sets, such as `landing`).
2. Define **scenes** with `Render` + `Inputs`.
3. Define **actions** returning `[]components.Effect`.
4. Optional **spaces** for shared multi-user state.
5. **`Build()`** — fix validation errors (missing action refs, bad transitions).
6. **`Serve`** with `Options` (persistence, limits, inspect on loopback in dev); smoke-test over SSH.

## Plans

```bash
go run github.com/MrBlindMouse/Portico/cmd/portico generate guestbook
go run github.com/MrBlindMouse/Portico/cmd/portico generate -list
```

Each command copies starter Go into the site and prints a `Merge(…)` line. Edit the file; wire the merge in `main.go`. Plans are **not** a library—only the copies in your repo count. Names: `landing`, `guestbook`, `chat`, `register`, `form`, `blog`, `lobby`. Each builds on its own; later `Merge` wins when ids collide.

## Options you will actually use

| Field | Notes |
|-------|--------|
| `DataDir` | Persistence + host key |
| `LogLevel` | `debug` for input timing / protocol noise |
| `InspectAddr` | Init template uses `127.0.0.1:9922`; empty disables with `Serve` |
| `AssetDir` | For `asset` package + blog plan |
| `Accounts` / `HashedAccounts` | `generate register`; production uses hashed accounts |
| `Theme` | `render.Theme`. Nil keeps `render.Default` (`render.Plain` if `NO_COLOR` is set) |

## Rendering

Scenes return a string. Read the window with `ctx.Size()` and the theme with `ctx.Theme()`. Draw with `render` (`Wrap`, `Box`, `Table`, `List`, `Progress`, `SpinnerFrame`). Widths count runes.

A menu stores its index in local state. The action returns `installation.MoveList`. The scene calls `installation.RenderList` and binds `up` / `down` with `Keys: true`. See [site-agents.md](site-agents.md#rendering).

The caret is the row under the frame. `render.NewScreen` plus `Input` and `Show` puts that caret in one rectangle. See [site-agents.md](site-agents.md#caret).

## Hot reload

Advanced: `portico.NewHost(addr, inst, opts)`, then `host.HotReload(builder)` with a newly built installation. Invalid graphs or sessions stranded in removed scenes or spaces are rejected. Use `POST /inspect/reload-preview` before applying.

## Inspection

Inspect and status HTTP have no authentication. The init template binds them to loopback. Do not publish that port on the internet.

```bash
curl -s http://127.0.0.1:9922/inspect/snapshot
curl -s http://127.0.0.1:9922/inspect/session/<id>
```

## AI agents

`portico init` writes **AGENTS.md** in the site repo as the primary agent contract. Extended notes: [site-agents.md](site-agents.md). Full reference: [manual/README.md](manual/README.md) at your `go.mod` tag.

## More

- [manual/README.md](manual/README.md) — operator and reference manual
- [manual/07-production.md](manual/07-production.md) — deploy and listen addresses
- [DESIGN.md](../DESIGN.md) — architecture vision (not all items are shipped)
