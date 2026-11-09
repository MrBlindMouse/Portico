# AGENTS.md

Ops contract for agents changing **this repository** (the Portico framework).

Building an SSH site is a different job: import `github.com/MrBlindMouse/Portico` in **your** module. Start with `portico init` and [docs/site-agents.md](docs/site-agents.md). Do not add site scenes to this tree. `examples/` are reference consumers, not the product.

**DESIGN.md is the vision contract; "Planned" there is not implemented until shipped.**

## Audience

| Who | Owns |
|-----|------|
| Framework (`experience`, `runtime`, `portico`, …) | Layers, invariants, extension points |
| Site author (`examples/*`, user repos) | Scenes, actions, domain, `main` + `Serve`; see `docs/site-agents.md` |
| `cmd/portico` | Operator helpers; **not** the primary site entry (use `-demo` only for bundled smoke test) |

## Run and test

These commands are for **this repo** (framework CI and the bundled examples). A site repo runs `go run .` in its own module.

```bash
go test ./...
go vet ./...
go run ./examples/hello   # reference consumer, SSH on :2222
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

CI runs `go test`, `go vet`, and `internal/archtest` on every push.

## Principles (ceilings and non-goals)

- Session and space mutation happens only in `runtime/apply` — not in actions, scenes, or site code.
- `runtime.Session` fields are unexported; construct sessions only inside `runtime`.
- `experience` must not import `runtime`.
- Single-process scale: harden locks/metrics locally; clustering and real-time SLOs are out of scope (DESIGN §3).
- Dev account auth may use plaintext `runtime.AccountStore`; production sites opt into `identity.FileAccountStore` (bcrypt).
- No domain logic in `runtime` — only orchestration.

## Architecture map

| Package       | Role |
|---------------|------|
| `portico`     | Public `New()`, `Serve()`, `ServeDev()`, `NewHost()`, `Host.HotReload`, `Options` |
| `experience`  | Scenes, actions, spaces, builder, effects (descriptions) |
| `runtime`     | SSH server, sessions, dispatch, apply, spaces registry |
| `identity`    | SSH fingerprint / account identity, file-backed KV, nicknames |
| `render`      | Terminal layout, theme roles, list/table/progress helpers |
| `inspect`     | Read-only JSON snapshots for agents |
| `scaffold`    | Composable starter definitions (optional for sites) |
| `asset`       | Named filesystem assets and poll watcher |
| `testkit`     | In-memory engine/store fixtures for tests (`runtime_test` and sites; not from `package runtime` tests) |
| `cmd/portico` | CLI: `init`, `docs`, `inspect`, `scaffold`, `wire`, `identity-export`, `account-export`, `serve -demo` |
| `internal/*`  | archtest, faketerm, fakessh — not importable by sites |

## Gotchas

- Dev server uses an ephemeral or data-dir host key; CI uses high port only (no :22).
- Hot-reload swaps the whole validated `Experience`; invalid reloads and stranded sessions are rejected.
- Space fan-out uses in-memory locks (ponytail: not suitable for huge concurrency).
- Inspect HTTP binds localhost by default — do not expose without auth.
- `Render` uses `RenderContext` only; effect helpers on `SceneContext` apply in actions, not during render.
- Prefix input dispatch: first token of a line can match scene input keys (`register user pass`).

## Child AGENTS index

- `runtime/AGENTS.md` — apply/dispatch invariants
- `experience/AGENTS.md` — builder validation, contexts, effects
- `scaffold/AGENTS.md` — composing scaffolds into sites

## Shipped vs planned (v0.2)

**v0.1 baseline (shipped):** experience builder, effects/apply, SSH runtime, identity KV, spaces, inspect HTTP, scaffolds (landing/guestbook/chat/registration/nickname), assets, hot-reload, status/health, rate limits, dev plaintext accounts.

**v0.2 (shipped in tree):** reload preview inspect endpoint, richer snapshots, form/theme scaffolds, `ShutdownGrace` drain, `/ready` + `/metrics`, `identity.AccountStore` + `FileAccountStore`, session lifecycle effects, action hooks, blog/lobby/recipe scaffolds, space/scene validators, `portico wire` CLI.

**Still out of scope:** clustered spaces, in-process event bus, SMTP/email verification in core, interpreted JSON experiences.

**v0.2 acceptance:** each phase ships tests (`go test ./...`, `go vet`, `internal/archtest`) and README touch where operators/agents are affected.

**Observability:** structured `slog` for session connect/disconnect, action dispatch, apply/dispatch errors; `config.NewLogger` + `-log-level` / config `log_level`.
