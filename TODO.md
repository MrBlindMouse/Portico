# TODO

DESIGN.md describes the framework. Section 16 of that file is the list of systems a site or another program builds. They are not framework tasks. Section 15 records undecided future research only (not shipped). Do not promote the terminal graphics layer or a theme canvas while doing the asset or tick work below.

What is still missing inside the framework:

## 1. Process-driven effects (shipped)

`Host.Submit`, `SubmitToSpace`, and `SubmitAll` apply effects and repaint. See `docs/manual/reference/effects.md`. No timer and no event bus. Item 6 is the optional loop on top of this.

## 2. Testkit assertion surface

DESIGN §7 asks for helpers that send input or invoke an action and assert on scene, rendered output, local state, persisted values, transitions, and broadcasts, plus multi-session fixtures. `docs/manual/08-testing.md` already tells sites to "drive actions, assert on scene, local state, renders". The kit does not do that yet.

**What exists.** `testkit.Harness` (`testkit/testkit.go`) holds `*runtime.Engine` and `identity.Store`, built with `NewApplier` and `NewEngine`, so apply is the real one. `Session` / `SessionWithIdentity` call `runtime.NewSessionForTest` and do **not** `RegisterSession`. `MustBuild` fails the test on `Build`. The only harness test (`testkit/testkit_test.go`) drives `Engine.RunLoop` with `internal/faketerm`. Site tests that want a one-shot action today call `Engine.HandleInput` or `Applier.Apply` and read `Session.CurrentScene`, `LocalCopy`, `LastRender`, and `Store.Get` themselves. `Engine` and `Session` are `runtime` types. A site file that only uses inferred types from `h.Session(...)` does not need its own `runtime` import; a file that names those types does. Assertions should keep the `runtime` names inside `testkit`.

**What to add.** Methods on `Harness`. They call the engine; they do not reimplement apply.

```go
func (h *Harness) Do(sess *runtime.Session, line string) error
func (h *Harness) MustDo(t *testing.T, sess *runtime.Session, line string)

func (h *Harness) AssertScene(t *testing.T, sess *runtime.Session, want components.SceneID)
func (h *Harness) AssertLocal(t *testing.T, sess *runtime.Session, key string, want any)
func (h *Harness) AssertPersisted(t *testing.T, idn identity.Identity, key string, want any)
func (h *Harness) AssertRenderedContains(t *testing.T, sess *runtime.Session, substr string)

func (h *Harness) SessionsIn(space components.SpaceID) []*runtime.Session
```

`Do` is `Engine.HandleInput`. That runs rate limits and hooks when the test set them, and it records broadcasts by rendering members. It does not render the actor (`RunLoop` does). `AssertRenderedContains` should call `Engine.RenderText`, which fills `LastRender`, rather than requiring a prior `RunLoop`. Compare `want` with `reflect.DeepEqual`. Failure text includes session id, key, got, and want.

`SessionsIn` reads space membership, which exists only after `JoinSpace` has been applied. `NewSessionForTest` does not register or join. Document that multi-session tests `RegisterSession` (already required for reload tests in `runtime/engine_reload_test.go`) and join via `Do` or apply. Add `Harness.Register(sess *runtime.Session)` as a one-liner over `Engine.RegisterSession` so the site test does not mention `runtime` for that step.

**Broadcast assertion.** There is no broadcast log. `HandleInput` renders each member and moves on. Do not add a counter in the engine for tests. Assert the observable result: after `Do` on a session that returns `BroadcastEffect`, another joined session's `RenderText` contains the updated space data. A helper can wrap that:

```go
func (h *Harness) AssertMemberRenderedContains(t *testing.T, space components.SpaceID, substr string)
```

It renders each `SessionsIn` member. An empty member list fails the test with the space id. That is the honest check. A count of `ApplyResult.Broadcast` would mean returning the result from `Do` (`Do` returns `error` today because `HandleInput` does). Optional: `Do` stays `error`; tests that need the result call `HandleInput` only if we later export a result. Not required for the first cut.

