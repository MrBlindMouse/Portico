# Portico

## Design Document

**Status:** Aspiration (draft). Shipped behavior is the README, `docs/site-agents.md`, and `AGENTS.md`. Unshipped ideas are listed in `TODO.md`.  
**Version:** 0.5  
**Name:** Portico

---

## 1. Purpose

Portico is a framework for building interactive, multi-user terminal experiences served over SSH.

It exists to make these experiences as productive and predictable to create as Rails made web applications. The primary authors are AI agents. Humans act as directors: they describe intent, review results, and steer iteration. Portico is designed so that an agent can reliably assemble, modify, inspect, and extend a site without fighting low-level concerns such as session lifecycle, input routing, presence, rendering, or persistence.

Portico is opinionated about structure and flow. It is deliberately unopinionated about the content of the experiences themselves.

---

## 2. Goals

Portico optimizes for the following, in order:

1. **Agent productivity.** An agent should be able to go from a short description to a running, correctly structured site with minimal specification, then iterate through inspection and small edits.
2. **Adoption.** The framework should be approachable to people already building or experimenting with SSH sites, and the resulting code should be readable by both humans and agents.
3. **Long-term maintainability.** Sites and the framework itself should remain understandable and safely modifiable over time, even when many agents or contributors touch them.
4. **Structural predictability.** Any Portico site should have a recognizable shape. An agent that has seen one site should be able to navigate another.
5. **Safe extensibility.** Common cases stay on a short happy path. Advanced needs are met through explicit extension points rather than by abandoning the structure.

Secondary goals include fast local iteration, simple deployment (single binary, listen on port 22), and sufficient strictness that illegal states are difficult to represent.

---

## 3. Non-Goals

The following are explicitly out of scope:

- A fixed catalog of domain-specific features that every site must include.
- A separate authoring language, an interpreted experience file, or site code that mutates sessions outside the effect model. A site is an ordinary Go program that imports Portico and stays on the builder, scenes, actions, and effects.
- Replacement of existing low-level SSH or terminal libraries. Portico sits above them.
- Built-in support for any particular visual metaphor, game style, or interaction pattern beyond the primitive layers defined in this document.
- High-scale multiplayer infrastructure, real-time guarantees, or resource behavior under thousands of concurrent sessions.
- Full ownership/borrow checking or Rust-level aliasing control. Strictness is pursued through API design, unexported state, and automated checks, not through a different language.
- Acting as a general web framework, API server, or non-SSH application host.

---

## 4. Design Principles

### 4.1 Convention Over Configuration

Portico makes a large number of decisions once so that agents and humans do not have to remake them for every site. Session handling, identity derivation from SSH keys, input dispatch, presence in shared spaces, hot-reloading of definitions, and the relationship between scenes and actions are all conventions.

Only deviations from these conventions require explicit configuration. Following the conventions produces a working site with minimal specification.

### 4.2 Named Layers With Clear Responsibilities

Every piece of a site belongs to a known layer. An agent that has seen one site can look at another and immediately understand where scenes, actions, spaces, identity data, and assets live. This shared shape reduces cognitive load and makes sites written by different agents mutually intelligible.

### 4.3 Plans as the Default Starting Point

The most common patterns are available as **plans**. An agent's first action is `portico init`, then `portico generate`, which writes the pattern into the site as ordinary Go. The author then edits that file. Plans are not an imported library and not opaque generated artifacts.

### 4.4 Opinionated Defaults With Explicit Escape Hatches

The framework owns the runtime loop and the relationships between layers. Custom logic is possible and is expected for advanced use cases, but it attaches at defined extension points. Abandoning the structure entirely is possible but is treated as an explicit, local decision rather than the normal path.

### 4.5 Inspectability and Tight Feedback Loops

The running system exposes structured information about its current state. An agent can query what a given session currently sees, which actions are available, what shared state exists, and what the effect of a proposed change would be. Changes to definitions take effect with minimal delay so that iteration remains conversational.

### 4.6 Progressive Disclosure

