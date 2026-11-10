# Portico SSH site

This repository is an **SSH terminal experience**: visitors connect with `ssh`, see text drawn by your scenes, and type commands that run your actions. Under the hood it is a normal Go program that imports [Portico](https://github.com/MrBlindMouse/Portico).

Out of the box you get a single **home** scene and a **greet** command. Everything else—guestbook, chat, menus, accounts—is yours to add in Go, often starting from a generated **plan**.

## Run locally

```bash
go run .
```

In another terminal:

```bash
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

Type `greet` at the prompt. Identity data and the SSH host key are stored under `.portico-data/` (override with `PORTICO_DATA`).

**Listen address:** default `:2222` (`-addr` flag in `main.go`).

**Debug logs:** `PORTICO_LOG_LEVEL=debug` when your `main` reads it.

While the server runs, inspect endpoints on loopback (from init’s defaults) include:

```bash
curl -s http://127.0.0.1:9922/inspect/snapshot
```

Inspect has no login. Do not expose it on the public internet.

## Grow the site

1. **Plans** — convenient starters you own and edit (this repo already depends on Portico in `go.mod`):

   ```bash
   portico generate guestbook
   portico generate -list
   ```

   If the CLI is not on your PATH: `go install github.com/MrBlindMouse/Portico/cmd/portico@<same-version-as-go.mod>` once, or `go run github.com/MrBlindMouse/Portico/cmd/portico generate guestbook` from this directory (uses the module version in `go.mod`, not `@latest`).

   Wire the printed `Merge(…)` line into `main.go`, then change the generated file.

2. **By hand** — add scenes and actions with `portico.New()` in `main.go` or a `site/` package. Actions return effects; scenes only render.

3. **Docs** — full reference for your Portico version:

   `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/manual/README.md`

   Replace `<tag>` with the module version in `go.mod`. Short guide: [building-a-site](https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/building-a-site.md).

## Project layout

| Path | Purpose |
|------|---------|
| `main.go` | Build the installation, set `portico.Options`, call `Serve` |
| `*.go` from `generate` | Plan code you edit (e.g. `guestbook.go`) |
| `.portico-data/` | Persistence + host key (add to `.gitignore`) |
| [AGENTS.md](AGENTS.md) | Contract for AI agents working in this repo |

## Production

Local `go run .` uses **`Serve`** with `-addr` default **`:2222`** and loopback inspect/status (see `main.go`). For a real deployment, build a binary and extend **`portico.Options`** in `main`—do not use **`ServeDev`** unless you deliberately want dev defaults.

| Concern | What to do |
|---------|------------|
| Listen | `go run . -addr 203.0.113.10:22` or `./my-site -addr :22` — any Go `host:port` |
| Data | Set `Options.DataDir` (or `PORTICO_DATA`); back up that directory |
| Accounts | `HashedAccounts: true` with `DataDir`, or `AccountBackend` with `identity.FileAccountStore` — not dev plaintext `AccountStore` |
| Inspect / status | Empty `InspectAddr` / `StatusAddr` disables HTTP; keep loopback or off on the public internet (no authentication) |
| Limits | `MaxConns`, `MaxConnsPerIP`, `AllowKeys` / `DenyKeys`, `RequireAccount`, `LogLevel` |
| Shutdown | `ShutdownGrace` when a supervisor sends SIGTERM |

Example shape (add fields in `main.go` as needed):

```go
opts := portico.Options{
    DataDir:        dataDir,
    HashedAccounts: true,
    MaxConns:       256,
    InspectAddr:    "", // off in production unless protected
    StatusAddr:     "",
}
portico.Serve(*addr, inst, opts)
```

### Two SSH servers on port 22

You cannot bind two daemons to the same **IP:port**. If **OpenSSH** already serves admin SSH on port 22:

1. Add a **second public IP** on the machine (or VM).
2. Point a **hostname** at it (`play.example.com` → that address).
3. Restrict **sshd** with `ListenAddress` on the primary IP (not `0.0.0.0` on :22 if Portico also uses :22).
4. Run this site on the second IP: `./my-site -addr 203.0.113.10:22`.

Visitors use `ssh play.example.com` for the experience; admin SSH stays on the original host.

### Privileged port 22 on Linux

Ports below 1024 need extra permission. Prefer the **second-IP** layout when sshd and Portico both use :22. On a dedicated listener only:

```bash
go build -o my-site
sudo setcap 'cap_net_bind_service=+ep' ./my-site
./my-site -addr :22
```

Running as root works but is discouraged.

More detail: [Production](https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/manual/07-production.md) and [Options](https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/manual/reference/options.md) (replace `<tag>` in `go.mod`).

## Agents

If you use coding agents, point them at **[AGENTS.md](AGENTS.md)** first—plans, rules, and version-pinned links to upstream docs.
