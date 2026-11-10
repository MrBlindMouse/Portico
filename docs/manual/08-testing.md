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
eng := testkit.New(exp)
// drive actions, assert on scene, local state, renders
```

Use this for integration-style tests of actions and effects without opening a TCP port. Sites may import `testkit`; framework `runtime` tests use `fakessh` internally.

## What to skip

Full SSH end-to-end is optional for most sites. Prefer:

1. Unit tests on parsing and render helpers.
2. `Build()` and wire validation.
3. `testkit` for effect paths.

## Reload preview

Before hot reload in production, `POST /inspect/reload-preview` with a `installation.Wire` body. See [Hot reload](appendices/hot-reload.md).
