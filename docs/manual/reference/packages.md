# Package import map (sites)

| Import | Use |
|--------|-----|
| `github.com/MrBlindMouse/Portico` | `New`, `Serve`, `ServeDev`, `Options`, `NewHost`, `Host` |
| `.../experience` | Builder, scenes, actions, spaces, effects, wire |
| `.../scaffold` | Starter `Merge` builders |
| `.../render` | Terminal layout and theme |
| `.../asset` | Loaders when using `AssetDir` |
| `.../identity` | Public store types (site-chosen backends via Options) |
| `.../config` | Optional JSON operator config |
| `.../testkit` | In-memory tests |

**Do not import** from site code:

- `.../runtime`
- `.../inspect`
- `.../internal/...`

API details: `go doc` or [pkg.go.dev](https://pkg.go.dev/github.com/MrBlindMouse/Portico) at your module version.

Framework contributors also use `runtime`, `internal/archtest`, etc. See root [AGENTS.md](../../AGENTS.md).
