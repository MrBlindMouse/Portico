# portico.Options

Passed to `Serve`, `ServeDev`, or `NewHost`. Many fields merge from [JSON config](config-file.md) via `Options.MergeConfig`.

| Field | Purpose |
|-------|---------|
| `DataDir` | Host key, identity KV, hashed accounts directory |
| `HostKeyPath` | Explicit host key file |
| `InspectAddr` | Inspect HTTP bind address; empty = off |
| `StatusAddr` | Status HTTP bind address; empty = off |
| `IdleTimeout` | `0` = 30m default; negative = disabled |
| `MaxConns`, `MaxConnsPerIP` | Connection caps |
| `AllowKeys`, `DenyKeys` | SSH fingerprint lists |
| `ActionRate`, `ActionBurst` | Per-session action limit; negative rate = off |
| `AssetDir`, `AssetPollEvery` | Filesystem assets + poll interval |
| `Logger`, `LogLevel` | `slog`; levels: debug, info, warn, error |
| `AccountBackend` | Primary `identity.AccountStore` |
| `HashedAccounts` | Bcrypt files under `DataDir` |
| `AccountStore` | Dev plaintext store |
| `IdentityStore` | Override identity KV backend |
| `Accounts` | `runtime.AccountConfig` wrapper |
| `AllowAccounts`, `DenyAccounts` | Account name lists |
| `RequireAccount` | Reject guests / unlinked keys |
| `ShutdownGrace` | Drain on shutdown |
| `KeepAliveInterval`, `KeepAliveCountMax` | SSH keepalive; `0` = off |
| `ActionBefore`, `ActionAfter` | Hooks around dispatch |
| `OnReady` | Called with `*Host` after wiring |
| `Theme` | Session theme; nil uses Default/Plain |

Account resolution order: `AccountBackend` → hashed files → `AccountStore` → `Accounts.Accounts`.

Listen address is **not** on `Options`; pass it to `Serve(addr, ...)`.

`MergeConfig` overlays file fields onto `Options` except `Listen` (use with `Serve(cfg.Listen, ...)` in your `main`).
