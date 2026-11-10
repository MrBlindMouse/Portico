# components/

Scenes, actions, spaces, their IDs, effects, and the context interfaces those signatures use.

`installation` imports this package to build a validated `Installation`. This package does not import `installation` or `runtime`.

## Invariants

- Actions return `[]Effect` only; they do not mutate session or space state.
- `RenderContext` / `SceneContext` / `ActionContext` are implemented only in `runtime`.
- `Scene.Render` must use `RenderContext` (no effect application during render).
