# Effects reference

Actions return `[]components.Effect`. **Only** `runtime` apply mutates sessions and spaces. Effect slices are **not transactional**: if one effect in a slice fails, earlier effects in that slice may already be applied.

| Type | Helper | Purpose |
|------|--------|---------|
| `SetLocal` | `SetLocalEffect(key, value)` | Session-local KV |
| `Transition` | `TransitionEffect(sceneID)` | Change current scene |
| `Persist` | `PersistEffect(key, value)` | Identity-scoped storage |
| `JoinSpace` | `JoinSpaceEffect(spaceID)` | Enter shared space |
| `LeaveSpace` | `LeaveSpaceEffect(spaceID)` | Leave space |
| `SpaceSet` | `SpaceSetEffect(space, key, value)` | Space KV (last-write-wins) |
| `SpaceCAS` | `SpaceCASEffect(space, key, expect, value)` | Set if current value matches |
| `AppendSpaceList` | `AppendSpaceListEffect(space, key, value)` | Append to list in space |
| `Broadcast` | `BroadcastEffect(spaceID)` | Re-render members in space |
| `ReserveNickname` | `ReserveNicknameEffect(name)` | Unique nickname for identity |
| `RegisterAccount` | `RegisterAccountEffect(user, pass)` | Create account (store required) |

## Space hooks

`components.Space` may define:

- `OnJoin`, `OnLeave` — effect lists run through apply when membership changes
- `ValidateSet` — reject invalid space writes

## Scene validation

`Scene.ValidateLocal` may reject session-local writes during apply.

## Cross-session pattern

```
SpaceSetEffect / AppendSpaceListEffect → BroadcastEffect
```

No in-process event bus.

## Submit from the process

Actions are not the only caller. A site goroutine (from `Options.OnReady`, or one the site starts) may apply the same effect values through `portico.Host`:

| Method | Targets |
|--------|---------|
| `Submit(sessionID, effects)` | That registered session. Unknown id returns `no such session`. |
| `SubmitToSpace(space, effects)` | Current members of the space. |
| `SubmitAll(effects)` | Every registered session. |

The site owns the trigger. Portico does not run a timer. `SubmitAll(nil)` repaints every registered session and applies nothing. `portico.Every` is still unbuilt.

`SpaceSet`, `SpaceCAS`, `AppendSpaceList`, and `Broadcast` run **once**. `SetLocal`, `Transition`, `Persist`, `JoinSpace`, `LeaveSpace`, `ReserveNickname`, and `RegisterAccount` run **once per target**. Applying the whole slice once per session would append a list many times.

An empty member list or an empty session list applies nothing. To write a space nobody has joined, call `Submit` on any live session and include the space effects in that slice.

After a successful apply, each target is repainted, and so is every member named by `Broadcast`. The line editor is restored. Action hooks, the action rate limiter, and the action counter are not used.

Calls return `portico: shutting down` once the server is draining.

A `Transition` that turns `Keys` on or off is noticed on the next read. The read already in progress is not interrupted.

Apply is still not a transaction. If a later session rejects `ValidateLocal`, earlier sessions in that call are already updated. Those earlier sessions are repainted before the error is returned. The session that failed is repainted too, so any effect in its slice that applied before the failure is visible.
