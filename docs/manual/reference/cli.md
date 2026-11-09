# CLI reference (`portico`)

Install: `go install github.com/MrBlindMouse/Portico/cmd/portico@latest`

**Your site's entry point is `go run .` in the site module**, not `portico serve`.

## Commands

| Command | Purpose |
|---------|---------|
| `init` | Create site files (`main.go`, `AGENTS.md`, `README.md`) |
| `docs` | Print doc URLs for a git tag |
| `scaffold` | Print `Merge(scaffold.*)` one-liners |
| `wire validate` | Validate `experience.Wire` JSON file |
| `wire codegen` | Emit starter Go from wire JSON |
| `inspect` | HTTP client for inspect API |
| `identity-export` | Export identity KV (see `main.go` / identity CLI) |
| `account-export` | Export bcrypt accounts from `DataDir` |
| `account-import` | Import account JSON into `DataDir` |
| `serve -demo` | Bundled smoke test server only |

## init

```
portico init [-dir .] [-style community|scratch] [-module path] [-force]
```

Runs `go mod init` when needed, `go get` Portico, writes templates.

## docs

```
portico docs [-tag main]
```

Prints links to `docs/building-a-site.md`, `docs/site-agents.md`, `docs/manual/`, and examples.

## scaffold

```
portico scaffold <landing|guestbook|chat|register|nickname|form|blog|lobby|recipe>
portico scaffold recipe <community|accounts>
```

Prints merge snippets to stdout.

## wire

```
portico wire validate -file site.wire.json
portico wire codegen -file site.wire.json [-out main.go]
```

Codegen writes a skeletal `main` package to stdout or `-out`.

## inspect

```
portico inspect [-url http://127.0.0.1:9922] [-resource snapshot] [-method GET] [-file body.json]
```

Default resource `snapshot`. POST for `validate` and `reload-preview`.

## account-export / account-import

```
portico account-export -data .portico-data
portico account-import -data .portico-data -file accounts.json
```

Requires file-backed account store under `DataDir`.

## serve (demo only)

```
portico serve -demo [-addr :2222] [-data .portico-data] [-inspect addr] [-status addr] [-config file.json] [-accounts] [-log-level info]
```

Exits with error without `-demo`.
