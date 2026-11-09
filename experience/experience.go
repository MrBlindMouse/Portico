package experience

import (
	"fmt"
)

// Experience is a validated site definition.
type Experience struct {
	root            SceneID
	scenes          map[SceneID]Scene
	actions         map[ActionID]Action
	spaces          map[SpaceID]Space
	assets          map[string]string
	onSessionStart  []Effect
	onSessionEnd    []Effect
}

// OnSessionStart returns effects run when a session connects.
func (e *Experience) OnSessionStart() []Effect {
	if e == nil {
		return nil
	}
	return append([]Effect(nil), e.onSessionStart...)
}

// OnSessionEnd returns effects run when a session disconnects.
func (e *Experience) OnSessionEnd() []Effect {
	if e == nil {
		return nil
	}
	return append([]Effect(nil), e.onSessionEnd...)
}

// Root returns the entry scene.
func (e *Experience) Root() SceneID { return e.root }

// Scene returns a scene by id.
func (e *Experience) Scene(id SceneID) (Scene, bool) {
	s, ok := e.scenes[id]
	return s, ok
}

// Action returns an action by id.
func (e *Experience) Action(id ActionID) (Action, bool) {
	a, ok := e.actions[id]
	return a, ok
}

// Space returns a space definition by id.
func (e *Experience) Space(id SpaceID) (Space, bool) {
	s, ok := e.spaces[id]
	return s, ok
}

// Scenes returns all scene ids.
func (e *Experience) Scenes() []SceneID {
	out := make([]SceneID, 0, len(e.scenes))
	for id := range e.scenes {
		out = append(out, id)
	}
	return out
}

// Actions returns all action ids.
func (e *Experience) Actions() []ActionID {
	out := make([]ActionID, 0, len(e.actions))
	for id := range e.actions {
		out = append(out, id)
	}
	return out
}

// Spaces returns all space ids.
func (e *Experience) Spaces() []SpaceID {
	out := make([]SpaceID, 0, len(e.spaces))
	for id := range e.spaces {
		out = append(out, id)
	}
	return out
}

// Asset returns static asset content registered on the builder.
func (e *Experience) Asset(name string) (string, bool) {
	v, ok := e.assets[name]
	return v, ok
}

// CloneWithAssets returns a copy of e with assets replaced.
func (e *Experience) CloneWithAssets(assets map[string]string) *Experience {
	cpAssets := make(map[string]string, len(assets))
	for k, v := range assets {
		cpAssets[k] = v
	}
	cpScenes := make(map[SceneID]Scene, len(e.scenes))
	for k, v := range e.scenes {
		cpScenes[k] = v
	}
	cpActions := make(map[ActionID]Action, len(e.actions))
	for k, v := range e.actions {
		cpActions[k] = v
	}
	cpSpaces := make(map[SpaceID]Space, len(e.spaces))
	for k, v := range e.spaces {
		cpSpaces[k] = v
	}
	return &Experience{
		root:           e.root,
		scenes:         cpScenes,
		actions:        cpActions,
		spaces:         cpSpaces,
		assets:         cpAssets,
		onSessionStart: append([]Effect(nil), e.onSessionStart...),
		onSessionEnd:   append([]Effect(nil), e.onSessionEnd...),
	}
}

// Builder constructs an Experience with validation at Build().
type Builder struct {
	root            SceneID
	scenes          map[SceneID]Scene
	actions         map[ActionID]Action
	spaces          map[SpaceID]Space
	assets          map[string]string
	onSessionStart  []Effect
	onSessionEnd    []Effect
}

// NewBuilder starts an experience definition.
func NewBuilder() *Builder {
	return &Builder{
		scenes:  make(map[SceneID]Scene),
		actions: make(map[ActionID]Action),
		spaces:  make(map[SpaceID]Space),
		assets:  make(map[string]string),
	}
}

// Root sets the scene new sessions enter.
func (b *Builder) Root(id SceneID) *Builder {
	b.root = id
	return b
}

// Scene registers a scene.
func (b *Builder) Scene(id SceneID, s Scene) *Builder {
	b.scenes[id] = s
	return b
}

// Action registers an action.
func (b *Builder) Action(id ActionID, a Action) *Builder {
	b.actions[id] = a
	return b
}

