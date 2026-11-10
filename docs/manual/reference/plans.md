# Plans

`portico generate <name>` copies one **plan** into the site as ordinary Go. Edit that file. Plans are not an importable package—only the generated files in your repo.

Each plan is a whole installation: it sets a root scene and `Build()` succeeds on its own. Each plan has one job. Merge them when a site needs more than one. A nickname is identity data (`ReserveNicknameEffect`), not a plan.

| Name | File | Function | Job |
|------|------|----------|-----|
| `landing` | `landing.go` | `Landing(destinations)` | Entrance menu. Fill the map with scenes to open |
| `guestbook` | `guestbook.go` | `Guestbook()` | One shared board of signed lines. Also writes `rest.go` |
| `chat` | `chat.go` | `Chat()` | One room where members post lines. Also writes `rest.go` |
| `register` | `register.go` | `Registration()` | Create a password account. Needs accounts in Options |
| `form` | `form.go` | `Form()` | Walk a two-field form |
| `blog` | `blog.go` | `Blog()` | List and open post assets (`post-<id>`). Use `AssetDir` |
| `lobby` | `lobby.go` | `Lobby()` | A room that starts when members are ready |

`guestbook` and `chat` also write `rest.go`. If that file is already there, generating the other one needs `-force`.

`-dir` chooses the directory (default `.`). `-pkg` sets the package clause (default: the package already in that directory, or `main`). An existing file is left in place unless `-force`.

`Landing` keeps its destinations argument. The printed call site leaves the map for you to fill.

## Merge order

Later `Merge` wins when scene, action, or space **ids** collide. Merge landing last if that menu should be the entrance.

## Commands (examples)

- Guestbook: `sign <message>` (shared list plus `Persist` of the signer's last line when they have a key or account)
- Chat: `join-chat`, `leave`, `nick <name>`, `send <message>`
- Lobby: `join`, `leave`, `ready`, `unready`, `start` (`ready` uses `SpaceCAS`)
- Form: `start`, `next <value>`, `cancel` (finished answers are `Persist`ed for a key or account)
- Blog: `open <id>`, `back` (last opened id is `Persist`ed for a key or account)
- Registration: `register user pass`

Each generated file opens with a **How to change this plan** comment. That comment is the editing guide for the copy.

```bash
portico generate guestbook
portico generate -list
```

## Plan sets (not shipped)

A **plan set** would write several plan files and print wiring (see TODO.md). The CLI today is one plan per `generate` invocation.
