package runtime

import (
	"fmt"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
)

// ApplyResult summarizes effect application for hooks and metrics.
type ApplyResult struct {
	Transitioned bool
	Broadcast    []experience.SpaceID
}

// Applier applies effects to session and shared state.
type Applier struct {
	spaces   *spaceRegistry
	store    identity.Store
	accounts identity.AccountStore
	pending  []persistFlush
}

type persistFlush struct {
	id  identity.Identity
	key string
	val any
}

// SetAccountStore enables RegisterAccount effects when password auth is enabled.
func (a *Applier) SetAccountStore(s identity.AccountStore) {
	a.accounts = s
}

// NewApplier constructs an effect applier.
func NewApplier(spaces *spaceRegistry, store identity.Store) *Applier {
	if store == nil {
		store = identity.NewMemoryStore()
	}
	if spaces == nil {
		spaces = newSpaceRegistry()
	}
	return &Applier{spaces: spaces, store: store}
}

// Apply mutates session and spaces according to effects. Only this function writes session.current and session.local.
func (a *Applier) Apply(sess *Session, exp *experience.Experience, effects []experience.Effect) (ApplyResult, error) {
	var res ApplyResult
	for _, eff := range effects {
		switch e := eff.(type) {
		case experience.SetLocal:
			sc, ok := exp.Scene(sess.CurrentScene())
			if ok && sc.ValidateLocal != nil {
				if err := sc.ValidateLocal(e.Key, e.Value); err != nil {
					return res, err
				}
			}
			sess.mu.Lock()
			sess.local[e.Key] = e.Value
			sess.mu.Unlock()
		case experience.Transition:
			sess.mu.Lock()
			if _, ok := exp.Scene(e.To); !ok {
				sess.mu.Unlock()
				return res, fmt.Errorf("portico: undefined scene %q", e.To)
			}
			sess.current = e.To
			sess.mu.Unlock()
			res.Transitioned = true
		case experience.Persist:
			idn := sess.Identity()
			a.pending = append(a.pending, persistFlush{id: idn, key: e.Key, val: e.Value})
			if err := a.store.Set(idn, e.Key, e.Value); err != nil {
				return res, err
			}
		case experience.JoinSpace:
			sp, ok := exp.Space(e.Space)
			if !ok {
				return res, fmt.Errorf("portico: undefined space %q", e.Space)
			}
			sess.mu.Lock()
			sess.spaces[e.Space] = struct{}{}
			sess.mu.Unlock()
			a.spaces.join(e.Space, sess)
			if len(sp.OnJoin) > 0 {
				joinRes, err := a.Apply(sess, exp, sp.OnJoin)
				if err != nil {
					return res, err
				}
				res.Transitioned = res.Transitioned || joinRes.Transitioned
				res.Broadcast = append(res.Broadcast, joinRes.Broadcast...)
			}
		case experience.LeaveSpace:
			sp, ok := exp.Space(e.Space)
			if !ok {
				return res, fmt.Errorf("portico: undefined space %q", e.Space)
			}
			if len(sp.OnLeave) > 0 {
				leaveRes, err := a.Apply(sess, exp, sp.OnLeave)
				if err != nil {
					return res, err
				}
				res.Transitioned = res.Transitioned || leaveRes.Transitioned
				res.Broadcast = append(res.Broadcast, leaveRes.Broadcast...)
			}
			sess.mu.Lock()
			delete(sess.spaces, e.Space)
			sess.mu.Unlock()
			a.spaces.leave(e.Space, sess)
		case experience.SpaceSet:
			sp, ok := exp.Space(e.Space)
			if !ok {
				return res, fmt.Errorf("portico: undefined space %q", e.Space)
			}
			if sp.ValidateSet != nil {
				old, _ := a.spaces.getValue(e.Space, e.Key)
				if err := sp.ValidateSet(e.Key, old, e.Value); err != nil {
					return res, err
				}
			}
			a.spaces.set(e.Space, e.Key, e.Value)
		case experience.SpaceCAS:
			sp, ok := exp.Space(e.Space)
			if ok && sp.ValidateSet != nil {
				old, _ := a.spaces.getValue(e.Space, e.Key)
				if err := sp.ValidateSet(e.Key, old, e.Value); err != nil {
					return res, err
				}
			}
			if !a.spaces.cas(e.Space, e.Key, e.Expect, e.Value) {
				return res, fmt.Errorf("portico: space cas failed for %q", e.Key)
			}
		case experience.AppendSpaceList:
			a.spaces.appendList(e.Space, e.Key, e.Value)
		case experience.Broadcast:
			res.Broadcast = append(res.Broadcast, e.Space)
		case experience.ReserveNickname:
			if err := identity.ValidateNickname(e.Name); err != nil {
				return res, err
			}
			if err := a.store.ReserveNickname(sess.Identity(), e.Name); err != nil {
				return res, err
			}
			sess.setIdentity(sess.Identity().WithNickname(e.Name))
		case experience.RegisterAccount:
			if a.accounts == nil {
				return res, fmt.Errorf("portico: account registration not enabled")
			}
			if e.Username == "" || e.Password == "" {
				return res, fmt.Errorf("portico: username and password required")
			}
			if err := a.accounts.Register(e.Username, e.Password); err != nil {
				return res, err
			}
		default:
			return res, fmt.Errorf("portico: unknown effect %T", eff)
		}
	}
	return res, nil
}

// LoadPersisted reads identity store into scene context helpers (not session local).
func (a *Applier) LoadPersisted(idn identity.Identity, key string) (any, bool, error) {
	return a.store.Get(idn, key)
}

// Store returns the identity persistence backend.
func (a *Applier) Store() identity.Store {
	return a.store
}
