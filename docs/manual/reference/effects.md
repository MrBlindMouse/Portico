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
