# Hot reload

## API

```go
host, err := portico.NewHost(addr, exp, opts)
// or capture via Options.OnReady during Serve wiring

err := host.HotReload(builder) // *installation.Builder after Build()-able edits
```

`HotReload` swaps the validated `Installation`. Rejected when:

- `Build()` fails on the new graph
- Live sessions would be stranded (scene or joined space removed)

Same rules as `installation.PreviewReload`.

## Dry-run

While the server runs, export the live wiring:

```bash
curl -s http://127.0.0.1:9922/inspect/wire | jq . > live.wire.json
```

Propose a change (edit JSON or build from Go with `exp.Wire()`), then preview:

```bash
curl -s -X POST http://127.0.0.1:9922/inspect/reload-preview \
  -H 'Content-Type: application/json' \
  -d @proposed.wire.json
```

Check `ok` and `stranded_sessions` before calling `HotReload`. Applying wire JSON does not swap the installation — only `HotReload` with real `Render`/`Run` funcs does.

## Development loop

`Options.OnReady` can retain `*Host` and start `portico.Every` for timers. There is no built-in watch on plan `.go` files; restart or wire your own watcher.

## Assets

`Host` wiring loads `AssetDir` on start. Asset poll continues to call `Engine.UpdateAssets` on changes.
