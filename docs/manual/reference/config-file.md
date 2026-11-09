# JSON config file

Type: `config.File` in package `github.com/MrBlindMouse/Portico/config`.

Used by `portico serve -demo -config path.json`. Site authors can load the same struct in `main`:

```go
cfg, err := config.Load("portico.json")
opts := portico.Options{}.MergeConfig(cfg)
portico.Serve(cfg.Listen, exp, opts) // or your flag for addr
```

## Fields

| JSON key | Go field | Notes |
|----------|----------|-------|
| `listen` | `Listen` | SSH address (not stored on Options) |
| `data_dir` | `DataDir` | |
| `log_level` | `LogLevel` | |
| `inspect_addr` | `InspectAddr` | |
| `status_addr` | `StatusAddr` | |
| `idle_timeout` | `IdleTimeout` | Go `time.Duration` JSON (nanoseconds as number) |
| `max_conns` | `MaxConns` | |
| `max_conns_per_ip` | `MaxConnsPerIP` | |
| `allow_keys`, `deny_keys` | string arrays | |
| `action_rate`, `action_burst` | floats | |
| `asset_dir` | `AssetDir` | |
| `asset_poll_interval` | `AssetPollEvery` | duration |
| `accounts_enabled` | `Accounts` | demo CLI enables dev store |
| `hashed_accounts` | `HashedAccounts` | |
| `allow_accounts`, `deny_accounts` | string arrays | |
| `require_account` | `RequireAccount` | |
| `shutdown_grace` | `ShutdownGrace` | duration |
| `keepalive_interval` | `KeepAliveInterval` | duration |
| `keepalive_count_max` | `KeepAliveCountMax` | int |

`config.Load` returns zero value and no error when the file is missing.

## Example

```json
{
  "listen": "127.0.0.1:2222",
  "data_dir": ".portico-data",
  "log_level": "info",
  "inspect_addr": "127.0.0.1:9922",
  "status_addr": "127.0.0.1:9923",
  "hashed_accounts": true,
  "max_conns": 128,
  "idle_timeout": 1800000000000
}
```

(`idle_timeout` example = 30 minutes in nanoseconds.)
