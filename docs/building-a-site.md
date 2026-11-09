# Building a Portico SSH site

A Portico site is a normal Go program: compose an **experience** (scenes, actions, optional spaces), then call **`ServeDev`** or **`Serve`**.

## Quick start

Same two starts as the [README](../README.md#create-a-site).

Generate a site (`init` runs `go mod init` when `go.mod` is missing, then `go get`):

```bash
mkdir my-site && cd my-site
go run github.com/MrBlindMouse/Portico/cmd/portico@latest init -module github.com/you/my-site
go run .
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

Default style is `community` (`RecipeCommunity`), the same graph as [examples/hello](../examples/hello). `-style scratch` writes the same program as [examples/from-scratch](../examples/from-scratch). [examples/notes](../examples/notes) is a scaffold copied into the site and edited.

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
  main.go          # ServeDev + options
  site/            # optional: BuildExperience() etc.
  .portico-data/   # host key, identity KV (gitignore)
  AGENTS.md        # agent contract for this repo (from init)
```

Keep `main` thin; grow a `site` package when definitions get large.

## Authoring checklist

1. Pick **root scene** (`Root("home")` or scaffold default).
2. Define **scenes** with `Render` + `Inputs`.
3. Define **actions** returning `[]experience.Effect`.
4. Optional **spaces** for shared multi-user state.
5. **`Build()`** — fix validation errors (missing action refs, bad transitions).
6. **`ServeDev`** with `DataDir`; smoke-test over SSH.

## Scaffolds and recipes

```go
import "github.com/MrBlindMouse/Portico/scaffold"

exp, err := portico.New().
    Merge(scaffold.RecipeCommunity()).
    Build()
```

To customize behavior, **copy** the scaffold definitions into your repo and edit them there ([examples/notes](../examples/notes)). Do not patch files inside the module cache.

`portico scaffold recipe community` prints the merge line.

## Options you will actually use

| Field | Notes |
|-------|--------|
| `DataDir` | Persistence + host key |
| `LogLevel` | `debug` for input timing / protocol noise |
| `InspectAddr` | Default `127.0.0.1:9922` in dev |
| `AssetDir` | For `asset` package + blog scaffold |
| `Accounts` | Registration scaffold; not for production as plaintext |

## Hot reload

Advanced: `portico.NewHost(exp, opts)`, then `host.HotReload(builder)` with a newly built experience. Invalid graphs or sessions stranded in removed scenes or spaces are rejected. Use `POST /inspect/reload-preview` before applying.

## Inspection

While the server runs (localhost only by default):

```bash
curl -s http://127.0.0.1:9922/inspect/snapshot
curl -s http://127.0.0.1:9922/inspect/session/<id>
```

## Examples in the module

| Example | Teaches |
|---------|---------|
| [from-scratch](../examples/from-scratch) | Custom scene + action + effects |
| [hello](../examples/hello) | `RecipeCommunity` unchanged |
| [notes](../examples/notes) | Guestbook copied into the site and edited |

Paths above live in the Portico repo at the **tag matching your `go.mod`** version.

## AI agents

Put [site-agents.md](site-agents.md) in your workflow: site repos should include a short **AGENTS.md** (see `portico init`) that points agents at the tagged upstream contract and local run commands.

## More

- [README.md](../README.md) — API tables and production
- [scaffold/AGENTS.md](../scaffold/AGENTS.md) — scaffold catalog
- [experience/AGENTS.md](../experience/AGENTS.md) — effects and contexts
- [DESIGN.md](../DESIGN.md) — architecture (vision; not all items are shipped)
