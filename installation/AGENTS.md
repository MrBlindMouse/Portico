# installation/

Builder, `Installation`, wire, preview, and form/list helpers. Scene, action, space, effects, and contexts live in `components`. This package imports `components` and does not re-export those types.

Site modules import both. Do not add a site's domain scenes here.

## Invariants

- `Build()` must reject undefined action refs, scene refs, and invalid space effect refs.
- This package must not import `runtime`.

## Contexts

| Type | Used in | May |
|------|---------|-----|
| `RenderContext` | `Scene.Render` | Read identity, local, assets, space state, `Size()`, `Theme()`, `RecentCommands()` |
| `SceneContext` | (via actions) | Above + `SetLocal`/`Transition` **effects** |
| `ActionContext` | `Action.Run` | Above + `InputLine()`, `Space()`, `Persist()` |

## Effects (runtime applies)

`SetLocal`, `Transition`, `Persist`, `JoinSpace`, `LeaveSpace`, `SpaceSet`, `SpaceCAS`, `AppendSpaceList`, `Broadcast`, `ReserveNickname`, `RegisterAccount` (requires runtime account store).

## Cross-session events

DESIGN “emit events” = write domain data with `SpaceSetEffect` / `AppendSpaceListEffect`, then `BroadcastEffect`. No in-process event bus.

## Prefer / avoid

- Prefer builder methods, typed IDs, and `Merge` of builders the site owns (`go run …/cmd/portico generate`).
- Prefer `InputLine()` for multi-word commands when using prefix dispatch.
- `Scene.Keys` dispatches one bound key per action (one rune, or `up`/`down`/`left`/`right`). Leave it unset for line entry.
- List selection is local state: `MoveList` returns a `SetLocal` effect, `RenderList` only reads. Default key `LocalListIndex`.
- `PlaceInput` describes the caret for this frame. It does not write session state. One rectangle. `Keys` plus a placed input fails at render.
- Avoid importing `runtime`, `internal`, or `examples` from site code.
