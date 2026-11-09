# runtime/

SSH server, session lifecycle, input dispatch, and effect application. **Site authors rarely import this** except advanced hosting (`HotReload` on `*runtime.Server`).

## Invariants

- Only `apply.go` mutates `Session` and space registry state (`session.go` updates identity nickname and render cache).
- Session constructor is unexported (`newSession`).
- Space hooks `OnJoin` / `OnLeave` run as effect lists through `Apply`.
- `SetExperience` refuses definitions that orphan live sessions.

## Prefer / avoid

- Prefer `Terminal` interface over direct SSH imports in loop code.
- Avoid importing site-specific packages or embedding domain rules.
