# Portico site agents

Contract for AI agents and operators building an **SSH site** that imports `github.com/MrBlindMouse/Portico`. You are in a **consumer repo**, not the framework tree.

## Version alignment

Read `require github.com/MrBlindMouse/Portico` in **this repo’s `go.mod`**. When unsure about an API, prefer that version over training data:

- Upstream contract: `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/site-agents.md`
- Human guide: `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/building-a-site.md`
- Offline API: `go doc github.com/MrBlindMouse/Portico/experience`

Use the **same git tag** as the module version (not `main` unless this repo tracks main).

**DESIGN.md** is vision; if something appears only in DESIGN and not in README, godoc, or these docs, treat it as **not shipped**.

## Hard rules

1. **Actions return effects only** (`[]experience.Effect`). Never mutate session or space state in site code.
2. **Scenes only render** with `experience.RenderContext`. No state changes in `Render`.
3. **Do not import** `github.com/MrBlindMouse/Portico/runtime`, `.../internal/...`, or `.../inspect` from site code.
4. **Typed IDs** — `experience.SceneID`, `ActionID`, `SpaceID`.
5. **`Build()` validates** — fix builder errors before chasing runtime bugs.
6. **Domain lives here** — scenes, actions, spaces, and assets in this repo; copy scaffolds to edit, do not fork framework packages.

## Layout

| Piece | Typical location |
|-------|------------------|
| `Serve` / `ServeDev`, options | `main.go` |
| Experience definition | `main.go` or `site/` package |
| Data + host key | `Options.DataDir` (e.g. `.portico-data`) |
| Static content | `Options.AssetDir` when using assets |

Entry pattern:

```go
exp, err := portico.New(). /* scenes, actions, Merge(scaffold.*) */ .Build()
portico.ServeDev(exp, portico.Options{DataDir: ".portico-data"})
```

## Iteration loop

```bash
go run .
# terminal 2
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
# terminal 3 (while server runs)
curl -s http://127.0.0.1:9922/inspect/snapshot | jq .
```

| Endpoint | Method | Use |
|----------|--------|-----|
| `/inspect/snapshot` | GET | Scenes, sessions, spaces summary |
| `/inspect/sessions` | GET | All sessions |
| `/inspect/session/{id}` | GET | One session render preview |
| `/inspect/spaces` | GET | Shared space keys |
| `/inspect/identity/{fingerprint}` | GET | Per-key persistence |
| `/inspect/validate` | POST | Validate a builder JSON body |
| `/inspect/reload-preview` | POST | Dry-run hot reload |

CLI helper (from dev machine with `go install` on Portico): `portico inspect -url http://127.0.0.1:9922 -resource snapshot`

Environment: `PORTICO_DATA`, `PORTICO_LOG_LEVEL` (when wired in `main`). Do not expose inspect/status HTTP on untrusted networks.

## Contexts and effects

| Context | Where | Read / do |
|---------|-------|-----------|
| `RenderContext` | `Scene.Render` | Identity, local, assets, space state |
| `ActionContext` | `Action.Run` | Above + `InputLine()`, space helpers, persist effects |

Common effects: `SetLocalEffect`, `TransitionEffect`, `PersistEffect`, `JoinSpaceEffect`, `LeaveSpaceEffect`, `SpaceSetEffect`, `AppendSpaceListEffect`, `BroadcastEffect`, `ReserveNicknameEffect`, `RegisterAccountEffect` (requires account store in options).

Cross-session “events”: update space data, then `BroadcastEffect`. There is no in-process event bus.

## Input dispatch

Scene `Inputs` map **keys** to `ActionID`:

- Exact line match, or **first word** prefix (`register user pass` → key `register`).
- Multi-word commands: use `ctx.InputLine()` in the action.

## Scaffolds

Merge with `portico.New().Merge(scaffold.Guestbook())` etc. Later merges win on ID collisions. Session start and end hooks append.

| Function | Adds |
|----------|------|
| `Landing(destinations)` | Root menu, `go-*` transitions |
| `Guestbook`, `Chat` | Shared spaces |
| `Registration` | Needs `Options.Accounts` |
| `NicknameCommands` | Bind `nick` on your scenes. `RecipeAccountsSite` does this |
| `Form`, `Blog`, `Lobby` | Demos / assets |
| `RecipeCommunity`, `RecipeAccountsSite` | Pre-wired graphs |

Hints: `portico scaffold <name>` prints merge snippets. See `scaffold/AGENTS.md` in the module at your tag.

`Guestbook` and `Chat` read the rest of the line (`sign hello`, `send hello`) via `InputLine()`. To change a scaffold, copy it into the site and edit that copy (`examples/notes`). Do not patch the module.

## Identity

A session that connects with no credentials is a guest. That path stays open when the site has no key allow/deny list, no password accounts, and `RequireAccount` is false, so `ssh` with no key still works.

SSH clients stop when "none" succeeds, so a key or password is only used once that path is off. Turning on accounts, a key list, or `RequireAccount` does that. A public key stays a key identity unless the fingerprint is linked to an account. A password login is the account named by the SSH user. `RequireAccount` rejects guests and keys that are not linked.

## Testing

Site integration tests can use **`testkit`** (`testkit.New(exp)`) for an in-memory engine and store. Prefer testing pure helpers and `Build()` validation when that is enough; full SSH e2e is optional.

## Production notes

- Dev: SSH **:2222**. Production may use `:22` + `setcap` (see README).
- Dev `Accounts` store is plaintext; production should use `identity.FileAccountStore` (bcrypt), or `HashedAccounts` with `DataDir`. Account store precedence is `AccountBackend`, then hashed files, then `AccountStore`, then `Accounts.Accounts`.

## Framework vs site docs

| Doc | Audience |
|-----|----------|
| This file | Site repo agents |
| `docs/building-a-site.md` | Humans + agents (how-to) |
| Framework `AGENTS.md` | Portico contributors only |
