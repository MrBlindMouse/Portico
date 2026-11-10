package installation

import (
	"fmt"

	"github.com/MrBlindMouse/Portico/components"
)

// Wire is a minimal serializable installation sketch for inspect validation (not a full site definition).
type Wire struct {
	Root    string               `json:"root"`
	Scenes  map[string]WireScene `json:"scenes"`
	Actions []string             `json:"actions"`
	Spaces  []string             `json:"spaces"`
	Assets  []string             `json:"assets,omitempty"`
}

// WireScene lists input bindings by action name.
type WireScene struct {
	Inputs map[string]string `json:"inputs"`
}

// ValidateWire checks wiring references without requiring Render/Run funcs.
func ValidateWire(w Wire) error {
	b := NewBuilder()
	if w.Root != "" {
		b.Root(components.SceneID(w.Root))
	}
	for name := range w.Scenes {
		b.Scene(components.SceneID(name), components.Scene{
			Render: func(components.RenderContext) string { return "" },
			Inputs: map[string]components.ActionID{},
		})
	}
	for _, name := range w.Actions {
		b.Action(components.ActionID(name), components.Action{Run: func(components.ActionContext) []components.Effect { return nil }})
	}
	for name, sc := range w.Scenes {
		sid := components.SceneID(name)
		cur, ok := b.scenes[sid]
		if !ok {
			return fmt.Errorf("portico: wire scene %q not registered", name)
		}
		inputs := make(map[string]components.ActionID, len(sc.Inputs))
		for in, act := range sc.Inputs {
			inputs[in] = components.ActionID(act)
		}
		cur.Inputs = inputs
		b.scenes[sid] = cur
	}
	for _, name := range w.Spaces {
		b.Space(components.SpaceID(name), components.Space{})
	}
	_, err := b.Build()
	return err
}