**Snapshots.** `LastRender()` is the lightweight snapshot store already. A helper that writes `LastRender` to `testdata/<name>.golden` and compares is enough. Gate updates on an env var the repo already would accept (`-update` flag on the test, or `PORTICO_UPDATE_GOLDEN`). Skip this until a plan starter test needs it. String compare via `AssertRenderedContains` covers the DESIGN bullet for most scenes.

**Transitions.** `AssertScene` after `MustDo` is the transition check. No extra helper.

**Starter tests.** DESIGN says a plan may include a starter test when a site copies it. Put one next to item 4, in the generated file's `_test.go`, using this harness. Do not add starter tests that import `runtime`.

**Constraint.** `installation` still must not import `runtime` (`internal/archtest`). `testkit` already imports `runtime`; that stays.

## 3. Asset model is text-only

DESIGN §6.6 describes named content: text, structured data, terminal-compatible graphics, audio references. DESIGN §15 leaves Kitty, Sixel, and iTerm graphics as research and says sites may embed pre-encoded sequences in assets. Do not build a graphics protocol while doing this.

**What exists.** `Builder.Asset(name, content string)` stores `map[string]string` (`installation/installation.go`). `Installation.Asset` / `Assets` / `CloneWithAssets` are string-only. `RenderContext.Asset` (`runtime/context.go`) returns `(string, bool)`. `asset.Loader` reads every regular file in a directory as `string(content)` (`asset/asset.go`). `asset.Poll` hashes those bytes and calls `Engine.UpdateAssets`, which replaces the whole map via `CloneWithAssets`. Names must be basenames (`ValidateNames`). Hot reload of an installation copies the string map. Binary files survive as a Go string (it is a byte container), and a scene can return that string. There is no typed value, and a poll replaces anything that was not in the directory.

**What to add.** Keep the string map and the directory loader as the default. Add a side registry for values the loader does not own:

```go
func (b *Builder) AssetAny(name string, v any) *Builder
func (e *Installation) AssetAny(name string) (any, bool)
```

`RenderContext` gains `AssetAny(name string) (any, bool)`, implemented in `runtime/context.go`. `Asset(name)` stays the string path. A name may exist in both maps; `Asset` does not look at the any-map. `CloneWithAssets` replaces strings and **keeps** the any-map, so a poll of `AssetDir` does not drop site-registered values. `Merge` copies both maps (later wins per name, same as strings today).

`UpdateAssets` stays `map[string]string`. A site that polls its own binary directory calls a second method, `Engine.UpdateAssetAny` or `Host` equivalent, that swaps the any-map the same way. Do not teach `Loader` to sniff file types. A site that wants images registers `[]byte` (or its own struct) at startup, or runs its own poll and calls the update method. That is the narrow hatch.

Empty name and path separators fail at `Build` or at `AssetAny`, using the same basename rule as `ValidateNames`.

**Tests.** Builder round-trip: string asset and any asset both readable after `Build`. `CloneWithAssets` drops a string key and keeps the any value. A context test (or engine `RenderText`) shows `AssetAny` on the render context. Poll test already covers strings; extend it to show a pre-registered any value still present after `UpdateAssets`.

**Docs.** `docs/manual/04-rendering.md` (or the asset part of authoring): directory files are text; `AssetAny` is for site-owned values; graphics sequences are still either a string asset or a site type. Point at DESIGN §15 so nobody treats this as the graphics layer.

**Out of scope.** Decoding images, protocol selection, audio playback, a new asset effect.

## 4. Plan extraction (shipped)

`portico generate <name> [-dir .] [-pkg name] [-force]` writes one plan into `-dir`. Names: `landing`, `guestbook`, `chat`, `register`, `form`, `blog`, `lobby`. Each plan builds on its own. `-list` prints them. `-pkg` defaults to the package clause already in `-dir`, otherwise `main`. The file is ordinary Go (`// Copied from Portico plan <name>. Edit freely.`). `guestbook` and `chat` also write `rest.go`, and a second plan that needs `rest.go` requires `-force` if that file is already there. There is no importable plans package for sites. Sources live in `internal/plans`. Plan sets are not part of this command. `portico init` stdout lists plan names; run `generate` after `init`.

