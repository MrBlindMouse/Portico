package runtime

import (
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/render"
)

type sceneCtx struct {
	sess  *Session
	exp   *experience.Experience
	sp    *spaceRegistry
	store identity.Store
	theme render.Theme
	place *inputPlace
}

// inputPlace is the caret chosen while Render runs. Nil means the call is not a paint.
type inputPlace struct {
	in render.Input
	ok bool
}

func (c sceneCtx) Identity() identity.Identity { return c.sess.Identity() }

func (c sceneCtx) Local(key string) (any, bool) {
	c.sess.mu.RLock()
	defer c.sess.mu.RUnlock()
	v, ok := c.sess.local[key]
	return v, ok
}

func (c sceneCtx) SetLocal(key string, value any) experience.Effect {
	return experience.SetLocalEffect(key, value)
}

func (c sceneCtx) Transition(to experience.SceneID) experience.Effect {
	return experience.TransitionEffect(to)
}

func (c sceneCtx) Asset(name string) (string, bool) {
	return c.exp.Asset(name)
}

func (c sceneCtx) Assets() map[string]string {
	return c.exp.Assets()
}

func (c sceneCtx) Persisted(key string) (any, bool) {
	if c.store == nil {
		return nil, false
	}
	v, ok, err := c.store.Get(c.sess.Identity(), key)
	if err != nil {
		return nil, false
	}
	return v, ok
}

func (c sceneCtx) Members(space experience.SpaceID) []identity.Identity {
	return c.sp.memberIdentities(space)
}

func (c sceneCtx) SpaceLocal(space experience.SpaceID, key string) (any, bool) {
	return c.sp.getValue(space, key)
}

func (c sceneCtx) SpaceList(space experience.SpaceID, key string) []any {
	return c.sp.list(space, key)
}

func (c sceneCtx) Size() (int, int) { return c.sess.size() }

func (c sceneCtx) Theme() render.Theme {
	if c.theme.Bold == nil {
		return render.Default
	}
	return c.theme
}

func (c sceneCtx) PlaceInput(in render.Input) {
	if c.place == nil {
		return
	}
	c.place.in = in
	c.place.ok = true
}

func (c sceneCtx) RecentCommands(n int) []string { return c.sess.RecentCommands(n) }

type actionCtx struct {
	sceneCtx
	inputLine string
}

func (c actionCtx) InputLine() string { return c.inputLine }

func (c actionCtx) Space() experience.SpaceContext { return spaceCtx{c.sess, c.sp} }

func (c actionCtx) Persist(key string, value any) experience.Effect {
	return experience.PersistEffect(key, value)
}

type spaceCtx struct {
	sess *Session
	sp   *spaceRegistry
}

func (s spaceCtx) In(space experience.SpaceID) bool { return s.sess.InSpace(space) }

func (s spaceCtx) Members(space experience.SpaceID) []identity.Identity {
	return s.sp.memberIdentities(space)
}

func (s spaceCtx) Local(space experience.SpaceID, key string) (any, bool) {
	return s.sp.getValue(space, key)
}

func (s spaceCtx) List(space experience.SpaceID, key string) []any {
	return s.sp.list(space, key)
}

func newActionContext(sess *Session, exp *experience.Experience, sp *spaceRegistry, store identity.Store, inputLine string, theme render.Theme) experience.ActionContext {
	return actionCtx{
		sceneCtx:  sceneCtx{sess: sess, exp: exp, sp: sp, store: store, theme: theme},
		inputLine: inputLine,
	}
}

func newSceneContext(sess *Session, exp *experience.Experience, sp *spaceRegistry, store identity.Store, theme render.Theme, place *inputPlace) experience.SceneContext {
	return sceneCtx{sess: sess, exp: exp, sp: sp, store: store, theme: theme, place: place}
}
