# Hot reload

## API

```go
host, err := portico.NewHost(addr, exp, opts)
// or capture via Options.OnReady during Serve wiring

err := host.HotReload(builder) // *experience.Builder after Build()-able edits
```

`HotReload` swaps the validated `Experience`. Rejected when:

- `Build()` fails on the new graph
- Live sessions would be stranded (scene or joined space removed)

Same rules as `experience.PreviewReload`.

## Dry-run

While the server runs:

```bash
curl -s -X POST http://127.0.0.1:9922/inspect/reload-preview \
  -H 'Content-Type: application/json' \
  -d @proposed.wire.json
```

Check `ok` and `stranded_sessions` before calling `HotReload`.

## Assets

`Host` wiring loads `AssetDir` on start. Asset poll continues to call `Engine.UpdateAssets` on changes.
