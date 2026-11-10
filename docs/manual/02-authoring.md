# Authoring an installation

## Mental model

```
SSH connect → session in root scene → user input
  → dispatcher matches scene Inputs (or Keys) → action Run()
  → []Effect → runtime apply (only mutator) → re-render scene
```

Your site code **never** mutates session or space state. Actions return **effects**; the runtime applies them.

## Builder API

```go
portico.New().
    Root("home").                    // required root scene id
    Scene("home", components.Scene{...}).
    Action("greet", components.Action{...}).
    Space("lobby", components.Space{...}). // optional
    Asset("welcome", "Hello\n").     // optional inline assets
    BindInput("home", "help", "show-help"). // optional sugar
    OnSessionStart(components.SetLocalEffect("k", "v")...).
    OnSessionEnd(...).
    Merge(Guestbook()).              // after go run …/cmd/portico generate guestbook
    Build()
```

### Build validation

`Build()` fails fast on:

- Missing root scene
- Scene `Inputs` referencing unknown actions
- Effects in space hooks referencing unknown scenes/spaces
- `Keys: true` scenes with illegal input names (must be one rune or `up`/`down`/`left`/`right`)

Fix builder errors before debugging SSH behavior.

## Scenes

```go
components.Scene{
    Render: func(ctx components.RenderContext) string { return "..." },
    Inputs: map[string]components.ActionID{"greet": "greet"},
    Keys:   false, // true = key dispatch, no line prompt
    ValidateLocal: func(key string, value any) error { return nil },
}
```

- **Render** is read-only via `RenderContext`.
- **Inputs** map command keys to actions (see [Input](03-input.md)).
- **Keys** switches to immediate key dispatch for list UIs.

## Actions

```go
components.Action{
    Run: func(ctx components.ActionContext) []components.Effect {
        return []components.Effect{
            components.SetLocalEffect("flash", "ok"),
        }
    },
}
```

Use `ctx.InputLine()` for the full submitted line. Use `components.ParseFields(line, n)` to split tokens.

## Spaces

```go
components.Space{
    OnJoin:  []components.Effect{...},
    OnLeave: []components.Effect{...},
    ValidateSet: func(key string, old, newValue any) error { return nil },
}
```

Cross-session updates: write space state with space effects, then `components.BroadcastEffect(spaceID)` so members re-render. There is no in-process event bus.

## Merging generated plans

```bash
go run github.com/MrBlindMouse/Portico/cmd/portico generate guestbook
go run github.com/MrBlindMouse/Portico/cmd/portico generate landing
```

```go
portico.New().
    Merge(Guestbook()).
    Merge(Landing(map[string]components.SceneID{"guestbook": "guestbook"})).
    Build()
```

Later `Merge` calls override scenes, actions, and spaces with the same ids. Session start/end hook slices append. Edit the generated files. Do not patch the module cache.

## Wire and validation

`installation.Wire` is a JSON-serializable sketch (ids and input bindings, no funcs). Use it for:

- `POST /inspect/validate` and `POST /inspect/reload-preview`
- `go run github.com/MrBlindMouse/Portico/cmd/portico wire validate -file site.wire.json`

See [Wire format](reference/wire.md).

## Hard rules

1. Actions return effects only.
2. Scenes only render.
3. Do not import `runtime` or `inspect` from site code.
4. Use typed ids: `components.SceneID`, `ActionID`, `SpaceID`.

## Anti-patterns

- Mutating maps or session state inside `Render` or `Run` without effects.
- Importing `runtime` to reach sessions or spaces.
- Assuming DESIGN.md “Planned” items exist.
