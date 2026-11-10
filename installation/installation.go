package installation

import (
	"fmt"
	"github.com/MrBlindMouse/Portico/components"
	"sort"
)

// Installation is a validated site definition.
type Installation struct {
	root           components.SceneID
	scenes         map[components.SceneID]components.Scene
	actions        map[components.ActionID]components.Action
	spaces         map[components.SpaceID]components.Space
	assets         map[string]string
	onSessionStart []components.Effect
	onSessionEnd   []components.Effect
}

// OnSessionStart returns effects run when a session connects.
func (e *Installation) OnSessionStart() []components.Effect {
	if e == nil {
		return nil
	}
	return append([]components.Effect(nil), e.onSessionStart...)
}

// OnSessionEnd returns effects run when a session disconnects.
func (e *Installation) OnSessionEnd() []components.Effect {
	if e == nil {
		return nil
	}
	return append([]components.Effect(nil), e.onSessionEnd...)
}

// Root returns the entry scene.
func (e *Installation) Root() components.SceneID { return e.root }

// Scene returns a copy of a scene by id. The Inputs map is not shared with the installation.
func (e *Installation) Scene(id components.SceneID) (components.Scene, bool) {
	s, ok := e.scenes[id]
	if !ok {
		return components.Scene{}, false
	}
	return cloneScene(s), true
}

// Action returns an action by id.
func (e *Installation) Action(id components.ActionID) (components.Action, bool) {
	a, ok := e.actions[id]
	return a, ok
}

// Space returns a copy of a space definition by id. Hook slices are not shared.
func (e *Installation) Space(id components.SpaceID) (components.Space, bool) {
	s, ok := e.spaces[id]
	if !ok {
		return components.Space{}, false
	}
	return cloneSpace(s), true
}

