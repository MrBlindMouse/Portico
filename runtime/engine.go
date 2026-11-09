package runtime

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
)

// Engine runs the session loop against a Terminal.
type Engine struct {
	mu        sync.RWMutex
	exp       *experience.Experience
	spaces    *spaceRegistry
	applier   *Applier
	sessions  map[string]*Session
	nextID    atomic.Uint64
	onAction  func()
	limits    *rateLimiter
	log       *slog.Logger
}

// EnableRateLimit enables per-session action throttling.
func (e *Engine) EnableRateLimit(rate, burst float64) {
	e.limits = newRateLimiter(rate, burst)
}

// NewEngine creates a runtime engine.
func NewEngine(exp *experience.Experience, applier *Applier) *Engine {
	if applier == nil {
		applier = NewApplier(nil, identity.NewMemoryStore())
	}
	return &Engine{
		exp:      exp,
		spaces:   applier.spaces,
		applier:  applier,
		sessions: make(map[string]*Session),
	}
}

// SetExperience hot-swaps the experience definition after validation.
func (e *Engine) SetExperience(exp *experience.Experience) error {
	if exp == nil {
		return fmt.Errorf("portico: experience is required")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, sess := range e.sessions {
		if _, ok := exp.Scene(sess.CurrentScene()); !ok {
			return fmt.Errorf("portico: session %s in scene %q not in new experience", sess.ID(), sess.CurrentScene())
		}
	}
	e.exp = exp
	return nil
}

// UpdateAssets replaces asset content on the live experience.
func (e *Engine) UpdateAssets(assets map[string]string) error {
	e.mu.RLock()
	cur := e.exp
	e.mu.RUnlock()
	return e.SetExperience(cur.CloneWithAssets(assets))
}

// Experience returns the current definition.
func (e *Engine) Experience() *experience.Experience {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.exp
}

// Applier returns the identity/space applier.
func (e *Engine) Applier() *Applier { return e.applier }

// SpaceSnapshots returns read-only space summaries for inspection.
func (e *Engine) SpaceSnapshots() map[string]SpaceSnapshot {
	return e.spaces.snapshot()
}

// SetOnAction hooks action execution for metrics.
func (e *Engine) SetOnAction(fn func()) { e.onAction = fn }

// SetLogger enables structured logs for actions and render errors.
func (e *Engine) SetLogger(l *slog.Logger) {
	if l != nil {
		e.log = l
	}
}

// RegisterSession adds a session to the engine registry.
func (e *Engine) RegisterSession(sess *Session) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sessions[sess.ID()] = sess
}

// UnregisterSession removes a session.
func (e *Engine) UnregisterSession(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.sessions, id)
}

// Sessions returns all live sessions.
func (e *Engine) Sessions() []*Session {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]*Session, 0, len(e.sessions))
	for _, s := range e.sessions {
		out = append(out, s)
	}
	return out
}

// NewSessionID allocates a unique session id.
func (e *Engine) NewSessionID() string {
	return fmt.Sprintf("sess-%d", e.nextID.Add(1))
}

// RenderText returns the current scene output without writing a terminal.
func (e *Engine) RenderText(sess *Session) (string, error) {
	out, err := e.renderScene(sess)
	if err != nil {
		return "", err
	}
	sess.setLastRender(out)
	return out, nil
}

// Render writes the current scene to the terminal.
func (e *Engine) Render(sess *Session, term Terminal) error {
	out, err := e.RenderText(sess)
	if err != nil {
		return err
	}
	return term.WriteFrame(out)
}

func (e *Engine) renderScene(sess *Session) (string, error) {
	e.mu.RLock()
	exp := e.exp
	e.mu.RUnlock()
	sc, ok := exp.Scene(sess.CurrentScene())
	if !ok {
		return "", fmt.Errorf("portico: session scene %q undefined", sess.CurrentScene())
	}
	ctx := newSceneContext(sess, exp, e.spaces, e.applier.Store())
	return sc.Render(ctx), nil
}

// HandleInput resolves and runs an action for one input token.
func (e *Engine) HandleInput(sess *Session, input string) error {
	e.mu.RLock()
	exp := e.exp
	e.mu.RUnlock()
	if e.limits != nil {
		if err := e.limits.allow(sess.ID()); err != nil {
			return err
		}
	}
	aid, err := ResolveInput(exp, sess.CurrentScene(), input)
	if err != nil {
		if e.log != nil {
			e.log.Warn("dispatch_failed", "session", sess.ID(), "input", input, "error", err)
		}
		return err
	}
	act, ok := exp.Action(aid)
	if !ok {
		return fmt.Errorf("portico: undefined action %q", aid)
	}
	ctx := newActionContext(sess, exp, e.spaces, e.applier.Store(), input)
	effects := act.Run(ctx)
	res, err := e.applier.Apply(sess, exp, effects)
	if err != nil {
		if e.log != nil {
			e.log.Error("apply_failed", "session", sess.ID(), "action", aid, "error", err)
		}
		return err
	}
	if e.onAction != nil {
		e.onAction()
	}
	if e.log != nil {
		e.log.Info("action", "session", sess.ID(), "action", aid, "scene", sess.CurrentScene(), "transitioned", res.Transitioned)
	}
	if len(res.Broadcast) > 0 {
		for _, space := range res.Broadcast {
			for _, member := range e.spaces.members(space) {
				_ = e.Render(member, memberTerminal(member))
			}
		}
	}
	return nil
}

func notifyActivity(ch chan<- struct{}) {
	if ch == nil {
		return
	}
	select {
	case ch <- struct{}{}:
	default:
	}
}

// memberTerminal is set per session when attached to SSH; engine tests set it on session via wrapper.
var sessionTerminals sync.Map

func setSessionTerminal(sess *Session, t Terminal) {
	sessionTerminals.Store(sess.ID(), t)
}

func memberTerminal(sess *Session) Terminal {
	if v, ok := sessionTerminals.Load(sess.ID()); ok {
		return v.(Terminal)
	}
	return noopTerminal{}
}

type noopTerminal struct{}

func (noopTerminal) WriteFrame(string) error                      { return nil }
func (noopTerminal) ReadInput() (string, error)                   { return "", fmt.Errorf("no terminal") }
func (noopTerminal) OnResize(func(int, int))                      {}
func (noopTerminal) Close() error                                 { return nil }

// RunLoop drives read input → action → render until terminal closes.
// activity, when non-nil, receives a signal on each non-empty input (for idle timeouts).
func (e *Engine) RunLoop(sess *Session, term Terminal, activity chan<- struct{}) error {
	setSessionTerminal(sess, term)
	defer sessionTerminals.Delete(sess.ID())
	if err := e.Render(sess, term); err != nil {
		return err
	}
	for {
		in, err := term.ReadInput()
		if err != nil {
			return err
		}
		if in == "" {
			continue
		}
		notifyActivity(activity)
		if err := e.HandleInput(sess, in); err != nil {
			if e.log != nil {
				e.log.Warn("input_error", "session", sess.ID(), "input", in, "error", err)
			}
			_ = term.WriteFrame("Error: " + err.Error() + "\n")
		}
		if err := e.Render(sess, term); err != nil {
			return err
		}
	}
}
