# Portico SSH site

```bash
go run .
```

SSH (dev, default `-addr :2222`): `ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost`

Production listen address: `go run . -addr 203.0.113.10:22` (bind a dedicated IP; see upstream README **Production**).

Optional: `PORTICO_DATA`, `PORTICO_LOG_LEVEL=debug`

Agents: see [AGENTS.md](AGENTS.md). Manual: [docs/manual](https://github.com/MrBlindMouse/Portico/blob/main/docs/manual/README.md) (pin to your `go.mod` tag). Short guide: [building-a-site](https://github.com/MrBlindMouse/Portico/blob/main/docs/building-a-site.md).
