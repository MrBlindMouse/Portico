# Glossary

| Term | Meaning |
|------|---------|
| **Installation** | Immutable validated whole: components, assets, hooks; passed to `Serve` / hot-reload |
| **Component** | One scene, action, or space (building block); types in `components` package |
| **Scene** | Render + input bindings (+ optional key mode) |
| **Action** | Handler returning effects for one dispatch |
| **Effect** | Intent description applied only in runtime (`components` package) |
| **Space** | Shared multi-user context with membership |
| **Identity** | Guest, SSH key fingerprint, or account-linked user |
| **Plan** | One starter installation from `go run …/cmd/portico generate`: its own root, one purpose, owned Go you edit |
| **Plan set** | Several plans plus recommended wiring (not shipped; see TODO) |
| **Apply** | Runtime step that mutates session/space state |
| **Wire** | JSON wiring sketch without Go funcs |
| **Host** | Running server handle with `HotReload` |
| **Inspect** | Unauthenticated HTTP JSON; loopback only if you bind it there |
| **Stranded session** | Live session whose scene/space would disappear on reload |
