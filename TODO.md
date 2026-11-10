# TODO

DESIGN.md describes the framework. Section 16 of that file is the list of systems a site or another program builds. They are not framework tasks. Section 15 records undecided future research only (not shipped). Do not promote the terminal graphics layer or a theme canvas while doing the asset or tick work below.

What is still missing inside the framework:

## 1. Process-driven effects (shipped)

`Host.Submit`, `SubmitToSpace`, and `SubmitAll` apply effects and repaint. See `docs/manual/reference/effects.md`. `portico.Every` is the optional ticker loop on top (item 6, shipped). No event bus.

## 2. Testkit assertion surface (shipped)

`testkit.Harness` exposes `Register`, `Do`, `MustDo`, `AssertScene`, `AssertLocal`, `AssertPersisted`, `AssertRenderedContains`, `SessionsIn`, and `AssertMemberRenderedContains`. See `docs/manual/08-testing.md` and `testkit/harness_test.go`.

**Still out of scope (v1):** golden `testdata` helpers; `Do` returning `ApplyResult`; starter `_test.go` in generated plans until a plan needs it.

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

## 6. Background tick helper (shipped)

`portico.Every(ctx, interval, submit)` in package `portico`. Site-owned context and callback (`Host.Submit`, `RepaintScene`, …). See `docs/manual/reference/effects.md` and `docs/manual/06-running.md`. Not wired to SSH shutdown; not part of `Builder`.

**Out of scope:** cron, per-session timers in the installation, a hidden `runtime` loop.

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