Simple sites stay on the happy path and require almost no knowledge of the underlying machinery. More complex requirements can be met by dropping down to lower-level hooks. The presence of those hooks does not complicate the simple cases.

### 4.7 Structural Strictness Beyond Convention

While Portico relies on convention for productivity, it also uses the type system, API design, and automated checks to make illegal or inconsistent states difficult or impossible to represent in normal use. The goal is to make the wrong thing hard to write, not merely discouraged.

Key techniques include:

- **Unexported types and controlled constructors.** Core runtime types (sessions, spaces, scene identifiers, etc.) expose no public mutable fields. They can only be created and transitioned through package-level functions and methods. External code cannot construct a half-initialized session or directly mutate its current scene.
- **Narrow capability interfaces.** Actions and scenes receive small interfaces that expose only the operations they are permitted to perform (for example, a scene context that allows local state changes and transitions but not arbitrary space mutation). The type received defines the allowed effects.
- **Typed identifiers.** Scene, action, and space identifiers are distinct types rather than raw strings, so the compiler rejects mismatched identifiers.
- **Centralized mutation.** Actions and scenes return descriptions of intended effects. The runtime is the sole component that applies those effects to session or space state. This keeps control flow auditable and prevents scattered mutation.
- **Builders and functional options.** Configuration of experiences, spaces, and custom actions uses builders or option patterns that only yield a usable value once required elements are supplied.
- **Automated enforcement.** Custom analyzers, architecture tests, and targeted linters run in continuous integration to forbid direct field access, require approved constructors, and prevent package boundary violations. The site itself stays ordinary Go.

These measures do not provide full ownership or borrow checking. They do, however, move the framework from pure convention to a design in which most illegal states are either unrepresentable or automatically rejected before they can be merged. This supports long-term maintainability when the primary authors are agents and when multiple contributors modify sites over time.

---

## 5. Architectural Layers

Portico defines the following layers. Each has a single primary responsibility.

| Layer | Responsibility |
|-------|----------------|
| **Installation** | The top-level definition of a site. Declares which scenes exist, which spaces exist, how identity and persistence are configured, and which assets are available. |
| **Scene** | A renderable state that a session can be in. Describes what is shown and which inputs are meaningful in that context. |
| **Action** | A named operation that can be invoked from a scene. Actions may change session state, shared space state, or switch the current scene. |
| **Space** | An optional multi-user context with shared state and presence. Sessions may join and leave spaces. Spaces handle broadcasting and conflict resolution. |
| **Identity & Persistence** | Per-visitor data keyed by SSH public key fingerprint. Provides nickname reservation, saved state, and simple key-value storage. |
| **Asset** | Named content (text, structured data, terminal-compatible graphics, audio references) that scenes and actions can reference. |

The runtime owns the session lifecycle, the mapping of input to actions, the rendering of the current scene, presence within spaces, and the loading and hot-reloading of definitions. Agents rarely interact with these mechanisms directly.

---

## 6. Core Primitives

### 6.1 Session

A session represents a single SSH connection. It is isolated by default. It carries:

- An identity derived from the client’s SSH public key (or an anonymous identity if no key is presented). When the session has no account, the SSH username is the display label for that connection only.
- A current scene.
- Local state that is not shared with other sessions.
- Optional membership in one or more spaces.

### 6.2 Scene

A scene is a declarative description of a view. It specifies:

- What content is rendered (text, structured layouts, references to assets).
- An input map that binds keys, commands, or other input events to actions.
- Optional local state that exists only while the session is in this scene.

Scenes do not contain imperative rendering loops. The runtime is responsible for turning the description into terminal output and for dispatching input.

### 6.3 Action

An action is a named, side-effecting operation. Actions are the verbs of the system. An action may:

- Mutate session-local state.
- Mutate state in a space the session belongs to.
- Switch the session to a different scene.
- Broadcast a state change to the other sessions in a space.
- Read or write identity-scoped persistence.

Actions are defined declaratively where possible. When custom logic is required, it attaches at a defined extension point rather than replacing the action system. In the implementation, actions return descriptions of effects; the runtime applies them.

