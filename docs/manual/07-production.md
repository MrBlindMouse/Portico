# Production and operations

Production guidance for sites that import Portico. Pair with [reference/options.md](reference/options.md).

## Listen addresses

Pass any Go TCP `host:port` to `Serve`:

| Address | Meaning |
|---------|---------|
| `:2222` | All interfaces, high port (dev) |
| `203.0.113.10:22` | Specific IP, port 22 |
| `127.0.0.1:2222` | Loopback only |

Wire via `-addr` in `main`, env, or supervisor args.

## Two SSH servers on port 22

One process per **IP:port**. Typical layout when OpenSSH already owns admin SSH:

1. Add a **second public IP** on the host.
2. DNS: `play.example.com` → that IP (primary hostname stays on admin SSH).
3. Configure **sshd** `ListenAddress` on the **primary IP only** (not `0.0.0.0` if both use :22).
4. Run Portico: `Serve("203.0.113.10:22", exp, opts)`.

Users: `ssh play.example.com` (Portico) vs existing admin host unchanged.

## Privileged ports (Linux)

Ports &lt; 1024 need capability or root. When Portico alone binds :22:

```bash
go build -o my-site
sudo setcap 'cap_net_bind_service=+ep' ./my-site
./my-site -addr :22
```

Prefer second-IP layout over `setcap` when coexisting with sshd on :22.

## Accounts and data

- `HashedAccounts` + `DataDir` or `AccountBackend` file store.
- Persistent `.portico-data` on durable disk; see [backup and restore](#backup-and-restore) below.

## Backup and restore

Back up the whole data directory (or at minimum):

| Path | Contents |
|------|----------|
| `identities/`, `nicknames.json` | Identity KV (see [file layout](05-identity-accounts.md#file-layout)) |
| `accounts/` | Bcrypt account records when using hashed accounts |
| `store-version` | Layout marker (absent = version 1) |
| Host key file | `Options.HostKeyPath` or default under data dir |

Export from a running layout:

```bash
portico identity-export -data .portico-data > identities.json
portico account-export -data .portico-data > accounts.json
```

Restore: stop the site, restore files into a fresh `DataDir`, or use `account-import` for accounts. A newer `store-version` than the binary supports is refused — use the matching framework version or treat restore as a clean break (wipe and re-import JSON).

## SSH host key rotation

The host key is **not** inside the identity store layout. `Options.HostKeyPath` points at the key file (default under `DataDir`).

Rotation is a **clean break**: generate a new key path, update `Options`, restart, and have clients accept the new fingerprint (`known_hosts` / `ssh-keygen -R`). Use distinct keys per environment (dev vs production).

## Connection instructions

When `StatusAddr` is set, `GET /` on the status mux prints a short summary and an **ssh** hint derived from the listen address (concrete bind address when not `0.0.0.0`). For public sites, also document DNS and port in the site README. Status remains unauthenticated — bind to loopback in production or disable it.

## Inspect and status

- Default init templates bind inspect/status to **localhost**.
- Set `InspectAddr` / `StatusAddr` to `""` to disable.
- Never expose these endpoints on untrusted networks without auth.

## Limits and abuse

Set as needed:

- `MaxConns`, `MaxConnsPerIP`
- `IdleTimeout` (0 = 30m default; negative = off)
- `ActionRate` / `ActionBurst`
- `AllowKeys`, `DenyKeys`, `AllowAccounts`, `DenyAccounts`, `RequireAccount`
- `KeepAliveInterval`, `KeepAliveCountMax`

## Supervisors

On SIGINT/SIGTERM, `Serve` cancels context and honors `ShutdownGrace`. Point health checks at [status HTTP](reference/status-http.md) `/ready` when `StatusAddr` is enabled.

## Logging

Use `info` in production. `debug` is for diagnosing input. At debug, Portico logs the first word of each line and the line length. The rest of the line, including a password on `register user pass`, is not logged.

## Security checklist

- [ ] Bcrypt accounts (not dev plaintext)
- [ ] Inspect/status loopback or off
- [ ] Connection and rate limits appropriate for exposure
- [ ] Key/account allow lists if the site is not fully public
- [ ] Unique SSH host key per environment