// Scenes returns all scene ids in lexical order.
func (e *Installation) Scenes() []components.SceneID {
	out := make([]components.SceneID, 0, len(e.scenes))
	for id := range e.scenes {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Actions returns all action ids in lexical order.
func (e *Installation) Actions() []components.ActionID {
	out := make([]components.ActionID, 0, len(e.actions))
	for id := range e.actions {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Spaces returns all space ids in lexical order.
func (e *Installation) Spaces() []components.SpaceID {
	out := make([]components.SpaceID, 0, len(e.spaces))
	for id := range e.spaces {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Asset returns static asset content registered on the builder.
func (e *Installation) Asset(name string) (string, bool) {
	v, ok := e.assets[name]
	return v, ok
}

// Assets returns a copy of the named asset contents.
func (e *Installation) Assets() map[string]string {
	if e == nil {
		return nil
	}
	out := make(map[string]string, len(e.assets))
	for k, v := range e.assets {
		out[k] = v
	}
	return out
}

// CloneWithAssets returns a copy of e with assets replaced.
func (e *Installation) CloneWithAssets(assets map[string]string) *Installation {
	if e == nil {
		return nil
	}
	cpAssets := make(map[string]string, len(assets))
	for k, v := range assets {
		cpAssets[k] = v
	}
	cpScenes := make(map[components.SceneID]components.Scene, len(e.scenes))
	for k, v := range e.scenes {
		cpScenes[k] = cloneScene(v)
	}
	cpActions := make(map[components.ActionID]components.Action, len(e.actions))
	for k, v := range e.actions {
		cpActions[k] = v
	}
	cpSpaces := make(map[components.SpaceID]components.Space, len(e.spaces))
	for k, v := range e.spaces {
		cpSpaces[k] = cloneSpace(v)
	}
	return &Installation{
		root:           e.root,
		scenes:         cpScenes,
		actions:        cpActions,
		spaces:         cpSpaces,
		assets:         cpAssets,
		onSessionStart: append([]components.Effect(nil), e.onSessionStart...),
		onSessionEnd:   append([]components.Effect(nil), e.onSessionEnd...),
	}
}

// Builder constructs an Installation with validation at Build().
type Builder struct {
	root           components.SceneID
	scenes         map[components.SceneID]components.Scene
	actions        map[components.ActionID]components.Action
	spaces         map[components.SpaceID]components.Space
	assets         map[string]string
	onSessionStart []components.Effect
	onSessionEnd   []components.Effect
	err            error
}

// NewBuilder starts an installation definition.
func NewBuilder() *Builder {
	return &Builder{
		scenes:  make(map[components.SceneID]components.Scene),
		actions: make(map[components.ActionID]components.Action),
		spaces:  make(map[components.SpaceID]components.Space),
		assets:  make(map[string]string),
	}
}

// Root sets the scene new sessions enter.
func (b *Builder) Root(id components.SceneID) *Builder {
	b.root = id
	return b
}

// Scene registers a scene.
func (b *Builder) Scene(id components.SceneID, s components.Scene) *Builder {
	b.scenes[id] = s
	return b
}

// Action registers an action.
func (b *Builder) Action(id components.ActionID, a components.Action) *Builder {
	b.actions[id] = a
	return b
}

// BindInput adds a command key on an existing scene. A missing scene fails Build.
func (b *Builder) BindInput(scene components.SceneID, input string, action components.ActionID) *Builder {
	sc, ok := b.scenes[scene]
	if !ok {
		if b.err == nil {
			b.err = fmt.Errorf("portico: bind input scene %q not registered", scene)
		}
		return b
	}
	if sc.Inputs == nil {
		sc.Inputs = make(map[string]components.ActionID)
	}
	sc.Inputs[input] = action
	b.scenes[scene] = sc
	return b
}

// Space registers a space.
func (b *Builder) Space(id components.SpaceID, s components.Space) *Builder {
	b.spaces[id] = s
	return b
}

// Asset registers static text content by name.
func (b *Builder) Asset(name, content string) *Builder {
	b.assets[name] = content
	return b
}

// OnSessionStart registers effects applied when a session connects.
func (b *Builder) OnSessionStart(effects ...components.Effect) *Builder {
	b.onSessionStart = append(b.onSessionStart, effects...)
	return b
}

// OnSessionEnd registers effects applied when a session disconnects.
func (b *Builder) OnSessionEnd(effects ...components.Effect) *Builder {
	b.onSessionEnd = append(b.onSessionEnd, effects...)
	return b
}

// Merge incorporates another builder's registrations (later wins on id conflict).
// Session hooks append. Build snapshots scene inputs and space hooks so later edits do not change a built installation.
func (b *Builder) Merge(other *Builder) *Builder {
	if other == nil {
		return b
	}
	if b.err == nil {
		b.err = other.err
	}
	if other.root != "" {
		b.root = other.root
	}
	for id, s := range other.scenes {
		b.scenes[id] = s
	}
	for id, a := range other.actions {
		b.actions[id] = a
	}
	for id, s := range other.spaces {
		b.spaces[id] = s
	}
	for name, c := range other.assets {
		b.assets[name] = c
	}
	if len(other.onSessionStart) > 0 {
		b.onSessionStart = append(append([]components.Effect(nil), b.onSessionStart...), other.onSessionStart...)
	}
	if len(other.onSessionEnd) > 0 {
		b.onSessionEnd = append(append([]components.Effect(nil), b.onSessionEnd...), other.onSessionEnd...)
	}
	return b
}

// Build validates and returns the installation.
func (b *Builder) Build() (*Installation, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.root == "" {
		return nil, fmt.Errorf("portico: root scene is required")
	}
	if err := b.root.Validate(); err != nil {
		return nil, err
	}
	if _, ok := b.scenes[b.root]; !ok {
		return nil, fmt.Errorf("portico: undefined scene %q", b.root)
	}
	for id, sc := range b.scenes {
		if err := id.Validate(); err != nil {
			return nil, err
		}
		if sc.Render == nil {
			return nil, fmt.Errorf("portico: scene %q has nil Render", id)
		}
		if sc.Keys && len(sc.Inputs) == 0 {
			return nil, fmt.Errorf("portico: scene %q has Keys set and no inputs", id)
		}
		for input, aid := range sc.Inputs {
			if err := aid.Validate(); err != nil {
				return nil, fmt.Errorf("portico: scene %q input %q: %w", id, input, err)
			}
			if sc.Keys && !components.SingleKeyInput(input) {
				return nil, fmt.Errorf("portico: scene %q key %q must be one character or up/down/left/right", id, input)
			}
			if _, ok := b.actions[aid]; !ok {
				return nil, fmt.Errorf("portico: undefined action %q referenced from scene %q input %q", aid, id, input)
			}
		}
	}
	for id, act := range b.actions {
		if err := id.Validate(); err != nil {
			return nil, err
		}
		if act.Run == nil {
			return nil, fmt.Errorf("portico: action %q has nil Run", id)
		}
	}
	for id := range b.spaces {
		if err := id.Validate(); err != nil {
			return nil, err
		}
	}
	if err := b.validateEffectSceneRefs(); err != nil {
		return nil, err
	}
	cpScenes := make(map[components.SceneID]components.Scene, len(b.scenes))
	for k, v := range b.scenes {
		cpScenes[k] = cloneScene(v)
	}
	cpActions := make(map[components.ActionID]components.Action, len(b.actions))
	for k, v := range b.actions {
		cpActions[k] = v
	}
	cpSpaces := make(map[components.SpaceID]components.Space, len(b.spaces))
	for k, v := range b.spaces {
		cpSpaces[k] = cloneSpace(v)
	}
	cpAssets := make(map[string]string, len(b.assets))
	for k, v := range b.assets {
		cpAssets[k] = v
	}
	return &Installation{
		root:           b.root,
		scenes:         cpScenes,
		actions:        cpActions,
		spaces:         cpSpaces,
		assets:         cpAssets,
		onSessionStart: append([]components.Effect(nil), b.onSessionStart...),
		onSessionEnd:   append([]components.Effect(nil), b.onSessionEnd...),
	}, nil
}

func (b *Builder) validateEffectSceneRefs() error {
	checkEffects := func(effects []components.Effect) error {
		for _, eff := range effects {
			switch v := eff.(type) {
			case components.Transition:
				if _, ok := b.scenes[v.To]; !ok {
					return fmt.Errorf("portico: undefined scene %q in transition effect", v.To)
				}
			case components.JoinSpace, components.LeaveSpace, components.SpaceSet, components.SpaceCAS, components.Broadcast, components.AppendSpaceList:
				var sid components.SpaceID
				switch x := eff.(type) {
				case components.JoinSpace:
					sid = x.Space
				case components.LeaveSpace:
					sid = x.Space
				case components.SpaceSet:
					sid = x.Space
				case components.SpaceCAS:
					sid = x.Space
				case components.Broadcast:
					sid = x.Space
				case components.AppendSpaceList:
					sid = x.Space
				}
				if _, ok := b.spaces[sid]; !ok {
					return fmt.Errorf("portico: undefined space %q in effect", sid)
				}
			}
		}
		return nil
	}
	for _, sp := range b.spaces {
		if err := checkEffects(sp.OnJoin); err != nil {
			return err
		}
		if err := checkEffects(sp.OnLeave); err != nil {
			return err
		}
	}
	if err := checkEffects(b.onSessionStart); err != nil {
		return err
	}
	return checkEffects(b.onSessionEnd)
}

func cloneScene(s components.Scene) components.Scene {
	if s.Inputs != nil {
		in := make(map[string]components.ActionID, len(s.Inputs))
		for k, v := range s.Inputs {
			in[k] = v
		}
		s.Inputs = in
	}
	return s
}

func cloneSpace(s components.Space) components.Space {
	s.OnJoin = append([]components.Effect(nil), s.OnJoin...)
	s.OnLeave = append([]components.Effect(nil), s.OnLeave...)
	return s
}
