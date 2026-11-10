# Running the server

## ServeDev vs Serve

| API | SSH address | Inspect / status |
|-----|-------------|------------------|
| `ServeDev(exp, opts)` | Fixed `:2222` | Defaults `127.0.0.1:9922` / `9923` when empty |
| `Serve(addr, exp, opts)` | Your `addr` string | Only if `InspectAddr` / `StatusAddr` set; empty disables |

```go
portico.Serve(":2222", exp, opts)
```

Blocks until SIGINT/SIGTERM (after optional shutdown grace).

## Site entry pattern

Generated sites parse `-addr` and call `Serve`. Add flags or load [JSON config](reference/config-file.md) as needed.

```go
addr := flag.String("addr", ":2222", "SSH listen address")
flag.Parse()
opts := portico.Options{
    DataDir:     ".portico-data",
    InspectAddr: "127.0.0.1:9922",
    StatusAddr:  "127.0.0.1:9923",
}
portico.Serve(*addr, exp, opts)
```

`PORTICO_LOG_LEVEL` and `PORTICO_DATA` are honored when your `main` reads them (init template reads `PORTICO_DATA`).

## NewHost and hot reload

```go
host, err := portico.NewHost(addr, exp, opts)
// ...
host.HotReload(builder) // rejects invalid graphs or stranded sessions
```

Use `Options.OnReady` to capture `*portico.Host` from `Serve` wiring. Details: [Hot reload](appendices/hot-reload.md).

## Demo CLI (not your site entry)

```bash
portico serve -demo -addr :2222 -data .portico-data -config portico.json
```

Bundled guestbook, chat, and landing smoke test only. Real sites use `go run .` in their module.

## Auxiliary HTTP

| Server | Option | Default in ServeDev / init |
|--------|--------|------------------------------|
| Inspect | `InspectAddr` | `127.0.0.1:9922` |
| Status | `StatusAddr` | `127.0.0.1:9923` |

Bind to loopback in production unless you authenticate in front.

## Assets

`Options.AssetDir` + `AssetPollEvery` load and poll files; updates call `Engine.UpdateAssets`. Blog plan expects assets like `post-<id>`.

## Logging

`LogLevel` or `Logger` — see [Options](reference/options.md). Structured `slog` for connect, dispatch, apply errors.

## Shutdown

`ShutdownGrace` drains sessions on SIGTERM before exit (when set).

Full option list: [reference/options.md](reference/options.md).
