# scaffold/

Optional **starter definitions** for site authors. Each function returns `*experience.Builder` to `Merge` into your site — not opaque generated code.

| Function | Adds |
|----------|------|
| `Landing(destinations)` | Root menu + `go-*` transitions |
| `Guestbook()` | Shared signature space |
| `Chat()` | Chat space + scene |
| `Registration()` | `register <user> <pass>` (needs `Options.Accounts`) |
| `NicknameCommands()` | `reserve-nick` action; wire `nick` into scene inputs |

Scaffolds may assume scene/action id strings — merge order matters when ids collide (later `Merge` wins).

Sites should **not** fork scaffolds into `runtime`; copy and edit definitions in the site repo instead.