### 6.4 Space

A space is a named multi-user context. It provides:

- Shared state visible to all members.
- Presence (who is currently in the space).
- Broadcasting of events and state changes to members.
- Join and leave semantics.

Spaces are optional. A site that needs no multi-user interaction simply omits them. When present, the framework handles the mechanics of membership and distribution so that actions and scenes can focus on domain behavior.

### 6.5 Identity & Persistence

By default, identity is derived from the SSH public key fingerprint presented by the client. The framework provides:

- Stable identity across connections for the same key.
- Optional nickname reservation that is unique within the site.
- Per-identity key-value storage for saved state.
- The ability for actions to read and write this storage.

Anonymous sessions (those without a key) are supported but have no persistent identity. Their display label is the SSH username for that connection when the client sent one. The same label applies to a key that is not linked to an account. It is not written to the nickname store. Persistence still follows the key fingerprint or the account.

Sites that need lasting accounts may enable an optional account layer. When enabled, an account (username, password hash, and zero or more linked key fingerprints) becomes the primary identity. The runtime can accept password authentication during the SSH handshake (alone or in combination with a key) and attach the resulting account to the session. Registration and nickname management are ordinary scenes and actions. Key-only and anonymous access remain available unless a site’s policy disables them. Any further fields a site stores on an account are that site’s data.

### 6.6 Asset

Assets are named pieces of content referenced by scenes and actions. They may be static (text, structured data, terminal graphics) or generated. The framework treats them as first-class so that agents can add, replace, or reference them without embedding large payloads inside scene definitions.

---

## 7. Base Capabilities

The following capabilities belong in the core runtime so that sites and agents do not re-implement them. They follow the same philosophy as the rest of Portico: opinionated defaults, narrow extension points, and no domain-specific features.

### Session and connection
- Solid handling of terminal resize and redraw.
- Configurable idle timeout and SSH keepalives.
- Per-IP and global connection limits.
- Clean disconnect handling (leave hooks and presence removal).

### Input
- Input mapping that supports keys, sequences, and basic command entry.
- Shared line-editing and form helpers so scenes do not reinvent them.
- Optional rate limiting on action invocation.

### State and persistence
- Structured session and space state with validation on write.
- Simple conflict handling for shared state (last-write-wins by default, with an optional compare-and-set primitive).
- Export and import of identity and account data.
- A `Store` interface for identity, accounts, and key-value data, with a file-backed default under the data directory so a site runs without choosing a database. A site that wants another backend implements the interface. Session-local state and space state stay in memory. A site saves what it cares about with `Persist`.

### Observability and agent support
- Structured logs and basic metrics for connections, actions, and errors.
- Reliable inspection of running sessions, spaces, and proposed definition changes.
- Validation of definitions on load and hot-reload, with clear errors for missing references.

### Content and rendering
- Small layout helpers (columns, boxes, alignment) for scene output.
- A single place for default theme and emphasis styles.
- Hot-reload of asset files alongside definitions.

### Security and abuse resistance
- Optional allow and deny lists by key fingerprint or account.
- Basic connection and action throttling.
- Audit logging for privileged operations when accounts are enabled.

### Operations
- Graceful shutdown that finishes or cleanly drops sessions.
- Health and readiness signals for process supervisors and the optional HTTP status endpoint.
- Configuration via flags and a simple config file (listen address, data directory, log level, feature flags).

### Included helpers
- A simple form pattern for multi-step prompts. Using it is optional.
- A way for the site’s own process to submit effects without a keypress, so a timer or other signal in that process can update spaces and sessions. The site owns the timer. Portico applies the effects.

### Testing
Testing support is a base capability so that test-driven development is the low-friction path for both agents and humans.

Portico ships a testkit (separate from the production runtime) that provides:

