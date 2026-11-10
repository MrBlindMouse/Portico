# Hello

A probe for this repository. It is not a site to copy. Create a site with `portico init`, then `portico generate <name>` for each plan you want to own and edit.

`main` merges the guestbook, chat, and landing plans and serves them with `ServeDev`. Run it, or `go test .`, to see SSH, scenes, spaces, and broadcast without creating a site.

```bash
go run .
go test .
export PORTICO_DATA=/path/to/data   # optional persistence directory
export PORTICO_LOG_LEVEL=debug      # PTY input timing + Kitty protocol discard logs
```

SSH: `ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost`

You start on the landing menu.

| Type | What happens |
|------|----------------|
| `guestbook` | Shared signatures. `sign ships ahoy` posts `<ssh-user>: ships ahoy`. `sign` alone posts `<ssh-user>: was here`. |
| `chat` | Shared messages. `send hello there` posts `<ssh-user>: hello there`. `join-chat` marks you present. `send` alone posts nothing. |

`<ssh-user>` is the username OpenSSH sends (your local login when you omit `user@`). It lasts for that connection. A reserved nickname replaces it. In-process tests that build an anonymous session with no SSH user still show `guest`.

Those scenes have no command back to the menu. Disconnect and connect again to return to landing.

## Next steps

| Goal | Where |
|------|--------|
| Bootstrap a new repo | `portico init` + [docs/building-a-site.md](../../docs/building-a-site.md) |
| Own a plan and edit it | `portico generate <name>` |
| Agent contract | Site **AGENTS.md** from `portico init`; extended [docs/site-agents.md](../../docs/site-agents.md) |
