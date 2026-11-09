# Authoring an experience

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
    Scene("home", experience.Scene{...}).
    Action("greet", experience.Action{...}).
    Space("lobby", experience.Space{...}). // optional
    Asset("welcome", "Hello\n").     // optional inline assets
    BindInput("home", "help", "show-help"). // optional sugar
    OnSessionStart(experience.SetLocalEffect("k", "v")...).
    OnSessionEnd(...).
    Merge(scaffold.Guestbook()).     // optional; later Merge wins on id collision
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
experience.Scene{
    Render: func(ctx experience.RenderContext) string { return "..." },
    Inputs: map[string]experience.ActionID{"greet": "greet"},
    Keys:   false, // true = key dispatch, no line prompt
    ValidateLocal: func(key string, value any) error { return nil },
}
```

- **Render** is read-only via `RenderContext`.
- **Inputs** map command keys to actions (see [Input](03-input.md)).
- **Keys** switches to immediate key dispatch for list UIs.

## Actions

```go
experience.Action{
    Run: func(ctx experience.ActionContext) []experience.Effect {
        return []experience.Effect{
            experience.SetLocalEffect("flash", "ok"),
        }
    },
}
```

Use `ctx.InputLine()` for the full submitted line. Use `experience.ParseFields(line, n)` to split tokens.

## Spaces

```go
experience.Space{
    OnJoin:  []experience.Effect{...},
    OnLeave: []experience.Effect{...},
    ValidateSet: func(key string, old, newValue any) error { return nil },
}
```

Cross-session updates: write space state with space effects, then `experience.BroadcastEffect(spaceID)` so members re-render. There is no in-process event bus.

## Merging scaffolds

```go
portico.New().
    Merge(scaffold.Guestbook()).
    Merge(scaffold.Landing(map[string]experience.SceneID{"guestbook": "guestbook"})).
    Build()
```

Later `Merge` calls override scenes, actions, and spaces with the same ids. Session start/end hook slices append.

Copy a scaffold into your repo to customize it ([examples/notes](../../examples/notes)); do not edit files inside the module cache.

## Wire and validation

`experience.Wire` is a JSON-serializable sketch (ids and input bindings, no funcs). Use it for:

- `POST /inspect/validate` and `POST /inspect/reload-preview`
- `portico wire validate -file site.wire.json`

See [Wire format](reference/wire.md).

## Hard rules

1. Actions return effects only.
2. Scenes only render.
3. Do not import `runtime`, `inspect`, or `internal/...` from site code.
4. Use typed ids: `experience.SceneID`, `ActionID`, `SpaceID`.

## Anti-patterns

- Mutating maps or session state inside `Render` or `Run` without effects.
- Importing `runtime` to reach sessions or spaces.
- Assuming DESIGN.md “Planned” items exist.
