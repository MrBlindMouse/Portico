# Wire format

`installation.Wire` is a JSON sketch of a scene graph: root, input bindings, action ids, space ids, and text asset names. It does not carry `Render`, `Run`, space hooks, or session hooks.

Use a wire when an agent is checking a proposed graph, or when `portico wire codegen` should start a file. A site that already has Go calls `Host.HotReload` with a builder. It does not need a wire.

## From a built installation

```go
w := exp.Wire()
body, err := json.Marshal(w)
```

`(*Installation).Wire` on nil returns the zero value. `ValidateWire` builds stub scenes and actions and runs `Build()`.

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

`ok` is false when `stranded_sessions` is non-empty. That is the same rule as `Engine.SetInstallation`: a live session whose scene or joined space is missing from the proposal is stranded.

```bash
curl -s -X POST http://127.0.0.1:9922/inspect/reload-preview \
  -H 'Content-Type: application/json' \
  -d @proposed.wire.json
```

Posting that body does not swap the live installation. `portico wire validate` checks a file. `portico wire codegen` writes a `main` package whose `Render` and `Run` are empty. Applying the change is still `Host.HotReload` with a builder that contains the real functions.

Asset names on the wire warn when the live installation has not loaded them. They do not upload asset bytes.
