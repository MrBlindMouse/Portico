# Status HTTP API

Served when `Options.StatusAddr` is set (dev default `127.0.0.1:9923`).

## Routes

| Path | Response |
|------|----------|
| `/health` | `200` body `ok\n` |
| `/ready` | `200` `ready\n` if engine loaded and not draining; else `503` `not ready\n` |
| `/metrics` | JSON: `actions_total`, `sessions_active`, `reload_failures` |
| `/` | Plain-text summary: SSH addr, action count, sessions, reload failures, ready flag, connect hint |

## Use with supervisors

- **Liveness:** `GET /health`
- **Readiness:** `GET /ready` (respects drain during shutdown)

Do not expose on untrusted networks.