- In-memory session fixtures that exercise the real effect-application path with no network and an in-memory store.
- Helpers to send input or invoke actions and assert on the current scene, rendered output, local state, persisted values, transitions, and broadcasts.
- Multi-session fixtures for testing spaces, presence, and shared state.
- Exposure of the same definition validation used at load time, so tests fail fast on broken references.
- Optional lightweight string snapshots for non-trivial scene output.

Plans may include a starter test when a site copies them. The testkit reuses the real runtime apply logic and `Store` interface so tests remain honest. Full end-to-end tests over a live SSH port are possible but are not the default path.

These capabilities stay inside the runtime and the identity layer.

---

## 8. Package Layout and Key Types

Portico is implemented in Go. The initial package layout keeps the public surface small while concentrating strictness in the runtime and installation packages.

```
portico/
├── portico.go                 // public entry points, builder alias, Serve
├── components/                // scenes, actions, spaces, IDs, effects, contexts
├── installation/
│   ├── installation.go        // builder and validated Installation
│   ├── wire.go
│   ├── form.go
│   └── list.go
├── runtime/
│   ├── server.go              // SSH server setup
│   ├── session.go             // Session type and lifecycle
│   ├── dispatcher.go          // input → action routing
│   └── apply.go               // applies effects from actions
├── identity/
│   └── identity.go            // Identity, Store interface
├── testkit/
│   └── testkit.go             // in-memory session fixtures and assertions
└── internal/
    └── plans/                 // plan sources; portico generate copies these into a site
        ├── guestbook.go
        ├── chat.go
        └── landing.go
```

### Key types (illustrative)

```go
// components/ids.go
type SceneID string
type ActionID string
type SpaceID string

// installation/installation.go
type Installation struct {
    scenes  map[SceneID]Scene
    actions map[ActionID]Action
    spaces  map[SpaceID]Space
}

type Builder struct{ /* ... */ }
func New() *Builder
func (b *Builder) Scene(id SceneID, s Scene) *Builder
func (b *Builder) Action(id ActionID, a Action) *Builder
func (b *Builder) Space(id SpaceID, s Space) *Builder
func (b *Builder) Build() (*Installation, error)

// components/scene.go
type Scene struct {
    Render func(ctx RenderContext) string
    Inputs map[string]ActionID
}

// components/action.go
type Action struct {
    Run func(ctx ActionContext) []Effect
}

type ActionContext interface {
    SceneContext
    Space() SpaceContext
    Persist(key string, value any) Effect
}

type Effect interface{ isEffect() }

// runtime/session.go
type Session struct {
    id       string
    identity identity.Identity
    current  components.SceneID
    local    map[string]any
    spaces   map[components.SpaceID]struct{}
}
```

### Request flow

1. An SSH connection arrives. The runtime creates a Session via an unexported constructor.
2. The session is placed in the root scene.
3. Input arrives. The dispatcher resolves the ActionID from the current scene’s input map.
4. The action’s Run method is called with a narrow ActionContext and returns a slice of Effects.
5. The runtime’s apply step is the only place that mutates Session or Space state.
6. The current scene is re-rendered and written to the terminal.

### Public usage

Shipped sites are separate Go modules that import this library. Illustrative shape (see README for the current `Serve` / `ServeDev` signatures):

```go
exp, err := portico.New().
    Scene("home", homeScene).
    Action("go-guestbook", goGuestbook).
    Scene("guestbook", guestbookScene).
    Build()

portico.Serve(":22", exp, portico.Options{})
```

This structure keeps the happy path short while ensuring that core state remains unexported, mutation is centralized, and identifiers are distinct types.

---

## 9. Agent Interaction Model

Agents building a site work in **that site's repository**, importing `github.com/MrBlindMouse/Portico`. They do not author the site inside the framework tree. There is no separate experience-file format: the definition is Go (`portico.New()`, scenes, actions, `Build`, `Serve` or `ServeDev`). Operational contract: **AGENTS.md** in the site repo (`portico init`), extended by `docs/site-agents.md` and the manual at the module tag.

Agents interact with a running site through three complementary mechanisms.

### 8.1 Declarative Definitions

