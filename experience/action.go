package experience

import "github.com/MrBlindMouse/Portico/identity"

// Action is a named operation invoked from a scene.
type Action struct {
	Run func(ctx ActionContext) []Effect
}

// ActionContext extends SceneContext with space and persistence effects.
type ActionContext interface {
	SceneContext
	InputLine() string
	Space() SpaceContext
	Persist(key string, value any) Effect
}

// SpaceContext exposes read-only space state for the current session.
type SpaceContext interface {
	In(space SpaceID) bool
	Members(space SpaceID) []identity.Identity
	Local(space SpaceID, key string) (any, bool)
	List(space SpaceID, key string) []any
}
