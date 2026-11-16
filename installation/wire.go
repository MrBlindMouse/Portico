package installation

import (
	"fmt"
	"sort"

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

// Wire returns a sketch of this installation.
// Render, Run, space hooks, and session hooks are omitted; the wire format has no place for them.
// A nil installation returns the zero Wire.
func (e *Installation) Wire() Wire {
	if e == nil {
		return Wire{}
	}
	w := Wire{
		Root:    string(e.root),
		Scenes:  make(map[string]WireScene, len(e.scenes)),
		Actions: make([]string, 0, len(e.actions)),
		Spaces:  make([]string, 0, len(e.spaces)),
	}
	for _, id := range e.Scenes() {
		sc, _ := e.Scene(id)
		inputs := make(map[string]string, len(sc.Inputs))
		for in, act := range sc.Inputs {
			inputs[in] = string(act)
		}
		w.Scenes[string(id)] = WireScene{Inputs: inputs}
	}
	for _, id := range e.Actions() {
		w.Actions = append(w.Actions, string(id))
	}
	for _, id := range e.Spaces() {
		w.Spaces = append(w.Spaces, string(id))
	}
	if names := assetNames(e.assets); len(names) > 0 {
		w.Assets = names
	}
	return w
}

func assetNames(assets map[string]string) []string {
	if len(assets) == 0 {
		return nil
	}
	names := make([]string, 0, len(assets))
	for name := range assets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
