package components

import (
	"unicode/utf8"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/render"
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
	// Size is the terminal window in columns and rows.
	// A session with no reported size returns 80 by 24.
	Size() (width, height int)
	// Theme is the site theme for this session.
	// The runtime uses render.Default unless Options.Theme or NO_COLOR selects another.
	// Roles produce terminal strings. Fit text first, then apply roles. On render.Screen,
	// keep grid draws plain; use Screen.DrawList or Screen.MarkRole (Show applies the role on those cells).
	Theme() render.Theme
	// PlaceInput records the single caret rectangle for this frame.
	// It does not change session state. A second call replaces the first.
	// A scene with Keys set fails when it renders if a caret was placed.
	// Leave it unset and the caret is the full row under the frame.
	PlaceInput(in render.Input)
	// RecentCommands returns the last n successful lines for this session, oldest first.
	// n <= 0 returns the whole list. The list is empty after a scene change unless
	// the host set TranscriptKeep. TranscriptOff leaves it empty.
	RecentCommands(n int) []string
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