## 5. Identity store version and migration

DESIGN §7 already says the file store is the default and another backend implements `Store`. DESIGN §16 says another database is a site concern. This item is the on-disk format's version, not a query API and not SQL.

**What exists.** `identity.Store` is Get/Set/ListKeys, nickname reservation, ExportAll/ImportAll (`identity/store.go`). `FileStore` layout under `DataDir`:

- `identities/<fingerprint>.json` — `{"values": { ... }}` (`idFile`). No version field.
- `nicknames.json` — `{"by_nickname": { name: fingerprint }}`. No version field.

`FileAccountStore` (`identity/accounts.go`) is a different interface, `AccountStore`: `accounts/<username>.json` as `AccountRecord` (username, bcrypt `password_hash`, optional email, linked fingerprints). Export and import exist for both. `OpenFileStore` and `OpenFileAccountStore` create directories and decode whatever JSON is there. Unknown fields are dropped by `encoding/json`. A future required field would fail unmarshal or silently change meaning. The manual (`docs/manual/01-getting-started.md`, `05-identity-accounts.md`) says the directory holds the host key, identity KV, and hashed accounts. It does not specify the JSON.

Host key path is `Options.HostKeyPath`, not part of this format.

**What to add.**

- A version marker. One file, `DataDir/store-version`, containing `1\n`. Absence of the file means version 1, which is the layout above. Do not rewrite every identity file to add a version field; existing directories must keep opening.
- `OpenFileStore` reads the marker. Version 1 proceeds. A higher version returns an error that names the file and the version, so an older binary refuses to clobber newer data.
- `identity.Migrate(dataDir string, from, to int) error`. Version 1 to 1 is a no-op that writes the marker if it was missing. Unknown pairs return an error. No automatic migration inside `OpenFileStore` beyond "missing marker is v1". A later format change adds a case here and bumps the constant.
- Account files stay beside this. If they ever change, the same marker covers the directory, and `Migrate` grows a case. Do not invent a second version file.
- `Store` gains nothing. A SQL backend does not implement `Migrate`. Document that `Migrate` is for the file layout only.

**Tests.** Temp dir with a v1 identity file and no marker opens and reads the value. After open, the marker may stay absent (do not write on every open). `Migrate(dir, 1, 1)` writes `store-version`. A marker of `2` makes `OpenFileStore` fail. `Migrate(dir, 1, 99)` fails.

**Docs.** One subsection in `docs/manual/05-identity-accounts.md`: directory listing, JSON shape, version file, and "a different backend implements `Store` / `AccountStore`". `docs/manual/01-getting-started.md` can link there instead of growing.

**Out of scope.** Schema for values inside `values`. Migrations of site-level keys. Changing `ExportAll` / `ImportAll`.

## 6. Background tick helper

This is a loop on top of `Host.Submit` (item 1, shipped). It is still unbuilt. It is not a scheduler inside `runtime`.

**What exists.** `asset.Poll` is the only periodic helper, and it updates assets, not effects. `render.SpinnerFrame` tells the site to advance the index itself. `Options.ShutdownGrace` and `Server` drain do not cancel site goroutines. `Host` has no `context` of its own today; `NewHost` starts the asset poll with `context.Background()`.

**What to add.** A function in package `portico`, not in `runtime`:

```go
func Every(ctx context.Context, interval time.Duration, submit func() error)
```

The site owns the body: build the effect slice (it may change every tick) and call `Host.Submit`, `SubmitToSpace`, or `SubmitAll`. `Every` is `time.NewTicker`, call `submit`, stop the ticker when `ctx` is done. Log or return submit errors through an optional `onErr func(error)` argument only if the first form is awkward to use; prefer the site checking the error inside `submit` so `Every` stays one loop. Document that a nil ctx panics the same way `context` usage usually does, and that `interval <= 0` is a programming error (panic or return). Pick one and test it.

