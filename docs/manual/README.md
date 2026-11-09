# Portico manual

Operator and reference documentation for **site authors** who import `github.com/MrBlindMouse/Portico`. Pin this tree to the **same git tag** as the `require` line in your site's `go.mod`.

**Normative contract:** If a feature is not described here, in [building-a-site.md](../building-a-site.md), [site-agents.md](../site-agents.md), the [README](../../README.md), or `go doc`, treat it as **not shipped**—even if it appears in [DESIGN.md](../../DESIGN.md).

## How to use this manual

| Track | Chapters | Use when |
|-------|----------|----------|
| **Tutorial** | [Getting started](01-getting-started.md) → [Authoring](02-authoring.md) | First site |
| **Guide** | [Input](03-input.md), [Rendering](04-rendering.md), [Identity](05-identity-accounts.md) | Daily development |
| **Reference** | [reference/](reference/) | Lookup APIs, HTTP, CLI, config |
| **Operations** | [Running](06-running.md), [Production](07-production.md) | Deploy and run |

Discovery from the CLI (replace `vX.Y.Z` with your module version):

```bash
portico docs -tag vX.Y.Z
```

## Contents

### Guides

1. [Getting started](01-getting-started.md)
2. [Authoring an experience](02-authoring.md)
3. [Input and dispatch](03-input.md)
4. [Rendering](04-rendering.md)
5. [Identity and accounts](05-identity-accounts.md)
6. [Running the server](06-running.md)
7. [Production and operations](07-production.md)
8. [Testing](08-testing.md)

### Reference

- [Effects](reference/effects.md)
- [Contexts](reference/contexts.md)
- [portico.Options](reference/options.md)
- [JSON config file](reference/config-file.md)
- [Inspect HTTP API](reference/inspect-http.md)
- [Status HTTP API](reference/status-http.md)
- [CLI (`portico`)](reference/cli.md)
- [Scaffolds](reference/scaffolds.md)
- [Wire format](reference/wire.md)
- [Import map (packages)](reference/packages.md)

### Appendices

- [Hot reload](appendices/hot-reload.md)
- [Glossary](appendices/glossary.md)

## Related docs

| Doc | Audience |
|-----|----------|
| [building-a-site.md](../building-a-site.md) | Short how-to |
| [site-agents.md](../site-agents.md) | AI agent contract |
| [README.md](../../README.md) | Project entry + options summary |
| [AGENTS.md](../../AGENTS.md) | Framework contributors only |