The primary artifact is the site's Go installation (components, identity configuration, and asset references). Agents create and edit that code in the site module. The runtime loads the built `Installation` and can hot-reload a new one.

The definition format is designed to be easy for an agent to generate, validate, and incrementally modify. It is also human-readable so that directors can review and adjust it.

### 8.2 High-Level Operations

The framework exposes operations that agents can invoke:

- Generate a common plan (guestbook, chat space, blog reader, multiplayer lobby, etc.).
- Add or modify a scene, action, or space.
- Query the current state of a running site or of a specific session.
- Hot-reload definitions.
- Inspect available extension points.

These operations return structured results that the agent can reason about.

### 8.3 Runtime Inspection

While a site is running, an agent (or a human using an agent) can query:

- The current scene and rendered content of any session.
- Available actions in the current context.
- Shared state and membership of any space.
- Identity and persistence data for a given key.
- The effect of a proposed definition change before it is applied.

This inspection capability is what makes conversational iteration practical.

---

## 10. Plans and plan sets

Plans are the preferred way to begin after `portico init`. `portico generate <name>` writes one plan into the site as ordinary Go the author owns and edits.

Example plans (illustrative, not exhaustive):

- **Guestbook** — A scene that displays recent signatures and an action that appends a new signature with the visitor’s identity.
- **Chat Space** — A space with shared message history, presence, and actions for sending messages and joining/leaving.
- **Blog Reader** — Scenes for listing posts and reading a single post, with assets holding the post content.
- **Simple Multiplayer Lobby** — A space that tracks participants and provides actions for signaling readiness or starting a shared activity.
- **Personal Landing** — A root scene that presents a menu of available destinations and basic identity information.

Plans may be composed by hand. An agent can generate a landing page, a guestbook, and a chat space, then wire them together by editing the resulting definitions.

**Plan sets** would be higher-level bundles of plans plus recommended wiring. They are not shipped. See TODO.md.

---

## 11. Runtime Behavior

The runtime is responsible for:

- Accepting SSH connections and establishing sessions.
- Deriving identity from presented keys.
- Loading the installation definition and hot-reloading it when it changes.
- Maintaining the current scene for each session.
- Rendering scenes to the terminal and handling resize events.
- Mapping input events to actions according to the current scene’s input map.
- Executing actions and applying their effects to session state, space state, or scene transitions.
- Managing membership and broadcasting within spaces.
- Persisting and retrieving identity-scoped data.
- Exposing inspection endpoints for agents.

The runtime does not contain domain logic. Domain behavior lives in the definitions of scenes, actions, and spaces.

---

## 12. Extension Points

Custom behavior attaches at explicit points so that the majority of a site can remain in the declarative happy path.

Documented extension points include:

- Custom action implementations that go beyond declarative effects.
- Custom scene renderers for specialized output.
- Hooks on session start, session end, space join, and space leave.
- Custom identity resolution or persistence backends.
- Middleware-style interceptors for input or action execution.

Using an extension point is an explicit choice. The rest of the site continues to use the standard layers and conventions.

---

## 13. Deployment and Hosting

A site is a set of definitions plus any custom extension code, packaged with the runtime. The expected deployment model is:

- A single process that listens on the SSH port.
- Definitions and assets loaded from the local filesystem or from a configured source.
- Hot-reload of definitions without process restart for the common case.
- Optional simple HTTP endpoint that returns connection instructions and basic status so the site is discoverable.

Portico does not mandate a particular hosting provider. It aims to make the resulting artifact trivial to run on any machine that can accept inbound SSH connections.

---

## 14. Summary

Portico applies the Rails methodology—convention over configuration, named layers, scaffolding, opinionated defaults with escape hatches, structural strictness through types and API design, and a shared structural language—to the domain of interactive SSH experiences.

Agents and humans are the authors, working in their own module that imports Portico. They define scenes, actions, spaces, identity, and assets in Go. The runtime owns session lifecycle, rendering, input dispatch, presence, and persistence. Common patterns are written into the site with `portico generate` and then edited. Custom logic stays on those extension points.

