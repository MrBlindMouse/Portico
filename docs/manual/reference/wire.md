# Wire format

`installation.Wire` validates graph wiring without `Render`/`Run` functions. Used for inspect POST endpoints and `portico wire`.

## JSON shape

```json
{
  "root": "home",
  "scenes": {
    "home": {
      "inputs": { "greet": "greet" }
    }
  },
  "actions": ["greet"],
  "spaces": ["chat"],
  "assets": ["welcome"]
}
```

| Field | Meaning |
|-------|---------|
| `root` | Root scene id (required for reload preview) |
| `scenes` | Map of scene id → input bindings (action names as strings) |
| `actions` | Action ids that must exist |
| `spaces` | Space ids |
| `assets` | Optional asset names (reload preview warns if missing live) |

## Validate

```go
installation.ValidateWire(w)
```

Builds stub scenes/actions and runs `Build()`.

## Reload preview

`installation.PreviewReload(current, wire, liveSessions)` returns `ReloadPreview` with `ok`, `errors`, `stranded_sessions`, `warnings`.

Stranding occurs when a live session's current scene or joined space is absent from the proposed wire.
