# Hello

Framework **smoke test** for CI and local verification. It is not a site to copy. Create a site with `go run github.com/MrBlindMouse/Portico/cmd/portico@v0.1.0 init`, then `go run github.com/MrBlindMouse/Portico/cmd/portico generate <name>` in that module for each plan you want to own and edit.

`main` merges every plan (guestbook, chat, register, form, lobby, blog) behind a landing menu and serves them with `ServeDev`. A dev plaintext account store is on so `register` can create an account. One blog post is seeded as `post-1.md`.

```bash
go run .
go test .
export PORTICO_DATA=/path/to/data   # optional persistence directory
export PORTICO_LOG_LEVEL=debug      # PTY input timing + Kitty protocol discard logs
```

SSH: `ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost`

You start on the landing menu. Those scenes have no command back to the menu. Disconnect and connect again to return.

| Type | What happens |
|------|----------------|
| `guestbook` | Shared signatures. `sign ships ahoy` posts `<ssh-user>: ships ahoy`. `sign` alone posts `<ssh-user>: was here`. A key or account also stores that line for the next connection. |
| `chat` | Shared messages. `send hello there` posts `<ssh-user>: hello there`. `join-chat` marks you present. `leave` drops you. `nick <name>` sets the display name when you have a key or account. `send` alone posts nothing. |
| `register` | `register <username> <password>` creates a dev account. Reconnect to sign in. |
| `form` | `start`, then `next <value>` for name and message. `cancel` clears the walk. A finished form is stored for a key or account. |
| `lobby` | `join`, then `ready` / `unready`. `leave` drops you. `start` runs when everyone joined is ready. |
| `blog` | Lists `post-*` assets. `open 1.md` shows the seeded post. `back` returns to the list. A key or account keeps the last opened id. |

`<ssh-user>` is the username OpenSSH sends (your local login when you omit `user@`). It lasts for that connection. A reserved nickname replaces it. In-process tests that build an anonymous session with no SSH user still show `guest`.

## Next steps

| Goal | Where |
|------|--------|
| Bootstrap a new repo | `portico init` + [docs/building-a-site.md](../../docs/building-a-site.md) |
| Own a plan and edit it | `go run github.com/MrBlindMouse/Portico/cmd/portico generate <name>` |
| Agent contract | Site **AGENTS.md** from `portico init`; extended [docs/site-agents.md](../../docs/site-agents.md) |
