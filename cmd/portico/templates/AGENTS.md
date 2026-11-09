# AGENTS.md (Portico SSH site)

You are building a site that **imports** `github.com/MrBlindMouse/Portico`. This repo is not the framework.

## Before you change Portico usage

1. Read `require github.com/MrBlindMouse/Portico` in **go.mod** and note the version tag.
2. Read the upstream contract at that tag:  
   `https://github.com/MrBlindMouse/Portico/blob/<tag>/docs/site-agents.md`
3. If still unsure: `go doc github.com/MrBlindMouse/Portico/experience`

Do not assume APIs from DESIGN or from a different Portico version.

## Run loop

```bash
go run .
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
curl -s http://127.0.0.1:9922/inspect/snapshot
```

## Hard rules

- Actions return `[]experience.Effect` only; scenes only `Render`.
- Do not import `.../runtime`, `.../internal/...`, or `.../inspect`.
- Domain code stays in this repo; copy scaffolds here to customize.

Details: Portico `docs/site-agents.md` at your module tag.
