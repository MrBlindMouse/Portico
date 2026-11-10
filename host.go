package portico

import (
	"context"
	"fmt"
	"github.com/MrBlindMouse/Portico/components"
	"os"
	"sync/atomic"

	"github.com/MrBlindMouse/Portico/asset"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/runtime"
)

// Host is a wired Portico site: SSH server, engine, inspect/status, optional asset poll.
// Use NewHost when you need HotReload or other hooks before blocking on SSH.
type Host struct {
	Server *runtime.Server
	engine *runtime.Engine
}

// HotReload validates b and swaps the live installation.
func (h *Host) HotReload(b *installation.Builder) error {
	return HotReload(h.Server, b)
}

// Submit applies effects to one live session and repaints it.
// The site owns the trigger. This is not an action: no hooks, no rate limit.
func (h *Host) Submit(sessID string, effects []components.Effect) error {
	if err := h.notDraining(); err != nil {
		return err
	}
	return h.engine.Submit(sessID, effects)
}

// SubmitToSpace applies effects to members of space and repaints them.
// Space effects run once. An empty space applies nothing.
func (h *Host) SubmitToSpace(space components.SpaceID, effects []components.Effect) error {
	if err := h.notDraining(); err != nil {
		return err
	}
	return h.engine.SubmitToSpace(space, effects)
}

// SubmitAll applies effects to every live session and repaints them.
// Space effects run once.
func (h *Host) SubmitAll(effects []components.Effect) error {
	if err := h.notDraining(); err != nil {
		return err
	}
	return h.engine.SubmitAll(effects)
}

func (h *Host) notDraining() error {
	if h.Server != nil && h.Server.Draining() {
		return fmt.Errorf("portico: shutting down")
	}
	return nil
}

// NewHost prepares a site without listening. Starts inspect/status HTTP and asset polling.
func NewHost(addr string, exp *installation.Installation, opts Options) (*Host, error) {
	if exp == nil {
		return nil, fmt.Errorf("portico: installation is required")
	}
	opts = opts.withDefaults()

	store, err := openStore(opts)
	if err != nil {
		return nil, err
	}

	applier := runtime.NewApplier(nil, store)
	acct, err := resolveAccountStore(opts)
	if err != nil {
		return nil, err
	}
	if acct != nil {
		applier.SetAccountStore(acct)
		if opts.AccountBackend == nil {
			opts.AccountBackend = acct
		}
	}

	eng := runtime.NewEngine(exp, applier)
	eng.SetLogger(opts.Logger)
	eng.SetTheme(resolveTheme(opts))
	eng.SetInputPolicy(opts.ScrollbackEcho, opts.Transcript)
	if opts.ActionRate >= 0 {
		rate, burst := opts.ActionRate, opts.ActionBurst
		if rate == 0 {
			rate = 5
		}
		if burst == 0 {
			burst = 10
		}
		eng.EnableRateLimit(rate, burst)
	}
	if opts.ActionBefore != nil || opts.ActionAfter != nil {
		before, after := opts.ActionBefore, opts.ActionAfter
		eng.SetActionHooks(runtime.ActionHooks{
			Before: func(sess *runtime.Session, action components.ActionID, line string) (string, error) {
				if before == nil {
					return line, nil
				}
				return before(sess, action, line)
			},
			After: func(sess *runtime.Session, action components.ActionID, _ runtime.ApplyResult, err error) {
				if after != nil {
					after(sess, action, err)
				}
			},
		})
	}

	var actions atomic.Uint64
	eng.SetOnAction(func() { actions.Add(1) })

	if err := loadAssetsFromDir(opts, eng); err != nil {
		return nil, err
	}

	cfg, err := serverConfig(addr, opts)
	if err != nil {
		return nil, err
	}
	srv, err := runtime.NewServer(cfg, eng)
	if err != nil {
		return nil, err
	}

	host := &Host{Server: srv, engine: eng}
	startAuxiliaryHTTP(opts, addr, srv, eng, &actions)
	startAssetPoll(context.Background(), opts, eng)

	if opts.OnReady != nil {
		opts.OnReady(host)
	}
	return host, nil
}

func loadAssetsFromDir(opts Options, eng *runtime.Engine) error {
	if opts.AssetDir == "" {
		return nil
	}
	if _, err := os.Stat(opts.AssetDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	assets, err := asset.NewLoader(opts.AssetDir).Reload()
	if err != nil {
		return err
	}
	if len(assets) == 0 {
		return nil
	}
	return eng.UpdateAssets(assets)
}
