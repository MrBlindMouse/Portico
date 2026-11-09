package runtime

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
)

// Engine runs the session loop against a Terminal.
type Engine struct {
	mu             sync.RWMutex
	exp            *experience.Experience
	spaces         *spaceRegistry
	applier        *Applier
	sessions       map[string]*Session
	nextID         atomic.Uint64
	onAction       func()
	hooks          ActionHooks
	limits         *rateLimiter
	log            *slog.Logger
	reloadFailures atomic.Uint64
}

// ActionHooks optionally intercept action dispatch.
type ActionHooks struct {
	Before func(sess *Session, action experience.ActionID, line string) (string, error)
	After  func(sess *Session, action experience.ActionID, result ApplyResult, err error)
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
		for _, sp := range sess.Spaces() {
			if _, ok := exp.Space(sp); !ok {
				return fmt.Errorf("portico: session %s in space %q not in new experience", sess.ID(), sp)
			}
		}
	}
	e.exp = exp
	return nil
}

// RedrawSessions paints the current scene for every session that has a terminal.
func (e *Engine) RedrawSessions() {
	for _, sess := range e.Sessions() {
		term := memberTerminal(sess)
		if _, skip := term.(noopTerminal); skip {
			continue
		}
		_ = e.Render(sess, term)
		if ed, ok := term.(LineEditor); ok {
			ed.RestoreInputLine()
		}
	}
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

// SetActionHooks configures before/after action callbacks.
func (e *Engine) SetActionHooks(h ActionHooks) { e.hooks = h }

// ReloadFailures returns failed hot-reload attempts.
func (e *Engine) ReloadFailures() uint64 { return e.reloadFailures.Load() }

// NoteReloadFailure increments reload failure metrics.
func (e *Engine) NoteReloadFailure() { e.reloadFailures.Add(1) }

// SetLogger enables structured logs for actions and render errors.
func (e *Engine) SetLogger(l *slog.Logger) {
	if l != nil {
		e.log = l
		if e.applier != nil {
			e.applier.SetLogger(l)
		}
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
	line := input
	aid, err := ResolveInput(exp, sess.CurrentScene(), line)
	if err != nil {
		if e.log != nil {
			e.log.Warn("dispatch_failed", "session", sess.ID(), "scene", sess.CurrentScene(), "input_len", len(line), "error", err)
			e.log.Debug("dispatch_failed_input", "session", sess.ID(), "input_preview", logInputPreview(line))
		}
		return err
	}
	if e.limits != nil {
		if err := e.limits.allow(sess.ID()); err != nil {
			return err
		}
	}
	act, ok := exp.Action(aid)
	if !ok {
		return fmt.Errorf("portico: undefined action %q", aid)
	}
	if e.hooks.Before != nil {
		var err error
		line, err = e.hooks.Before(sess, aid, line)
		if err != nil {
			return err
		}
	}
	applyStart := time.Now()
	ctx := newActionContext(sess, exp, e.spaces, e.applier.Store(), line)
	effects := act.Run(ctx)
	res, err := e.applier.Apply(sess, exp, effects)
	applyMs := time.Since(applyStart).Milliseconds()
	if e.hooks.After != nil {
		e.hooks.After(sess, aid, res, err)
	}
	if err != nil {
		if e.log != nil {
			e.log.Error("apply_failed", "session", sess.ID(), "action", aid, "error", err)
		}
		return err
	}
	if e.onAction != nil {
		e.onAction()
	}
	var broadcastMs int64
	broadcastN := 0
	if len(res.Broadcast) > 0 {
		bStart := time.Now()
		for _, space := range res.Broadcast {
			members := e.spaces.members(space)
			broadcastN += len(members)
			for _, member := range members {
				_ = e.Render(member, memberTerminal(member))
			}
		}
		broadcastMs = time.Since(bStart).Milliseconds()
	}
	if e.log != nil {
		e.log.Info("action",
			"session", sess.ID(),
			"action", aid,
			"scene", sess.CurrentScene(),
			"transitioned", res.Transitioned,
			"input_len", len(line),
			"apply_ms", applyMs,
			"broadcast_ms", broadcastMs,
			"broadcast_sessions", broadcastN,
		)
		e.log.Debug("action_input", "session", sess.ID(), "input_preview", logInputPreview(line))
	}
	return nil
}

func (e *Engine) sceneKeys(sess *Session) bool {
	sc, ok := e.Experience().Scene(sess.CurrentScene())
	return ok && sc.Keys
}

func (e *Engine) syncKeys(sess *Session, term Terminal) {
	setter, ok := term.(KeyMode)
	if !ok {
		return
	}
	sc, ok := e.Experience().Scene(sess.CurrentScene())
	if !ok || !sc.Keys {
		setter.SetKeys(nil)
		return
	}
	keys := make([]string, 0, len(sc.Inputs))
	for k := range sc.Inputs {
		keys = append(keys, k)
	}
	setter.SetKeys(keys)
}

func (e *Engine) attachLineEditor(sess *Session, term Terminal) {
	ed, ok := term.(LineEditor)
	if !ok {
		return
	}
	ed.SetRedraw(func() {
		_ = e.Render(sess, term)
	})
	ed.SetComplete(func(prefix string) string {
		e.mu.RLock()
		exp := e.exp
		e.mu.RUnlock()
		return completeSceneInput(exp, sess.CurrentScene(), prefix)
	})
}

func (e *Engine) rejectInput(term Terminal, msg string) {
	if p, ok := term.(InputPrompter); ok {
		if err := p.RejectInput(msg); err != nil && e.log != nil {
			e.log.Debug("input_prompt_failed", "error", err)
		}
		return
	}
	e.beginInputLine(term)
}

func (e *Engine) beginInputLine(term Terminal) {
	if p, ok := term.(InputPrompter); ok {
		if err := p.BeginInputLine(); err != nil && e.log != nil {
			e.log.Debug("input_prompt_failed", "error", err)
		}
	}
}

func (e *Engine) commitInputLine(term Terminal, line string) {
	if p, ok := term.(InputPrompter); ok {
		if err := p.CommitInputLine(line); err != nil && e.log != nil {
			e.log.Debug("input_prompt_failed", "error", err)
		}
	}
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

func (noopTerminal) WriteFrame(string) error    { return nil }
func (noopTerminal) ReadInput() (string, error) { return "", fmt.Errorf("no terminal") }
func (noopTerminal) OnResize(func(int, int))    {}
func (noopTerminal) Close() error               { return nil }

// RunLoop drives read input → action → render until terminal closes.
// Input filtering layers are documented in terminalinput.go.
// activity, when non-nil, receives a signal on each non-empty input (for idle timeouts).
func (e *Engine) RunLoop(sess *Session, term Terminal, activity chan<- struct{}) error {
	setSessionTerminal(sess, term)
	defer sessionTerminals.Delete(sess.ID())
	if err := e.Render(sess, term); err != nil {
		return err
	}
	e.attachLineEditor(sess, term)
	e.syncKeys(sess, term)
	if !e.sceneKeys(sess) {
		e.beginInputLine(term)
	}
	for {
		e.syncKeys(sess, term)
		readStart := time.Now()
		in, err := term.ReadInput()
		readMs := time.Since(readStart).Milliseconds()
		if err != nil {
			return err
		}
		if in == "" {
			if e.log != nil {
				e.log.Debug("input_empty", "session", sess.ID(), "read_ms", readMs)
			}
			continue
		}
		if IsTerminalProtocolLine(in) {
			if e.log != nil {
				preview, n := logInputAttrs(in)
				e.log.Debug("input_protocol",
					"session", sess.ID(),
					"kind", protocolLineKind(in),
					"input_len", n,
					"input_preview", preview,
					"read_ms", readMs,
				)
			}
			continue
		}
		e.mu.RLock()
		exp := e.exp
		e.mu.RUnlock()
		scene := sess.CurrentScene()
		if !InputEligible(exp, scene, in) {
			if e.log != nil {
				preview, n := logInputAttrs(in)
				e.log.Debug("input_ineligible",
					"session", sess.ID(),
					"scene", scene,
					"input_len", n,
					"input_preview", preview,
					"read_ms", readMs,
				)
			}
			e.rejectInput(term, "no such command")
			continue
		}
		notifyActivity(activity)
		handleStart := time.Now()
		handleErr := e.HandleInput(sess, in)
		handleMs := time.Since(handleStart).Milliseconds()
		if handleErr != nil {
			if e.log != nil {
				e.log.Warn("input_error",
					"session", sess.ID(),
					"scene", scene,
					"input_len", len(in),
					"read_ms", readMs,
					"handle_ms", handleMs,
					"error", handleErr,
				)
				e.log.Debug("input_error_preview", "session", sess.ID(), "input_preview", logInputPreview(in))
			}
			_ = term.WriteFrame("Error: " + handleErr.Error() + "\n")
		}
		renderStart := time.Now()
		if err := e.Render(sess, term); err != nil {
			return err
		}
		if !e.sceneKeys(sess) {
			e.commitInputLine(term, in)
		}
		renderMs := time.Since(renderStart).Milliseconds()
		if e.log != nil {
			totalMs := readMs + handleMs + renderMs
			attrs := []any{
				"session", sess.ID(),
				"scene", sess.CurrentScene(),
				"read_ms", readMs,
				"handle_ms", handleMs,
				"render_ms", renderMs,
				"total_ms", totalMs,
			}
			if handleErr != nil {
				attrs = append(attrs, "error", handleErr)
			}
			if tr, ok := term.(LineEndTracer); ok {
				attrs = append(attrs, "line_end", tr.LineEndReason())
				if csi := tr.SubmitCSI(); csi != "" {
					attrs = append(attrs, "submit_csi", csi)
				}
				if skip := tr.SkippedCSI(); skip != "" {
					attrs = append(attrs, "skipped_csi", skip)
				}
			}
			e.log.Debug("input_cycle", attrs...)
		}
	}
}
