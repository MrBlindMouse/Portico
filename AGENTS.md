# AGENTS.md

Ops contract for agents changing **this repository** (the Portico framework).

Building an SSH site is a different job: import `github.com/MrBlindMouse/Portico` in **your** module. Run `portico init` (writes **AGENTS.md** + **README.md** in the site); extend with [docs/site-agents.md](docs/site-agents.md) and the manual at your module tag. Do not add site scenes to this tree. `examples/hello` is the framework smoke test (merges every plan; `TestEveryPlanFromLanding` in CI).

**DESIGN.md is the vision contract; "Planned" there is not implemented until shipped.**

## Contributing (framework repo)

- **Bugs and feature requests:** open a [GitHub issue](https://github.com/MrBlindMouse/Portico/issues/new). Include repro steps, `go.mod` tag or commit, and expected vs actual behavior. Do not track site-only bugs here unless the framework API is wrong.
- **Pull requests:** target `main`; one logical change per PR; run `go test ./...`, `go vet ./...`, and `go test ./internal/archtest/...` before push. Touch README or manual only when operators or site authors are affected.
- **Plans:** after editing `internal/plans/*.go`, run `go generate ./cmd/portico` and `TestPlanEmbedMatchesSource`.
- **Scope:** changes here are framework layers and `examples/hello` smoke coverage — not production sites (those live in separate modules).

## Audience

| Who | Owns |
|-----|------|
| Framework (`installation`, `components`, `runtime`, `portico`, …) | Layers, invariants, extension points |
| Site author (user repos) | Components, domain, `main` + `Serve`; site **AGENTS.md** (init), then `docs/site-agents.md` |
| `cmd/portico` | Operator helpers; **not** the primary site entry (use `-demo` only for bundled smoke test) |

## Run and test

These commands are for **this repo** (framework CI and the bundled examples). A site repo runs `go run .` in its own module.

```bash
go test ./...
go vet ./...
go run ./examples/hello   # smoke test: every plan behind landing, SSH on :2222
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

CI runs `go test`, `go vet`, and `internal/archtest` on every push.

## Principles (ceilings and non-goals)

- Session and space mutation happens only in `runtime/apply` — not in actions, scenes, or site code.
- `runtime.Session` fields are unexported; construct sessions only inside `runtime`.
- `installation` must not import `runtime`.
- Single-process scale: harden locks/metrics locally; clustering and real-time SLOs are out of scope (DESIGN §3).
- Dev account auth may use plaintext `runtime.AccountStore`; production sites opt into `identity.FileAccountStore` (bcrypt).
- No domain logic in `runtime` — only orchestration.

## Architecture map

| Package       | Role |
|---------------|------|
| `portico`     | Public `New()`, `Serve()`, `ServeDev()`, `NewHost()`, `Host.HotReload`, `Options` |
| `components`  | Scenes, actions, spaces, IDs, effects, contexts |
| `installation`| Builder, wire, form/list helpers, validated `Installation` |
| `runtime`     | SSH server, sessions, dispatch, apply, spaces registry |
| `identity`    | SSH fingerprint / account identity, file-backed KV, nicknames |
| `render`      | Terminal layout, theme roles, list/table/progress helpers |
| `inspect`     | Read-only JSON snapshots for agents |
| `asset`       | Named filesystem assets and poll watcher |
| `testkit`     | In-memory engine/store fixtures for tests (`runtime_test` and sites; not from `package runtime` tests) |
| `cmd/portico` | CLI: `init`, `generate`, `docs`, `inspect`, `wire`, `identity-export`, `account-export`, `serve -demo` |
| `internal/*`  | archtest, faketerm, fakessh, plans — not importable by sites |

## Gotchas

- Dev server uses an ephemeral or data-dir host key; CI uses high port only (no :22).
- Hot-reload swaps the whole validated `Installation`; invalid reloads and stranded sessions are rejected.
- Space fan-out uses in-memory locks (ponytail: not suitable for huge concurrency).
- Inspect HTTP has no auth. `ServeDev` and init default to loopback; any other address is public.
- `Render` uses `RenderContext` only; effect helpers on `SceneContext` apply in actions, not during render.
- Prefix input dispatch: first token of a line can match scene input keys (`register user pass`).

## Child AGENTS index

- `runtime/AGENTS.md` — apply/dispatch invariants
- `installation/AGENTS.md` — builder and wire
- `components/AGENTS.md` — scenes, actions, spaces, effects, contexts
- `internal/plans/AGENTS.md` — plan sources `portico generate` copies into a site

## Shipped vs planned (v0.2)

**v0.1 baseline (shipped):** installation builder, effects/apply, SSH runtime, identity KV, spaces, inspect HTTP, plans (landing/guestbook/chat/registration), assets, hot-reload, status/health, rate limits, dev plaintext accounts.

**v0.2 (shipped in tree):** reload preview inspect endpoint, richer snapshots, form/blog/lobby plans, `ShutdownGrace` drain, `/ready` + `/metrics`, `identity.AccountStore` + `FileAccountStore`, session lifecycle effects, action hooks, `portico generate` (owned plan source), space/scene validators, `portico wire` CLI.

**Still out of scope:** clustered spaces, in-process event bus, SMTP/email verification in core, interpreted JSON installations.

**v0.2 acceptance:** each phase ships tests (`go test ./...`, `go vet`, `internal/archtest`) and README touch where operators/agents are affected.

**Observability:** structured `slog` for session connect/disconnect, action dispatch, apply/dispatch errors; `config.NewLogger` + `-log-level` / config `log_level`.
