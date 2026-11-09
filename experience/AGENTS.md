# experience/

Declarative **site definitions**: scenes, actions, spaces, builder validation. Site authors spend most of their time here.

## Invariants

- Actions return `[]Effect` only; they do not mutate session or space state.
- `Build()` must reject undefined action refs, scene refs, and invalid space effect refs.
- `RenderContext` / `SceneContext` / `ActionContext` are implemented only in `runtime`.
- `Scene.Render` must use `RenderContext` (no effect application during render).

## Contexts

| Type | Used in | May |
|------|---------|-----|
| `RenderContext` | `Scene.Render` | Read identity, local, assets, space state |
| `SceneContext` | (via actions) | Above + `SetLocal`/`Transition` **effects** |
| `ActionContext` | `Action.Run` | Above + `InputLine()`, `Space()`, `Persist()` |

## Effects (runtime applies)

`SetLocal`, `Transition`, `Persist`, `JoinSpace`, `LeaveSpace`, `SpaceSet`, `SpaceCAS`, `AppendSpaceList`, `Broadcast`, `ReserveNickname`, `RegisterAccount` (requires runtime account store).

## Cross-session events

DESIGN “emit events” = write domain data with `SpaceSetEffect` / `AppendSpaceListEffect`, then `BroadcastEffect`. No in-process event bus.

## Prefer / avoid

- Prefer builder methods, typed IDs, and `Merge` for scaffolds.
- Prefer `InputLine()` for multi-word commands when using prefix dispatch.
- Avoid importing `runtime`, `scaffold`, or `examples` from reusable site packages (your `main` may import scaffolds).
