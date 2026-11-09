package portico

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/MrBlindMouse/Portico/asset"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/runtime"
)

// Host is a wired Portico site: SSH server, engine, inspect/status, optional asset poll.
// Use NewHost when you need HotReload or other hooks before blocking on SSH.
type Host struct {
	Server *runtime.Server
	engine *runtime.Engine
}

// HotReload validates b and swaps the live experience.
func (h *Host) HotReload(b *experience.Builder) error {
	return HotReload(h.Server, b)
}

// NewHost prepares a site without listening. Starts inspect/status HTTP and asset polling.
func NewHost(addr string, exp *experience.Experience, opts Options) (*Host, error) {
	if exp == nil {
		return nil, fmt.Errorf("portico: experience is required")
	}
	opts = opts.withDefaults()

	store, err := openStore(opts)
	if err != nil {
		return nil, err
	}

	applier := runtime.NewApplier(nil, store)
	if acct := resolveAccountStore(opts); acct != nil {
		applier.SetAccountStore(acct)
	}

	eng := runtime.NewEngine(exp, applier)
	eng.SetLogger(opts.Logger)
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
			Before: func(sess *runtime.Session, action experience.ActionID, line string) (string, error) {
				if before == nil {
					return line, nil
				}
				return before(sess, action, line)
			},
			After: func(sess *runtime.Session, action experience.ActionID, _ runtime.ApplyResult, err error) {
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

	cfg := serverConfig(addr, opts)
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
