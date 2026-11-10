# CLI reference (`portico`)

Install: `go install github.com/MrBlindMouse/Portico/cmd/portico@latest`

**Your site's entry point is `go run .` in the site module**, not `portico serve`.

## Commands

| Command | Purpose |
|---------|---------|
| `init` | Create site files (`main.go`, `AGENTS.md`, `README.md`) |
| `generate` | Write one plan into the site as ordinary Go |
| `docs` | Print doc URLs for a git tag |
| `wire validate` | Validate `installation.Wire` JSON file |
| `wire codegen` | Emit starter Go from wire JSON |
| `inspect` | HTTP client for inspect API |
| `identity-export` | Export identity KV (see `main.go` / identity CLI) |
| `account-export` | Export bcrypt accounts from `DataDir` |
| `account-import` | Import account JSON into `DataDir` |
| `serve -demo` | Bundled smoke test server only |

## init

```
portico init [-dir .] [-module path] [-force]
```

Runs `go mod init` when needed, `go get` Portico, writes `main.go`, `AGENTS.md`, and `README.md`. `main.go` is one scene and a `greet` action. Stderr hints at `generate` and lists plan names.

## generate

```
portico generate [-dir .] [-pkg name] [-force] <name>
portico generate -list
```

Names: `landing`, `guestbook`, `chat`, `register`, `form`, `blog`, `lobby`. Each plan builds on its own. Writes that plan into `-dir` and prints a local `Merge` line. `guestbook` and `chat` also write `rest.go`. If that file is already there, generating the other one needs `-force`. See [Plans](plans.md).

## docs

```
portico docs [-tag main]
```

Prints links to `docs/manual/README.md`, `docs/building-a-site.md`, and `docs/site-agents.md` at the chosen tag.

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
