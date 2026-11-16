# TODO

DESIGN.md describes the framework. Section 16 of that file is the list of systems a site or another program builds. They are not framework tasks. Section 15 records undecided future research only (not shipped).

What is still missing inside the framework:

## 1. Plan sets

Not current work.

Plans are one file each. A plan set would write several of those files and print the wiring from the old `RecipeCommunity` / `RecipeAccountsSite` helpers (removed).

- **Community:** `Guestbook` + `Chat` + `Landing` with destinations `guestbook` and `chat`.
- **Accounts:** `Registration` + `Landing` with destination `register`.

Possible later shape: `portico generate` learns a plan set name that writes those files and prints the combined call site. `init` stays the one starter `main` (one scene and `greet`); it has no style flag. `portico serve -demo` merges guestbook, chat, and landing for smoke tests only.

## 2. Future research (builders / UX)

Not committed work. Normative behavior stays **characters + ANSI** via `render` and `Screen` until an item below is promoted with tests and manual updates. Full narrative: [DESIGN.md](DESIGN.md) §15.

### Terminal pixel / graphics layer

- Optional raster output (Sixel, Kitty graphics, iTerm sequences) alongside or instead of part of the text grid.
- Sites today: pre-encoded bytes in assets or `Render` output; `AssetAny` for site-owned blobs.
- Open questions: flicker with full-screen clear, resize, inspect snapshots, default protocol (Sixel vs Kitty), whether the framework adds a thin `render` helper or stays site-owned.

### Theme canvas (global background)

- Optional session-wide default background (ANSI or grid fill), beyond foreground `render.Theme` roles.
- Sites today: prefix scene output with background SGR manually.
- Open questions: interaction with `NO_COLOR`, full clear on clients, inspect fidelity.

Do not promote either item without resolving the corresponding DESIGN §15 question list.