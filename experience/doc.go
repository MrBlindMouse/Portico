// Package experience defines declarative scenes, actions, spaces, and the
// Experience builder for site authors.
//
// Invariants: Scene.Render uses RenderContext (read-only). Action.Run returns []Effect;
// the runtime apply step is the only mutator for session and space state. Build() validates
// references before Serve. See docs/site-agents.md in the Portico module at your go.mod version.
//
// Scenes render with RenderContext; actions use ActionContext (InputLine, space, persist effects).
package experience
