# How a Portico site fits together

A site is an ordinary Go module. Portico is the SSH entrance. You declare **components**, `Build()` an **installation**, and serve it.

```text
go run …/cmd/portico init     site shell (one scene, one action); pin @tag before go.mod exists
go run …/cmd/portico generate one plan: owned Go you edit; no @tag once Portico is in go.mod
components            scene, action, space, effects, contexts
installation          builder, Build(), the Installation value
render                optional terminal drawing
identity              who the session is, and what persists
runtime               sessions and apply — do not import
```

`portico.New()` returns a builder. `Merge` copies components in. Later merges win on the same id. `Build()` checks that every input names a real action and that a root scene exists. **`portico.Options` in `main`** configures the server; **`Serve(addr, inst, opts)`** runs until interrupt (`portico init` template). **`ServeDev`** is an optional dev shortcut.

Effects are the only way to change session or space state. A scene's `Render` only reads. Actions return `[]components.Effect`. The runtime applies them.

**Plans** are starter files, not a library. `go run …/cmd/portico generate guestbook` writes Go into the site that you own and edit. A **plan set** (several plans plus wiring) is not shipped.

**Identity** is not a component. It is the visitor on the session, plus the stores in `Options` (key/value and accounts). The register plan creates an account through `RegisterAccountEffect`. A nickname is identity data (`ReserveNicknameEffect`), not a plan.

**Render** draws text. It does not know about sessions. Scenes call it while rendering.

Inspect and status HTTP have no authentication. `portico init` binds them to loopback by default. Set a public address only if something else authenticates the port.
