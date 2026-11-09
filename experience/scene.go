package experience

import (
	"unicode/utf8"

	"github.com/MrBlindMouse/Portico/identity"
)

// Scene describes a renderable view and input bindings.
type Scene struct {
	Render func(ctx RenderContext) string
	Inputs map[string]ActionID
	// Keys dispatches each bound key immediately. Input names must be one rune
	// or up, down, left, right. The line prompt is not used.
	Keys bool
	// ValidateLocal optionally rejects session-local writes in apply.
	ValidateLocal func(key string, value any) error
}

// SingleKeyInput reports whether name is legal on a Scene with Keys set.
func SingleKeyInput(name string) bool {
	switch name {
	case "up", "down", "left", "right":
		return true
	default:
		return utf8.RuneCountInString(name) == 1 && name != ""
	}
}

// RenderContext is read-only context for scene rendering.
type RenderContext interface {
	Identity() identity.Identity
	Local(key string) (any, bool)
	Asset(name string) (string, bool)
	Assets() map[string]string
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

// SessionView is the session surface available to action hooks.
type SessionView interface {
	ID() string
	Identity() identity.Identity
	CurrentScene() SceneID
}
