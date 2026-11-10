# Package import map (sites)

| Import | Use |
|--------|-----|
| `github.com/MrBlindMouse/Portico` | `New`, `Serve`, `ServeDev`, `Options`, `NewHost`, `Host` |
| `.../components` | Scenes, actions, spaces, IDs, effects, contexts |
| `.../installation` | Builder, `Build()`, `Installation`, wire, form/list helpers |
| `.../render` | Terminal layout and theme |
| `.../asset` | Loaders when using `AssetDir` |
| `.../identity` | Public store types (site-chosen backends via Options) |
| `.../config` | Optional JSON operator config |
| `.../testkit` | In-memory tests |

Plans are not an import. `portico generate` writes them into the site. See [Plans](plans.md).

Do not import `runtime` or `inspect` from site code.

API details: `go doc` or [pkg.go.dev](https://pkg.go.dev/github.com/MrBlindMouse/Portico) at your module version.

Framework contributors: root [AGENTS.md](../../AGENTS.md).
