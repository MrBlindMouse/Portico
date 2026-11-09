// Package experience defines declarative scenes, actions, spaces, and the
// Experience builder for site authors.
//
// Invariants: Scene.Render uses RenderContext (read-only). Action.Run returns []Effect;
// the runtime apply step is the only mutator for session and space state. Build() validates
// references before Serve. See docs/site-agents.md in the Portico module at your go.mod version.
//
// Scenes render with RenderContext (including Size and Theme). Actions use
// ActionContext (InputLine, space, persist effects). MoveList stores a selection
// the same way forms store a step: the action returns the effect, RenderList only reads.
// A caret rectangle is render.Screen.Input then Screen.Show, which calls PlaceInput.
package experience
