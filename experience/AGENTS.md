# experience/

Scenes, actions, spaces, and builder validation. Site modules **import** this package (`github.com/MrBlindMouse/Portico/experience`) and define their experience in their own repo. Do not add a site's domain scenes to this package.

## Invariants

- Actions return `[]Effect` only; they do not mutate session or space state.
- `Build()` must reject undefined action refs, scene refs, and invalid space effect refs.
- `RenderContext` / `SceneContext` / `ActionContext` are implemented only in `runtime`.
- `Scene.Render` must use `RenderContext` (no effect application during render).

## Contexts

| Type | Used in | May |
|------|---------|-----|
| `RenderContext` | `Scene.Render` | Read identity, local, assets, space state, `Size()`, `Theme()` |
| `SceneContext` | (via actions) | Above + `SetLocal`/`Transition` **effects** |
| `ActionContext` | `Action.Run` | Above + `InputLine()`, `Space()`, `Persist()` |

## Effects (runtime applies)

`SetLocal`, `Transition`, `Persist`, `JoinSpace`, `LeaveSpace`, `SpaceSet`, `SpaceCAS`, `AppendSpaceList`, `Broadcast`, `ReserveNickname`, `RegisterAccount` (requires runtime account store).

## Cross-session events

DESIGN “emit events” = write domain data with `SpaceSetEffect` / `AppendSpaceListEffect`, then `BroadcastEffect`. No in-process event bus.

## Prefer / avoid

- Prefer builder methods, typed IDs, and `Merge` for scaffolds.
- Prefer `InputLine()` for multi-word commands when using prefix dispatch.
- `Scene.Keys` dispatches one bound key per action (one rune, or `up`/`down`/`left`/`right`). Leave it unset for line entry.
- List selection is local state: `MoveList` returns a `SetLocal` effect, `RenderList` only reads. Default key `LocalListIndex`.
- `PlaceInput` describes the caret for this frame. It does not write session state. One rectangle. `Keys` plus a placed input fails at render.
- Avoid importing `runtime`, `scaffold`, or `examples` from reusable site packages (your `main` may import scaffolds).
