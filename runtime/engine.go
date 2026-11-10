package runtime

import (
	"fmt"
	"github.com/MrBlindMouse/Portico/components"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// Engine runs the session loop against a Terminal.
type Engine struct {
	mu             sync.RWMutex
	exp            *installation.Installation
	spaces         *spaceRegistry
	applier        *Applier
	sessions       map[string]*Session
	nextID         atomic.Uint64
	onAction       func()
	hooks          ActionHooks
	limits         *rateLimiter
	log            *slog.Logger
	reloadFailures atomic.Uint64
	theme          render.Theme
	themeSet       bool
	scrollbackEcho bool
	transcript     TranscriptMode
}

// TranscriptMode controls RecentCommands across a scene change.
type TranscriptMode int

const (
	// TranscriptClear drops the list when the scene changes. Zero value.
	TranscriptClear TranscriptMode = iota
	// TranscriptKeep leaves the list in place, including the line that transitioned.
	TranscriptKeep
	// TranscriptOff does not record lines.
	TranscriptOff
)

// SetInputPolicy chooses prompt scrollback and the command transcript.
// Zero values match the defaults: a bare prompt, and a transcript cleared on scene change.
func (e *Engine) SetInputPolicy(scrollbackEcho bool, transcript TranscriptMode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.scrollbackEcho = scrollbackEcho
	e.transcript = transcript
}

// SetTheme installs the theme RenderContext.Theme returns.
// Call it before sessions connect. An unset theme is render.Default.
func (e *Engine) SetTheme(t render.Theme) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.theme = t
	e.themeSet = true
}

func (e *Engine) viewTheme() render.Theme {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if !e.themeSet {
		return render.Default
	}
	return e.theme
}

// ActionHooks optionally intercept action dispatch.
type ActionHooks struct {
	Before func(sess *Session, action components.ActionID, line string) (string, error)
	After  func(sess *Session, action components.ActionID, result ApplyResult, err error)
}

// EnableRateLimit enables per-session action throttling.
func (e *Engine) EnableRateLimit(rate, burst float64) {
	e.limits = newRateLimiter(rate, burst)
}

