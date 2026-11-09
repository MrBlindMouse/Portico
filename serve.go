package portico

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/MrBlindMouse/Portico/asset"
	"github.com/MrBlindMouse/Portico/config"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/inspect"
	"github.com/MrBlindMouse/Portico/runtime"
)

// ServeDev listens on :2222 with dev-friendly inspect/status defaults.
func ServeDev(exp *experience.Experience, opts Options) error {
	if opts.InspectAddr == "" {
		opts.InspectAddr = "127.0.0.1:9922"
	}
	if opts.StatusAddr == "" {
		opts.StatusAddr = "127.0.0.1:9923"
	}
	return Serve(runtime.DefaultConfig().Addr, exp, opts)
}

// Serve blocks serving exp on addr until interrupted (SIGINT/SIGTERM).
func Serve(addr string, exp *experience.Experience, opts Options) error {
	host, err := NewHost(addr, exp, opts)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	return host.Server.ListenAndServe(ctx)
}

func (o Options) withDefaults() Options {
	if o.Logger == nil {
		if o.LogLevel != "" {
			o.Logger = config.NewLogger(o.LogLevel)
		} else {
			o.Logger = config.NewLogger("info")
		}
	}
	return o
}

func openStore(opts Options) (identity.Store, error) {
	if opts.IdentityStore != nil {
		return opts.IdentityStore, nil
	}
	if opts.DataDir != "" {
		return identity.OpenFileStore(opts.DataDir)
	}
	return identity.NewMemoryStore(), nil
}

// resolveAccountStore picks AccountBackend, then a bcrypt file store when
// HashedAccounts is set, then AccountStore, then Accounts.Accounts.
// A hashed-account open that fails is returned; it does not fall through to another store.
func resolveAccountStore(opts Options) (identity.AccountStore, error) {
	if opts.AccountBackend != nil {
		return opts.AccountBackend, nil
	}
	if opts.HashedAccounts {
		if opts.DataDir == "" {
			return nil, fmt.Errorf("portico: hashed accounts require DataDir")
		}
		return identity.OpenFileAccountStore(opts.DataDir)
	}
	if opts.AccountStore != nil {
		return opts.AccountStore, nil
	}
	if opts.Accounts != nil && opts.Accounts.Accounts != nil {
		return opts.Accounts.Accounts, nil
	}
	return nil, nil
}

func serverConfig(addr string, opts Options) (runtime.ServerConfig, error) {
	cfg := runtime.DefaultConfig()
	cfg.Addr = addr
	cfg.DataDir = opts.DataDir
	cfg.HostKeyPath = opts.HostKeyPath
	if opts.IdleTimeout > 0 {
		cfg.IdleTimeout = opts.IdleTimeout
	} else if opts.IdleTimeout < 0 {
		cfg.IdleTimeout = 0
	}
	cfg.Logger = opts.Logger
	cfg.AllowKeys = opts.AllowKeys
	cfg.DenyKeys = opts.DenyKeys
	cfg.AllowAccounts = opts.AllowAccounts
	cfg.DenyAccounts = opts.DenyAccounts
	cfg.RequireAccount = opts.RequireAccount
	cfg.ShutdownGrace = opts.ShutdownGrace
	cfg.KeepAliveInterval = opts.KeepAliveInterval
	cfg.KeepAliveCountMax = opts.KeepAliveCountMax
	if opts.MaxConns > 0 {
		cfg.MaxConns = opts.MaxConns
	}
	if opts.MaxConnsPerIP > 0 {
		cfg.MaxConnsPerIP = opts.MaxConnsPerIP
	}
	acct, err := resolveAccountStore(opts)
	if err != nil {
		return runtime.ServerConfig{}, err
	}
	if opts.Accounts != nil {
		cfg.Accounts = opts.Accounts
		if cfg.Accounts.Accounts == nil {
			cfg.Accounts.Accounts = acct
		}
	} else if acct != nil {
		cfg.Accounts = &runtime.AccountConfig{Enabled: true, Accounts: acct}
	}
	return cfg, nil
}

func startAuxiliaryHTTP(opts Options, sshAddr string, srv *runtime.Server, eng *runtime.Engine, actions *atomic.Uint64) {
	if opts.InspectAddr != "" {
		go func() {
			opts.Logger.Info("inspect listening", "addr", opts.InspectAddr)
			_ = http.ListenAndServe(opts.InspectAddr, inspect.NewHandler(eng))
		}()
	}
	if opts.StatusAddr != "" {
		go func() {
			opts.Logger.Info("status listening", "addr", opts.StatusAddr)
			_ = http.ListenAndServe(opts.StatusAddr, statusHandler(sshAddr, actions, defaultStatus(srv, eng, opts.DataDir)))
		}()
	}
}

func startAssetPoll(ctx context.Context, opts Options, eng *runtime.Engine) {
	if opts.AssetDir == "" {
		return
	}
	interval := opts.AssetPollEvery
	if interval <= 0 {
		interval = time.Second
	}
	loader := asset.NewLoader(opts.AssetDir)
	poll := &asset.Poll{Loader: loader, Interval: interval}
	go poll.Run(ctx.Done(), func(assets map[string]string) error {
		return eng.UpdateAssets(assets)
	})
}
