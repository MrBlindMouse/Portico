# Getting started

## Prerequisites

- Go toolchain (see `go.mod` in the Portico module for the minimum version).
- An SSH client (`ssh`).
- For local dev, nothing else; for production SSH on port 22, see [Production](07-production.md).

## Create a site module

**Generate:**

```bash
mkdir my-site && cd my-site
go run github.com/MrBlindMouse/Portico/cmd/portico@latest init -module github.com/you/my-site
go run .
```

| Flag | Meaning |
|------|---------|
| `-force` | Overwrite `main.go`, `AGENTS.md`, `README.md` |
| `-module` | Required when `go.mod` is missing |

After `init`, add plans from the site directory:

```bash
go run github.com/MrBlindMouse/Portico/cmd/portico generate <name>
go run github.com/MrBlindMouse/Portico/cmd/portico generate -list
```

Agents: start with **AGENTS.md** in the site repo.

**Or write manually:** `go mod init`, `go get github.com/MrBlindMouse/Portico`, then build an installation in `main.go`. `portico init` writes that same program.

## First SSH session

Default listen address from `portico init` is **`:2222`** (`-addr` flag).

```bash
go run .
# another terminal
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

## Data directory

`Options.DataDir` (default `.portico-data` in templates) stores:

- SSH host key
- Per-identity persisted KV (`identity` file store)

Set `PORTICO_DATA` in templates that support it, or pass `DataDir` in code. Add `.portico-data/` to `.gitignore`.

## Inspect while developing

When `InspectAddr` is set (init templates use `127.0.0.1:9922`):

```bash
curl -s http://127.0.0.1:9922/inspect/snapshot | jq .
```

See [Inspect HTTP API](reference/inspect-http.md).

## Environment variables (init templates)

| Variable | Purpose |
|----------|---------|
| `PORTICO_DATA` | Override data directory. The init `main` reads this. |

## Next steps

1. Read [Authoring](02-authoring.md) for scenes, actions, and effects.
2. Skim [reference/effects.md](reference/effects.md).
3. Before production, read [Production](07-production.md).
