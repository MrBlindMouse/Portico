package experience

import "github.com/MrBlindMouse/Portico/identity"

// Scene describes a renderable view and input bindings.
type Scene struct {
	Render func(ctx RenderContext) string
	Inputs map[string]ActionID
}

// RenderContext is read-only context for scene rendering.
type RenderContext interface {
	Identity() identity.Identity
	Local(key string) (any, bool)
	Asset(name string) (string, bool)
	Persisted(key string) (any, bool)
	Members(space SpaceID) []identity.Identity
	SpaceLocal(space SpaceID, key string) (any, bool)
	SpaceList(space SpaceID, key string) []any
}

// SceneContext adds effect descriptions for actions (not applied during Render).
type SceneContext interface {
	RenderContext
	SetLocal(key string, value any) Effect
	Transition(to SceneID) Effect
}