The result is a system in which creating, iterating on, and maintaining an SSH site is predictable, inspectable, and fast, while remaining open to experiences that the original designers did not anticipate. The name Portico reflects its role as the structured entrance through which SSH sessions arrive and through which agents and humans build what lies beyond.

---

## 15. Future research (not shipped)

The following items are under consideration only. They are **not** part of the shipped plan catalog, **not** in `TODO.md` as committed work, and **not** implemented until an explicit design decision promotes them.

### Terminal pixel / graphics layer

Today, scene output is a **character grid**: `render` helpers, `Screen`, ANSI theme roles, and one composable caret rectangle. That model stays the default even if a graphics layer exists later.

**Research topic:** an optional **pixel or raster layer** (sometimes described as a separate engine) that would emit **terminal-specific graphics protocols** (for example Kitty graphics, Sixel, or iTerm image sequences) on the same SSH session, alongside or instead of part of the text grid for a region.

**Status:** undecided. May remain out of Portico entirely (sites embed escape sequences in assets or scene strings).

**Motivation:** SSH sites such as `ssh pieter.com` show **inline raster art** on connect (posters, menus). That effect is not the character `Screen`; it is **graphics escape output** on the PTY. Portico sites can approximate it today by **embedding pre-encoded sequences in assets or `Render` output**; there is no first-class helper yet.

**Protocol survey (research, not a commitment):**

| Approach | Role | SSH / heterogeneous clients | Fidelity | Caveats |
|----------|------|-----------------------------|----------|---------|
| **Sixel** | Bitmap strips in a DCS (`ESC P … q … ST`) | **Widest practical default** over SSH | Palette (~256 colors), dithering | **Ghostty does not support Sixel** (Kitty graphics instead). **tmux** often needs passthrough; nested sessions are awkward. |
| **Kitty graphics** | APC `ESC _ G …` (PNG/RGBA, placement, z-index) | Best on Kitty, WezTerm, Ghostty, Konsole | True color, placement, animation | Spec-rich; fewer “any SSH client” guarantees. See [Kitty graphics protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/). |
| **iTerm2 inline (OSC 1337)** | Base64 image in an OSC | macOS / iTerm2 ecosystem | Good | Narrow; third fallback at most. |
| **Chafa symbol modes** | Half-blocks and Unicode mosaics in **text** | **Universal** (no graphics terminal) | Lower resolution; reads as “pixel-ish” | Not true pixels; fits shipped `render` without a new engine. |

Consensus from terminal maintainers and tools (2024–2025): **Sixel = safe fallback for SSH**; **Kitty = best when the terminal is known**; many programs ship **Sixel + Kitty + text fallback**.

**Libraries and tools (prefer reuse; do not hand-roll encoders):**

