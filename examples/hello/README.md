# Hello (reference consumer)

Lives in the Portico module so you can read a finished `main`. Your SSH site is a **separate module** that imports `github.com/MrBlindMouse/Portico` (`portico init`). Do not add your scenes under `examples/`.

This `main` is `scaffold.RecipeCommunity()` + `portico.ServeDev` — the same graph as `portico init` (default `-style community`).

```bash
go run .
export PORTICO_DATA=/path/to/data   # optional persistence directory
export PORTICO_LOG_LEVEL=debug      # PTY input timing + Kitty protocol discard logs
```

SSH: `ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost`

You start on the landing menu.

| Type | What happens |
|------|----------------|
| `guestbook` | Shared signatures. `sign ships ahoy` posts `guest: ships ahoy`. `sign` alone posts `guest: was here`. |
| `chat` | Shared messages. `send hello there` posts `guest: hello there`. `join-chat` marks you present. `send` alone posts nothing. |

Those scenes have no command back to the menu. Disconnect and connect again to return to landing.

## Next steps

| Goal | Where |
|------|--------|
| Custom scene/action (no scaffolds) | [examples/from-scratch](../from-scratch) |
| Copy a scaffold and change it | [examples/notes](../notes) |
| Bootstrap a new repo | `portico init` + [docs/building-a-site.md](../../docs/building-a-site.md) |
| Agent contract | [docs/site-agents.md](../../docs/site-agents.md) |
