package runtime

import (
	"fmt"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
)

// Submit applies effects to one registered session and paints it.
// Space effects in the slice run once. Session effects run on that session.
// Unknown ids return an error. Hooks, the action rate limiter, and onAction are not used.
func (e *Engine) Submit(sessID string, effects []components.Effect) error {
	e.mu.RLock()
	sess := e.sessions[sessID]
	exp := e.exp
	e.mu.RUnlock()
	if sess == nil {
		return fmt.Errorf("portico: no such session %q", sessID)
	}
	return e.submitTargets(exp, []*Session{sess}, effects, sessID)
}

// SubmitToSpace applies effects to current members of space.
// Space effects run once. Session effects run once per member.
// An empty member list applies nothing.
func (e *Engine) SubmitToSpace(space components.SpaceID, effects []components.Effect) error {
	e.mu.RLock()
	exp := e.exp
	e.mu.RUnlock()
	targets := e.spaces.members(space)
	if len(targets) == 0 {
		return nil
	}
	return e.submitTargets(exp, targets, effects, string(space))
}

// SubmitAll applies effects to every registered session.
// Space effects run once. Session effects run once per session.
// An empty registry applies nothing.
func (e *Engine) SubmitAll(effects []components.Effect) error {
	e.mu.RLock()
	exp := e.exp
	targets := make([]*Session, 0, len(e.sessions))
	for _, sess := range e.sessions {
		targets = append(targets, sess)
	}
	e.mu.RUnlock()
	if len(targets) == 0 {
		return nil
	}
	return e.submitTargets(exp, targets, effects, "all")
}

func (e *Engine) submitTargets(exp *installation.Installation, targets []*Session, effects []components.Effect, label string) error {
	spaceEff, sessEff := splitSubmitEffects(effects)
	var transitioned bool
	var broadcasts []components.SpaceID

	painted := map[string]struct{}{}
	paint := func(sess *Session) {
		if sess == nil {
			return
		}
		if _, ok := painted[sess.ID()]; ok {
			return
		}
		painted[sess.ID()] = struct{}{}
		e.paintSession(sess)
		e.syncKeys(sess, memberTerminal(sess))
	}

	paintBroadcasts := func() int {
		n := 0
		seen := map[string]struct{}{}
		for _, space := range broadcasts {
			for _, member := range e.spaces.members(space) {
				if _, ok := seen[member.ID()]; ok {
					continue
				}
				seen[member.ID()] = struct{}{}
				n++
				paint(member)
			}
		}
		return n
	}

	if len(spaceEff) > 0 {
		res, err := e.applier.Apply(targets[0], exp, spaceEff)
		broadcasts = append(broadcasts, res.Broadcast...)
		if err != nil {
			for _, sess := range targets {
				paint(sess)
			}
			paintBroadcasts()
			return err
		}
		transitioned = res.Transitioned
	}
	for i, sess := range targets {
		if len(sessEff) == 0 {
			continue
		}
		res, err := e.applier.Apply(sess, exp, sessEff)
		if err != nil {
			for _, done := range targets[:i+1] {
				paint(done)
			}
			paintBroadcasts()
			return err
		}
		if res.Transitioned {
			transitioned = true
		}
		broadcasts = append(broadcasts, res.Broadcast...)
	}
	for _, sess := range targets {
		paint(sess)
	}
	broadcastN := paintBroadcasts()
	if e.log != nil {
		e.log.Info("submit",
			"target", label,
			"effects", len(effects),
			"transitioned", transitioned,
			"broadcast_sessions", broadcastN,
		)
	}
	return nil
}

func splitSubmitEffects(effects []components.Effect) (space, session []components.Effect) {
	for _, eff := range effects {
		switch eff.(type) {
		case components.SpaceSet, components.SpaceCAS, components.AppendSpaceList, components.Broadcast:
			space = append(space, eff)
		default:
			session = append(session, eff)
		}
	}
	return space, session
}