Cancellation: the site passes `ctx` and cancels it from the process signal path it already owns. Do not hook `Every` to SSH shutdown inside the runtime. Document that a tick in flight may still `Submit` once after cancel, and item 1's draining check rejects it.

Do not register ticks on `Builder`. A tick is process machinery, like the asset poll, not part of the installation graph.

**Tests.** A short test with a cancellable context: `submit` runs at least once, and after cancel it stops. Use a tiny interval. No SSH.

**Docs.** Same manual section as `Submit`. Example: ticker increments a spinner index with `SetLocalEffect` on one session, or `SpaceSetEffect` plus `BroadcastEffect` through the fan-out rule in item 1.

**Out of scope.** Cron expressions, per-session timers stored in the installation, a hidden runtime loop.

## 7. Wire and codegen docs

**What exists.** `installation.Wire` is a JSON sketch: root, scene input map, action ids, space ids, optional asset names (`installation/wire.go`). `ValidateWire` builds stub scenes and runs `Build`. `PreviewReload` (`installation/preview.go`) checks that sketch against `[]LiveSession` and reports stranded sessions and missing assets. `POST /inspect/reload-preview` runs that against the live engine (`inspect/inspect.go`). `portico wire validate` and `portico wire codegen` (`cmd/portico/wire_cmd.go`, `internal/codegen/wire.go`) validate a file and emit a `main` package whose `Render` and `Run` are empty. `Host.HotReload` takes a `*installation.Builder` with real functions. Posting a wire does not swap the live installation.

There is no `ToWire`. `Installation` can already supply the pieces: `Root`, `Scenes`, `Scene(id).Inputs`, `Actions`, `Spaces`, `Assets`. `Builder` fields are unexported, so the conversion belongs on `*Installation` after `Build`, or on a built installation the engine is running.

`docs/manual/reference/wire.md` lists the JSON fields and names `PreviewReload`. It does not show how to produce a wire from Go, the curl body, or how that relates to `HotReload`. `docs/manual/appendices/hot-reload.md` has the curl. `docs/manual/reference/cli.md` lists the commands. Authoring docs mention the endpoints in one bullet.

**What to add.**

- `func (e *Installation) Wire() Wire` in `installation/wire.go`. Fill root, scenes and their input bindings (action ids as strings), action ids, space ids, asset names. Omit `Render`, `Run`, space hooks, and session hooks; the wire format has nowhere to put them. Document that. Nil installation returns the zero `Wire`.
- One test: build a small graph, `Wire()`, `ValidateWire` succeeds, and a scene input survives the round trip. `PreviewReload` of that wire against a live session on a removed scene is stranded (the existing preview tests cover the rule; this test only checks the extractor).
- Expand `docs/manual/reference/wire.md`:
  - When to use it: an agent checking a proposed graph, or `wire codegen` starting a file. Site authors who already have Go do not need a wire to call `HotReload`.
  - Produce one with `exp.Wire()` and `json.Marshal`.
  - `POST /inspect/reload-preview` with that body. `ok` is false when `stranded_sessions` is non-empty. Same rule as `SetInstallation`.
  - `portico wire validate` checks the file. `portico wire codegen` writes stubs. Applying the change is still `Host.HotReload` with a builder that contains the real funcs.
  - Asset names on the wire warn when the live installation has not loaded them. They do not upload asset bytes.

**Out of scope.** A wire that round-trips `Render` and `Run`. Applying a POST as the live installation. An interpreted experience (DESIGN §16).

## 8. Plan sets

Not current work. Do not build this while doing the items above.

Plans are one file each. A plan set would write several of those files and print the wiring from the old `RecipeCommunity` / `RecipeAccountsSite` helpers (removed).

- **Community:** `Guestbook` + `Chat` + `Landing` with destinations `guestbook` and `chat`.
- **Accounts:** `Registration` + `Landing` with destination `register`.

Possible later shape: `portico generate` learns a plan set name that writes those files and prints the combined call site. `init` stays the one starter `main` (one scene and `greet`); it has no style flag. `portico serve -demo` merges guestbook, chat, and landing for smoke tests only.
