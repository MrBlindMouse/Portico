# TODO

DESIGN.md is the aspiration. This list is what is not the current path. Do not treat either file as a backlog of work that must land for the framework to be usable.

- **SQLite store.** File JSON under `DataDir` is the default. A site that wants SQLite implements `identity.Store` (and `identity.AccountStore` if it needs accounts).
- **In-process event bus.** Cross-session updates are space writes plus `BroadcastEffect`. A bus would be a second channel for the same job.
- **SMTP and email verification.** `AccountRecord` can store an email. Sending mail and proving it stays outside the process, on a site extension.
- **Transactional apply.** One failed effect leaves earlier effects in that slice applied. Wrapping a slice in a transaction is a larger runtime change.
- **Opaque identifier types.** `SceneID`, `ActionID`, and `SpaceID` are distinct string types. Opaque structs would make `Scene("home", ...)` harder without stopping mismatched string constants.
- **Parameterized scaffold ids.** Starters use fixed ids so a recipe can wire them. Renaming is copy-and-edit (`examples/notes`).
- **Inspect authentication.** Inspect and status bind to localhost by default and have no auth. Exposing them is an operator choice, not a framework login system.
- **Snapshotting session and space memory.** Identity `Persist` is the durable path. Session-local and space state stay in memory for the life of the process.
- **Richer testkit.** The kit builds an in-memory engine and store. Assertion helpers, string snapshots, and tests emitted by scaffolds are not part of the kit.
- **Registry code generation.** `portico wire` validates a JSON sketch and can emit a stub `main`. The site definition remains ordinary Go.
