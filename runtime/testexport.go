package runtime

import (
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
)

// NewSessionForTest exposes session construction for tests in other packages.
func NewSessionForTest(id string, idn identity.Identity, root experience.SceneID) *Session {
	return newSession(id, idn, root)
}

// SetSessionTerminalForTest binds a terminal for broadcast fan-out in tests.
func SetSessionTerminalForTest(sess *Session, t Terminal) {
	setSessionTerminal(sess, t)
}

// NewActionContextForTest builds an action context for tests.
func NewActionContextForTest(sess *Session, exp *experience.Experience, eng *Engine) experience.ActionContext {
	return newActionContext(sess, exp, eng.spaces, eng.Applier().Store(), "", eng.viewTheme())
}
