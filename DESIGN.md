# Portico

## Design Document

**Status:** Draft  
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
- A general-purpose programming language or arbitrary code-execution model as the primary authoring interface. Agents work with declarative definitions and narrow extension points.
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

### 4.3 Scaffolding as the Default Starting Point

The most common patterns are available as high-level scaffolds. An agent’s first action is typically to scaffold one or more of these patterns and then refine the resulting definitions. Scaffolds produce real, inspectable, editable definitions rather than opaque generated artifacts.

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
- **Automated enforcement.** Custom analyzers, architecture tests, and targeted linters run in continuous integration to forbid direct field access, require approved constructors, and prevent package boundary violations. Code generation is used for repetitive registries so wiring cannot drift.

These measures do not provide full ownership or borrow checking. They do, however, move the framework from pure convention to a design in which most illegal states are either unrepresentable or automatically rejected before they can be merged. This supports long-term maintainability when the primary authors are agents and when multiple contributors modify sites over time.

---

## 5. Architectural Layers

Portico defines the following layers. Each has a single primary responsibility.

| Layer | Responsibility |
|-------|----------------|
| **Experience** | The top-level definition of a site. Declares which scenes exist, which spaces exist, how identity and persistence are configured, and which assets are available. |
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

- An identity derived from the client’s SSH public key (or an anonymous identity if no key is presented).
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
- Emit events that other sessions or spaces can observe.
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

Anonymous sessions (those without a key) are supported but have no persistent identity.

Sites that need lasting accounts may enable an optional account layer. When enabled, an account (username, password hash, optional verified email, and zero or more linked key fingerprints) becomes the primary identity. The runtime can accept password authentication during the SSH handshake (alone or in combination with a key) and attach the resulting account to the session. Registration, nickname management, and optional email verification are expressed as ordinary scenes and actions; sending mail is an explicit extension point. Key-only and anonymous access remain available unless a site’s policy disables them.

### 6.6 Asset

Assets are named pieces of content referenced by scenes and actions. They may be static (text, structured data, terminal graphics) or generated. The framework treats them as first-class so that agents can add, replace, or reference them without embedding large payloads inside scene definitions.

---

## 7. Base Capabilities

The following capabilities belong in the core runtime so that sites and agents do not re-implement them. They follow the same philosophy as the rest of Portico: opinionated defaults, narrow extension points, and no domain-specific features.

### Session and connection
- Solid handling of terminal resize and redraw.
- Configurable idle timeout and SSH keepalives.
- Per-IP and global connection limits.
- Clean disconnect handling (leave hooks, presence removal, pending persistence).

### Input
- Input mapping that supports keys, sequences, and basic command entry.
- Shared line-editing and form helpers so scenes do not reinvent them.
- Optional rate limiting on action invocation.

### State and persistence
- Structured session and space state with validation on write.
- Simple conflict handling for shared state (last-write-wins by default, with an optional compare-and-set primitive).
- Export and import of identity and account data.
- A built-in SQLite store as the zero-config default for identity, accounts, and key-value data. The Store interface allows replacement with another backend when needed. Session-local and pure space state remain in memory unless explicitly snapshotted.

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
- A simple built-in form scene pattern for registration, checkout, and similar flows.
- An in-process event bus so background ticks or external signals can update spaces and sessions.

### Testing
Testing support is a base capability so that test-driven development is the low-friction path for both agents and humans.

Portico ships a testkit (separate from the production runtime) that provides:

- In-memory session fixtures that exercise the real effect-application path with no network and an in-memory store.
- Helpers to send input or invoke actions and assert on the current scene, rendered output, local state, persisted values, transitions, and broadcasts.
- Multi-session fixtures for testing spaces, presence, and shared state.
- Exposure of the same definition validation used at load time, so tests fail fast on broken references.
- Optional lightweight string snapshots for non-trivial scene output.

Scaffolds may emit a starter test alongside generated definitions. The testkit reuses the real runtime apply logic and Store interface so tests remain honest. Full end-to-end tests over a live SSH port are possible but are not the default path.

These capabilities stay inside the runtime and the identity layer. Payment providers, mail servers, search engines, media streaming, clustering, and unrestricted plugin execution remain outside the base framework and are reached only through explicit extension points when needed.

---

## 8. Package Layout and Key Types

Portico is implemented in Go. The initial package layout keeps the public surface small while concentrating strictness in the runtime and experience packages.

