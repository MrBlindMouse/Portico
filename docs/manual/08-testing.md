# Testing

## Build and pure helpers

Test `Build()` validation and small pure functions in your site package without SSH:

```go
inst, err := site.BuildInstallation()
if err != nil { t.Fatal(err) }
```

## testkit

Package `github.com/MrBlindMouse/Portico/testkit` provides an in-memory engine and store:

```go
h := testkit.New(exp)
sess := h.Session("s", exp.Root())
h.Register(sess)
h.MustDo(t, sess, "go")
h.AssertScene(t, sess, "other")
h.AssertRenderedContains(t, sess, "welcome")
```

`Harness` methods call the real engine apply path; they do not reimplement dispatch.

| Method | Purpose |
|--------|---------|
| `Register(sess)` | Add session to the engine registry (needed for `Submit` and many multi-session tests) |
| `Do` / `MustDo` | Run one input line (`HandleInput`) |
| `AssertScene` | Current scene id |
| `AssertLocal` | Session-local key (`reflect.DeepEqual`) |
| `AssertPersisted` | Identity store value |
| `AssertRenderedContains` | `Engine.RenderText` contains substring |
| `SessionsIn(space)` | Registered sessions that have joined the space |
| `AssertMemberRenderedContains` | At least one member's render contains substring (broadcast checks) |

`Session` / `SessionWithIdentity` do not register or join spaces. Call `Register` before `Do`. Join spaces via effects in actions or `h.Applier().Apply`. Use `SessionsIn` only after join.

`Do` does not repaint the actor the way `RunLoop` does; use `AssertRenderedContains` or `faketerm` + `Engine.Render` when you need terminal output.

Use this for integration-style tests of actions and effects without opening a TCP port. Sites may import `testkit`; framework `runtime` tests use `fakessh` internally.

## What to skip

Full SSH end-to-end is optional for most sites. Prefer:

1. Unit tests on parsing and render helpers.
2. `Build()` and wire validation.
3. `testkit` for effect paths.

## Reload preview

Before hot reload in production, `POST /inspect/reload-preview` with a `installation.Wire` body. See [Hot reload](appendices/hot-reload.md).