| Layer | Suggestion | Notes |
|-------|------------|-------|
| **Build-time assets** | [Chafa](https://hpjansson.org/chafa/) CLI | `chafa -f sixel` / `-f kitty` → store under `asset`; **zero encoder code** in the server. Fits landing posters. |
| **Go, multi-protocol** | [blacktop/go-termimg](https://github.com/blacktop/go-termimg) | Kitty, Sixel, iTerm2, half-block fallback; sizing and detection. Strong default if Portico adds a thin helper. |
| **Go, Sixel-only** | [mattn/go-sixel](https://github.com/mattn/go-sixel) | Pure Go encoder; used by go-termimg’s Sixel path. |
| **Go, lighter multi** | [kenshaw/rasterm](https://github.com/kenshaw/rasterm) | Kitty, iTerm, Sixel. |
| **Go, Kitty-focused** | [SerenaFontaine/kgp](https://serenafontaine.github.io/kgp/), [dolmen-go/kittyimg](https://pkg.go.dev/github.com/dolmen-go/kittyimg) | Placement and protocol detail when standardizing on Kitty/WezTerm/Ghostty. |
| **Reference C encoder** | [libsixel](https://github.com/saitoha/libsixel) | What `img2sixel` uses; prefer Go or Chafa in this tree unless CGO is explicitly chosen. |

**Site patterns under study (no framework API yet):**

1. **Precomputed** Sixel (or Kitty) blob in `asset`, prepended at landing `Render` / `Screen.Draw`.
2. **Runtime encode** only when the poster is dynamic; cache bytes in session-local or on disk; use go-termimg or mattn/go-sixel.
3. **Fallback scene** with Chafa symbol output or plain text when graphics are unavailable.
4. Optional **second asset** (Kitty variant) for clients that do not render Sixel.

**If promoted, a lean framework shape might be:** a `render` helper that returns **opaque bytes** to draw (asset name and/or `image.Image`, optional protocol hint), backed by **precomputed Sixel** first and a Go library second—not a new encoder in Portico.

**Portico-specific constraints (unchanged by research):**

- Each paint uses **full-screen clear** then scene text; graphics may **flicker, scroll, or vanish** depending on the client—posters often want a dedicated landing scene or acceptance of redraw cost.
- The runtime **filters Kitty APC noise on input**; emitting graphics on output is a separate concern.
- **Inspect** and string snapshots may embed large escape blobs; tests should treat graphics as optional.
- **Resize** does not reflow raster art; precompute per size or re-encode on resize (expensive).

**Questions to resolve before any promotion:**

- Whether the layer belongs in the framework or stays site-owned, given client fragmentation and the runtime’s non-goal of becoming a full VT emulator.
- How graphics interact with **full-screen clear**, **resize**, and the **single caret** (hit targets, focus, and whether input stays line-oriented).
- Whether **inspect** and string snapshots remain meaningful when output is partly non-text.
- Minimum viable API: one protocol vs a pluggable backend, and how that aligns with “lean primitives” and agent predictability.
- Whether the default recommendation is **Sixel assets + symbol fallback**, with Kitty as an optional second asset or detected upgrade path.

Until those are answered, agents and sites should assume **characters and ANSI only**. See README and `docs/site-agents.md` for what is shipped.

### Theme canvas (global background)

Today, `render.Theme` provides **foreground emphasis roles** (bold, muted, error, accent, ok, warn, inverse). Each role wraps a string with ANSI and resets SGR at the end. The runtime clears the screen with `ESC [2J` / home before each frame; **empty cells use the terminal default**, not a site-chosen color. There is no theme field for a session-wide background.

**Research topic:** an optional **canvas** on `Theme` (or an equivalent hook on `Options`) that would set a **default background** for the PTY for that session—typically one ANSI background sequence emitted after clear, or a convention for filling the `Screen` grid so unpainted cells match.

**Status:** undecided. Sites can prefix scene output with background SGR manually today; that is not inspect-friendly and does not cover the caret slot or the default `>` row unless the site styles those too.

**Questions to resolve before any promotion:**

- Whether canvas is one escape per frame, per-row fill on `Screen`, or both, and how it interacts with **NO_COLOR** and `render.Plain`.
- Whether **full clear** on some clients wipes background attributes and makes a global color flicker or fail.
- Whether **inverse** and per-role foreground colors remain the only text styling, or canvas implies default foreground as well.
- Whether **inspect** snapshots and tests should strip or record canvas sequences.

Until those are answered, agents and sites should assume **no global background in the theme API**. Foreground roles via `ctx.Theme()` remain the shipped model.

---

## 16. Built outside Portico

Portico stops at the blocks above. A site, or a program next to the site, owns anything that chooses a product. None of the following is a missing framework feature:

- Another database, such as SQLite, behind `Store`.
- Mail, including SMTP and email verification. The site’s scenes collect an address if they want one. Something else sends the mail.
- Payments, search, media streaming, and other domain services.
- A message bus, topic system, or plugin host. Cross-session updates are a space write and a broadcast. A clock or webhook in the site submits effects through the hook in §7.
- More than one process, or a cluster. Spaces live in the one SSH process.
- A login wall on the inspect or status port. Those ports are localhost tools for the author.
- An interpreted or generated experience that replaces the Go definition.