```
portico/
├── portico.go                 // public entry points, Experience builder, Serve
├── runtime/
│   ├── server.go              // SSH server setup
│   ├── session.go             // Session type and lifecycle
│   ├── dispatcher.go          // input → action routing
│   └── apply.go               // applies effects from actions
├── experience/
│   ├── experience.go          // Experience definition and builder
│   ├── scene.go
│   ├── action.go
│   ├── space.go
│   └── ids.go                 // SceneID, ActionID, SpaceID
├── identity/
│   └── identity.go            // Identity, Store interface
├── scaffold/
│   ├── guestbook.go
│   ├── chat.go
│   └── landing.go
├── testkit/
│   └── testkit.go             // in-memory session fixtures and assertions
└── internal/
    └── ...                    // shared unexported helpers
```

### Key types (illustrative)

```go
// experience/ids.go
type SceneID string
type ActionID string
type SpaceID string

// experience/experience.go
type Experience struct {
    scenes  map[SceneID]Scene
    actions map[ActionID]Action
    spaces  map[SpaceID]Space
}

type Builder struct{ /* ... */ }
func New() *Builder
func (b *Builder) Scene(id SceneID, s Scene) *Builder
func (b *Builder) Action(id ActionID, a Action) *Builder
func (b *Builder) Space(id SpaceID, s Space) *Builder
func (b *Builder) Build() (*Experience, error)

// experience/scene.go
type Scene struct {
    Render func(ctx SceneContext) string
    Inputs map[string]ActionID
}

type SceneContext interface {
    Identity() identity.Identity
    Local(key string) (any, bool)
    SetLocal(key string, value any)
    Transition(to SceneID) Effect
}

// experience/action.go
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
    current  experience.SceneID
    local    map[string]any
    spaces   map[experience.SpaceID]struct{}
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

```go
exp, err := portico.New().
    Scene("home", homeScene).
    Action("go-guestbook", goGuestbook).
    Scene("guestbook", guestbookScene).
    Build()

portico.Serve(":22", exp)
```

This structure keeps the happy path short while ensuring that core state remains unexported, mutation is centralized, and identifiers are distinct types.

---

## 9. Agent Interaction Model

Agents interact with Portico through three complementary mechanisms.

### 8.1 Declarative Definitions

The primary artifact is a structured definition of the experience (scenes, actions, spaces, identity configuration, and asset references). Agents create and edit these definitions. The runtime loads them and hot-reloads changes.

The definition format is designed to be easy for an agent to generate, validate, and incrementally modify. It is also human-readable so that directors can review and adjust it.

### 8.2 High-Level Operations

The framework exposes operations that agents can invoke:

- Scaffold a common pattern (guestbook, chat space, blog reader, multiplayer lobby, etc.).
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

## 10. Scaffolding and Recipes

Scaffolds are the preferred way to begin. Each scaffold produces a complete, minimal, working set of definitions for a common pattern. The resulting definitions are ordinary definitions that the agent can continue to edit.

Example scaffolds (illustrative, not exhaustive):

- **Guestbook** — A scene that displays recent signatures and an action that appends a new signature with the visitor’s identity.
- **Chat Space** — A space with shared message history, presence, and actions for sending messages and joining/leaving.
- **Blog Reader** — Scenes for listing posts and reading a single post, with assets holding the post content.
- **Simple Multiplayer Lobby** — A space that tracks participants and provides actions for signaling readiness or starting a shared activity.
- **Personal Landing** — A root scene that presents a menu of available destinations and basic identity information.

Scaffolds may be composed. An agent can scaffold a landing page, a guestbook, and a chat space, then wire them together by editing the resulting definitions.

Recipes are higher-level compositions of scaffolds plus recommended wiring. They exist to capture complete, opinionated starting points for common site types.

---

## 11. Runtime Behavior

The runtime is responsible for:

- Accepting SSH connections and establishing sessions.
- Deriving identity from presented keys.
- Loading the experience definition and hot-reloading it when it changes.
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

Agents are the primary authors. They work with declarative definitions of scenes, actions, spaces, identity, and assets. The runtime owns session lifecycle, rendering, input dispatch, presence, and persistence. Common patterns are available as scaffolds. Custom logic attaches at explicit extension points when the declarative path is insufficient.

The result is a system in which creating, iterating on, and maintaining an SSH site is predictable, inspectable, and fast, while remaining open to experiences that the original designers did not anticipate. The name Portico reflects its role as the structured entrance through which SSH sessions arrive and through which agents and humans build what lies beyond.