// NewEngine creates a runtime engine.
func NewEngine(exp *installation.Installation, applier *Applier) *Engine {
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

// SetInstallation hot-swaps the installation definition after validation.
func (e *Engine) SetInstallation(exp *installation.Installation) error {
	if exp == nil {
		return fmt.Errorf("portico: installation is required")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, sess := range e.sessions {
		if _, ok := exp.Scene(sess.CurrentScene()); !ok {
			return fmt.Errorf("portico: session %s in scene %q not in new installation", sess.ID(), sess.CurrentScene())
		}
		for _, sp := range sess.Spaces() {
			if _, ok := exp.Space(sp); !ok {
				return fmt.Errorf("portico: session %s in space %q not in new installation", sess.ID(), sp)
			}
		}
	}
	e.exp = exp
	return nil
}

// RedrawSessions paints the current scene for every session that has a terminal.
func (e *Engine) RedrawSessions() {
	for _, sess := range e.Sessions() {
		e.paintSession(sess)
	}
}

// paintSession renders one session and restores the line editor.
// A session with no terminal is skipped.
func (e *Engine) paintSession(sess *Session) {
	term := memberTerminal(sess)
	if _, skip := term.(noopTerminal); skip {
		return
	}
	_ = e.Render(sess, term)
	if ed, ok := term.(LineEditor); ok {
		ed.RestoreInputLine()
	}
}

// UpdateAssets replaces asset content on the live installation.
func (e *Engine) UpdateAssets(assets map[string]string) error {
	e.mu.RLock()
	cur := e.exp
	e.mu.RUnlock()
	return e.SetInstallation(cur.CloneWithAssets(assets))
}

// Installation returns the current definition.
func (e *Engine) Installation() *installation.Installation {
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
	out, _, err := e.renderScene(sess)
	if err != nil {
		return "", err
	}
	sess.setLastRender(out)
	return out, nil
}

// Render writes the current scene to the terminal.
// A placed input is painted inside its rectangle. Otherwise the caller draws the prompt row.
func (e *Engine) Render(sess *Session, term Terminal) error {
	out, slot, err := e.renderScene(sess)
	if err != nil {
		return err
	}
	sess.setLastRender(out)
	sess.setSlot(slot)
	syncInputSlot(term, slot)
	if slot != nil {
		if err := writeGrid(term, out); err != nil {
			return err
		}
		paintInputSlot(term)
		return nil
	}
	if e.sceneKeys(sess) {
		return writeGrid(term, out)
	}
	return term.WriteFrame(out)
}

func (e *Engine) renderScene(sess *Session) (string, *render.Input, error) {
	e.mu.RLock()
	exp := e.exp
	e.mu.RUnlock()
	sc, ok := exp.Scene(sess.CurrentScene())
	if !ok {
		return "", nil, fmt.Errorf("portico: session scene %q undefined", sess.CurrentScene())
	}
	place := &inputPlace{}
	ctx := newSceneContext(sess, exp, e.spaces, e.applier.Store(), e.viewTheme(), place)
	out := sc.Render(ctx)
	if sc.Keys && place.ok {
		return "", nil, fmt.Errorf("portico: scene %q places an input and has Keys set", sess.CurrentScene())
	}
	if !place.ok {
		return out, nil, nil
	}
	in := place.in
	return out, &in, nil
}

// inputSlotter is the line editor's caret rectangle. Nil clears it.
type inputSlotter interface {
	SetSlot(*render.Input)
	PaintSlot()
}

// slotRejecter reports an unmatched line inside the caret rectangle.
type slotRejecter interface {
	SlotActive() bool
	RejectSlot(msg string)
}

type gridWriter interface {
	WriteGrid(text string) error
}

func syncInputSlot(term Terminal, slot *render.Input) {
	s, ok := term.(inputSlotter)
	if !ok {
		return
	}
	s.SetSlot(slot)
}

func paintInputSlot(term Terminal) {
	s, ok := term.(inputSlotter)
	if !ok {
		return
	}
	s.PaintSlot()
}

func writeGrid(term Terminal, text string) error {
	if g, ok := term.(gridWriter); ok {
		return g.WriteGrid(text)
	}
	return term.WriteFrame(text)
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
	ctx := newActionContext(sess, exp, e.spaces, e.applier.Store(), line, e.viewTheme())
	effects := act.Run(ctx)
	res, err := e.applier.Apply(sess, exp, effects)
	applyMs := time.Since(applyStart).Milliseconds()
	if err == nil {
		e.recordCommand(sess, line, res.Transitioned)
	}
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
	sc, ok := e.Installation().Scene(sess.CurrentScene())
	return ok && sc.Keys
}

func (e *Engine) syncKeys(sess *Session, term Terminal) {
	setter, ok := term.(KeyMode)
	if !ok {
		return
	}
	sc, ok := e.Installation().Scene(sess.CurrentScene())
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
	if s, ok := term.(slotRejecter); ok && s.SlotActive() {
		s.RejectSlot(msg)
		return
	}
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

// recallClearer drops editor up/down recall. SSHTerminal implements it.
type recallClearer interface {
	ClearRecall()
}

func (e *Engine) recordCommand(sess *Session, line string, transitioned bool) {
	e.mu.RLock()
	mode := e.transcript
	e.mu.RUnlock()
	switch mode {
	case TranscriptOff:
		return
	case TranscriptKeep:
		sess.noteCommand(line)
	default:
		if transitioned {
			sess.clearCommands()
			return
		}
		sess.noteCommand(line)
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
	if _, slotted := sess.inputSlot(); !e.sceneKeys(sess) && !slotted {
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
			if in == "" {
				if e.log != nil {
					e.log.Debug("input_empty", "session", sess.ID(), "read_ms", readMs)
				}
				continue
			}
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
		if _, slotted := sess.inputSlot(); !e.sceneKeys(sess) && !slotted {
			e.mu.RLock()
			echo := e.scrollbackEcho
			e.mu.RUnlock()
			if echo {
				e.commitInputLine(term, in)
			} else {
				e.beginInputLine(term)
			}
		}
		if scene != sess.CurrentScene() {
			if c, ok := term.(recallClearer); ok {
				c.ClearRecall()
			}
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
