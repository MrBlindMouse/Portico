# Hello (reference site)

Minimal **framework consumer**: merges Portico scaffolds, runs `portico.ServeDev`.

Copy this directory as a starting point for your own site, then replace scaffolds with your scenes and actions.

```bash
go run .
export PORTICO_DATA=/path/to/data   # optional persistence directory
export PORTICO_LOG_LEVEL=debug      # PTY input timing + Kitty protocol discard logs
```

SSH: `ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost`
