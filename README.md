# Portico

Go framework for interactive, multi-user terminal **sites** served over SSH. Sites are ordinary modules that import this library; **this repository is the framework**, not a site.

Authors define **components** (scenes, actions, spaces), `Build()` an **installation**, and call `Serve` with **Options**. The framework runs SSH, sessions, input dispatch, effect application, spaces, and identity persistence. Vision and ceilings: [DESIGN.md](DESIGN.md). Day-to-day rules for changing this tree: [AGENTS.md](AGENTS.md).

## Verify

```bash
go test ./...
go vet ./...
go test ./internal/archtest/...
go run ./examples/hello   # smoke test: every plan behind landing; SSH :2222
```

```bash
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

`examples/hello` is the framework **smoke test**. It merges every plan from `internal/plans`; `go test ./examples/hello` hits each landing destination through the real dispatch path. CI and local `go test ./...` rely on it. Do not treat it as a site template.

## Layout

| Package | Role |
|---------|------|
| `portico` | Public `New()`, `Serve`, `ServeDev`, `NewHost`, `Options` |
| `components` | Scenes, actions, spaces, effects, contexts |
| `installation` | Builder, validated `Installation`, wire, form/list helpers |
| `runtime` | SSH server, sessions, dispatch, apply (sites do not import) |
| `identity` | Fingerprints, KV, nicknames, account stores |
| `render` | Terminal layout and theme helpers |
| `inspect` | Read-only HTTP snapshots (unauthenticated) |
| `asset` | Filesystem assets and poll watcher |
| `testkit` | In-memory engine fixtures |
| `cmd/portico` | `init`, `generate`, `docs`, `wire`, `inspect`, `serve -demo`, … |
| `internal/plans` | Sources for `portico generate` (embedded in CLI; not imported by sites) |

Package-level notes: [AGENTS.md](AGENTS.md) (architecture map and child AGENTS index).

## Working here

**Plans** — Starters shipped to sites via `portico generate`. Sources live in [`internal/plans/`](internal/plans/); sync embed with `go generate ./cmd/portico`. Each plan sets its own root and builds alone. See [`internal/plans/AGENTS.md`](internal/plans/AGENTS.md).

**CLI** — `cmd/portico` writes site scaffolds (`init`), copies plans (`generate`), and runs operator helpers. `portico serve` requires `-demo` (smoke test merging guestbook/chat/landing). Real sites use `go run .` in their own module.

**Invariants** (summary; details in AGENTS):

- Session and space mutation only in `runtime/apply`.
- `installation` must not import `runtime`.
- `components` must not import `installation` or `runtime`.

**Regenerate plan embeds** after editing `internal/plans/*.go`:

```bash
go generate ./cmd/portico
go test ./cmd/portico/ -run TestPlanEmbedMatchesSource
```

## Public API (sites)

Sites import `github.com/MrBlindMouse/Portico` and typically `components`, `installation`, `render`, and optionally `identity`, `asset`, `testkit`. Full reference: [docs/manual/README.md](docs/manual/README.md). Option fields and production deployment: [docs/manual/reference/options.md](docs/manual/reference/options.md), [docs/manual/07-production.md](docs/manual/07-production.md).

## For site authors

This repo is the wrong place to build a product. In a **new module**:

```bash
go run github.com/MrBlindMouse/Portico/cmd/portico@latest init -module github.com/you/my-site
```

| Doc | Audience |
|-----|----------|
| Site **AGENTS.md** / **README.md** (from `init`) | Agents and operators in that repo |
| [docs/building-a-site.md](docs/building-a-site.md) | Human how-to |
| [docs/site-agents.md](docs/site-agents.md) | Extended agent contract |
| [docs/manual/](docs/manual/README.md) | Full manual at your `go.mod` tag |

```bash
portico docs -tag vX.Y.Z
```
