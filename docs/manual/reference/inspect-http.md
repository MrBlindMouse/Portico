# Inspect HTTP API

Served when `Options.InspectAddr` is set (default in dev: `127.0.0.1:9922`). Handler: `inspect.NewHandler(engine)`.

**Security:** no authentication. The handler binds whatever `InspectAddr` you set. `ServeDev` and `portico init` default to `127.0.0.1`. A public address exposes live sessions and identity KV. Leave it unset, or put auth in front, outside trusted networks.

## Routes

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/inspect/snapshot` | Installation summary + sessions + spaces |
| GET | `/inspect/sessions` | Session array only |
| GET | `/inspect/session/{id}` | One session or 404 |
| GET | `/inspect/spaces` | Space summaries |
| GET | `/inspect/identity/{fingerprint}` | Identity KV snapshot |
| POST | `/inspect/validate` | Validate `installation.Wire` JSON body |
| POST | `/inspect/reload-preview` | Hot-reload dry-run vs live sessions |

POST bodies are limited to 1 MiB.

## Snapshot shape (summary)

```json
{
  "installation": {
    "root": "home",
    "hash": "...",
    "scenes": ["home", "..."],
    "action_count": 3,
    "space_count": 1,
    "builder_note": "..."
  },
  "sessions": [ { "id", "scene", "identity", "fingerprint", "inputs", ... } ],
  "spaces": { "space-id": { "member_count", "keys" } }
}
```

Session entries include `render_preview`, `local_keys`, `spaces`, `available_actions`, etc. See `inspect` package types: `Snapshot`, `SessionSnap`.

## Validate response

POST body: `installation.Wire`. Response JSON includes `ok` and `error` on failure (handler-specific).

## Reload preview response

`installation.ReloadPreview`:

```json
{
  "ok": true,
  "errors": [],
  "stranded_sessions": [],
  "warnings": []
}
```

`ok` is false when validation fails or live sessions would be stranded in removed scenes/spaces.

## CLI

```bash
portico inspect -url http://127.0.0.1:9922 -resource snapshot
portico inspect -url http://127.0.0.1:9922 -resource reload-preview -method POST -file wire.json
```

Resources: `snapshot`, `sessions`, `reload-preview`, `validate`.