// Space registers a space.
func (b *Builder) Space(id SpaceID, s Space) *Builder {
	b.spaces[id] = s
	return b
}

// Asset registers static text content by name.
func (b *Builder) Asset(name, content string) *Builder {
	b.assets[name] = content
	return b
}

// OnSessionStart registers effects applied when a session connects.
func (b *Builder) OnSessionStart(effects ...Effect) *Builder {
	b.onSessionStart = append(b.onSessionStart, effects...)
	return b
}

// OnSessionEnd registers effects applied when a session disconnects.
func (b *Builder) OnSessionEnd(effects ...Effect) *Builder {
	b.onSessionEnd = append(b.onSessionEnd, effects...)
	return b
}

// Merge incorporates another builder's registrations (later wins on conflict).
func (b *Builder) Merge(other *Builder) *Builder {
	if other == nil {
		return b
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
		b.onSessionStart = append([]Effect(nil), other.onSessionStart...)
	}
	if len(other.onSessionEnd) > 0 {
		b.onSessionEnd = append([]Effect(nil), other.onSessionEnd...)
	}
	return b
}

// Build validates and returns the experience.
func (b *Builder) Build() (*Experience, error) {
	if b.root == "" {
		return nil, fmt.Errorf("portico: root scene is required")
	}
	if err := b.root.validate(); err != nil {
		return nil, err
	}
	if _, ok := b.scenes[b.root]; !ok {
		return nil, fmt.Errorf("portico: undefined scene %q", b.root)
	}
	for id, sc := range b.scenes {
		if err := id.validate(); err != nil {
			return nil, err
		}
		if sc.Render == nil {
			return nil, fmt.Errorf("portico: scene %q has nil Render", id)
		}
		for input, aid := range sc.Inputs {
			if err := aid.validate(); err != nil {
				return nil, fmt.Errorf("portico: scene %q input %q: %w", id, input, err)
			}
			if _, ok := b.actions[aid]; !ok {
				return nil, fmt.Errorf("portico: undefined action %q referenced from scene %q input %q", aid, id, input)
			}
		}
	}
	for id, act := range b.actions {
		if err := id.validate(); err != nil {
			return nil, err
		}
		if act.Run == nil {
			return nil, fmt.Errorf("portico: action %q has nil Run", id)
		}
	}
	for id := range b.spaces {
		if err := id.validate(); err != nil {
			return nil, err
		}
	}
	if err := b.validateEffectSceneRefs(); err != nil {
		return nil, err
	}
	cpScenes := make(map[SceneID]Scene, len(b.scenes))
	for k, v := range b.scenes {
		cpScenes[k] = v
	}
	cpActions := make(map[ActionID]Action, len(b.actions))
	for k, v := range b.actions {
		cpActions[k] = v
	}
	cpSpaces := make(map[SpaceID]Space, len(b.spaces))
	for k, v := range b.spaces {
		cpSpaces[k] = v
	}
	cpAssets := make(map[string]string, len(b.assets))
	for k, v := range b.assets {
		cpAssets[k] = v
	}
	return &Experience{
		root:           b.root,
		scenes:         cpScenes,
		actions:        cpActions,
		spaces:         cpSpaces,
		assets:         cpAssets,
		onSessionStart: append([]Effect(nil), b.onSessionStart...),
		onSessionEnd:   append([]Effect(nil), b.onSessionEnd...),
	}, nil
}

func (b *Builder) validateEffectSceneRefs() error {
	checkEffects := func(effects []Effect) error {
		for _, eff := range effects {
			switch v := eff.(type) {
			case Transition:
				if _, ok := b.scenes[v.To]; !ok {
					return fmt.Errorf("portico: undefined scene %q in transition effect", v.To)
				}
			case JoinSpace, LeaveSpace, SpaceSet, SpaceCAS, Broadcast, AppendSpaceList:
				var sid SpaceID
				switch x := eff.(type) {
				case JoinSpace:
					sid = x.Space
				case LeaveSpace:
					sid = x.Space
				case SpaceSet:
					sid = x.Space
				case SpaceCAS:
					sid = x.Space
				case Broadcast:
					sid = x.Space
				case AppendSpaceList:
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
