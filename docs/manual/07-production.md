# Production and operations

This chapter matches the [README Production](../../README.md#production) section with operational detail.

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
- Persistent `.portico-data` on durable disk; backup identity KV and accounts.
- Rotate host key only with a deliberate migration plan.

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

Use `info` in production; `debug` only when diagnosing input issues.

## Security checklist

- [ ] Bcrypt accounts (not dev plaintext)
- [ ] Inspect/status loopback or off
- [ ] Connection and rate limits appropriate for exposure
- [ ] Key/account allow lists if the site is not fully public
- [ ] Unique SSH host key per environment
