# AGENTS.md

Ops contract for agents working on **Portico the framework**. Site-specific apps live in their own `main` packages (see `examples/hello`). **DESIGN.md is the vision contract; "Planned" there is not implemented until shipped.**

## Audience

| Who | Owns |
|-----|------|
| Framework (`experience`, `runtime`, `portico`, …) | Layers, invariants, extension points |
| Site author (`examples/*`, user repos) | Scenes, actions, domain, `main` + `Serve` |
| `cmd/portico` | Operator helpers; **not** the primary site entry (use `-demo` only for bundled smoke test) |

## Run and test

```bash
go test ./...
go vet ./...
go run ./examples/hello   # reference site, SSH on :2222
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

CI runs `go test`, `go vet`, and `internal/archtest` on every push.

## Principles (ceilings and non-goals)

- Session and space mutation happens only in `runtime/apply` — not in actions, scenes, or site code.
- `runtime.Session` fields are unexported; construct sessions only inside `runtime`.
- `experience` must not import `runtime`.
- Thousands of concurrent sessions, real-time guarantees, and full production account security are DESIGN goals; dev account auth is explicit and plaintext.
- No domain logic in `runtime` — only orchestration.

## Architecture map

| Package       | Role |
|---------------|------|
| `portico`     | Public `New()`, `Serve()`, `ServeDev()`, `NewHost()`, `Host.HotReload`, `Options` |
| `experience`  | Scenes, actions, spaces, builder, effects (descriptions) |
| `runtime`     | SSH server, sessions, dispatch, apply, spaces registry |
| `identity`    | SSH fingerprint / account identity, file-backed KV, nicknames |
| `render`      | Terminal layout helpers |
| `inspect`     | Read-only JSON snapshots for agents |
| `scaffold`    | Composable starter definitions (optional for sites) |
| `asset`       | Named filesystem assets and poll watcher |
| `cmd/portico` | CLI: `inspect`, `scaffold`, `identity-export`, `serve -demo` |
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

## Shipped vs planned (v0.1 snapshot)

**Shipped in tree:** experience builder, `RenderContext`/`ActionContext`, effects (incl. nickname, register account), apply, embedded SSH, identity file store, per-IP/global conn limits, idle reset on input, spaces broadcast/CAS, inspect HTTP, scaffolds (landing/guestbook/chat/registration/nickname), assets + poll, hot-reload with live-session check, status/health HTTP, configurable rate limit, dev account SSH + registration scaffold.

**Still lighter than DESIGN §7:** full JSON definition format (wire validate only), form/input helpers, event bus, email verification, hashed account passwords in production, clustered spaces, inspect HTTP reload without Builder.

**Observability:** structured `slog` for session connect/disconnect, action dispatch, apply/dispatch errors; `config.NewLogger` + `-log-level` / config `log_level`.
