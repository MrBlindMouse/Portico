# Contexts reference

Runtime implements these interfaces; site code only uses them.

## RenderContext (scenes)

Used in `Scene.Render` — **read only** (except `PlaceInput`, which records caret metadata for the frame only).

| Method | Purpose |
|--------|---------|
| `Identity()` | Current identity |
| `Local(key)` | Session-local value |
| `Persisted(key)` | Identity KV |
| `Asset(name)`, `Assets()` | Experience assets |
| `Members(space)`, `SpaceLocal`, `SpaceList` | Space read API |
| `Size()` | Terminal columns, rows |
| `Theme()` | `render.Theme` |
| `PlaceInput(in)` | Caret placement (one per frame) |

## ActionContext (actions)

Embeds read APIs plus effect helpers and input.

| Method | Purpose |
|--------|---------|
| `InputLine()` | Full submitted line |
| `Space()` | `SpaceContext` for membership checks |
| `SetLocal`, `Transition`, `Persist` | Return effect values (not applied in Run) |

## SessionView (hooks)

`portico.Options.ActionBefore` / `ActionAfter` receive `experience.SessionView` for observability without importing `runtime`.

## SpaceContext

Methods such as `In(space)`, `Members`, `Local`, `List` — see `go doc github.com/MrBlindMouse/Portico/experience`.
