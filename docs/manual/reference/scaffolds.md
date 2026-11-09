# Scaffolds

Package: `github.com/MrBlindMouse/Portico/scaffold`. Each function returns `*experience.Builder` to pass to `Merge`.

| Function | Adds |
|----------|------|
| `Landing(destinations map[string]SceneID)` | Root menu, `go-*` transitions |
| `Guestbook()` | Shared signature space |
| `Chat()` | Chat space + scene |
| `Registration()` | `register <user> <pass>` — needs accounts in Options |
| `NicknameCommands()` | `reserve-nick`; bind `nick` on scenes |
| `Form()` | Two-step local form demo |
| `Blog()` | List/open assets `post-<id>` — use `AssetDir` |
| `Lobby()` | Ready/start lobby space |
| `RecipeCommunity()` | Guestbook + chat + landing |
| `RecipeAccountsSite()` | Community + registration + nicknames |

## Merge order

Later `Merge` wins when scene, action, or space **ids** collide.

## Customization

Copy scaffold source into your site repo and edit locally ([examples/notes](../../examples/notes)). Do not patch the module cache.

## Commands (examples)

- Guestbook: `sign <message>` (uses `InputLine()`)
- Chat: join/send commands (see scaffold source at your tag)
- Registration: `register user pass`

Discover snippets:

```bash
portico scaffold guestbook
portico scaffold recipe community
```
