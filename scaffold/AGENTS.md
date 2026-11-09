# scaffold/

Optional **starter definitions** that a site `main` imports and `Merge`s. Each function returns `*experience.Builder` — not opaque generated code, and not a place to put one site's domain logic.

| Function | Adds |
|----------|------|
| `Landing(destinations)` | Root menu + `go-*` transitions |
| `Guestbook()` | Shared signature space |
| `Chat()` | Chat space + scene |
| `Registration()` | `register <user> <pass>` (needs `Options.Accounts`) |
| `NicknameCommands()` | `reserve-nick` action; wire `nick` into scene inputs |
| `Form()` | Two-step local form demo |
| `Blog()` | List/open post assets `post-<id>` |
| `Lobby()` | Ready/start lobby space |
| `RecipeCommunity()` / `RecipeAccountsSite()` | Merged starter graphs |

Scaffolds may assume scene/action id strings — merge order matters when ids collide (later `Merge` wins).

Sites should **not** fork scaffolds into `runtime`. Copy the definition into the site repo and edit it there (`examples/notes`).
