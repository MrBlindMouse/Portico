package portico

import (
	"context"
	"log/slog"
	"net/http"
	"os"
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
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
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
			o.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
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

func resolveAccountStore(opts Options) identity.AccountStore {
	if opts.AccountBackend != nil {
		return opts.AccountBackend
	}
	if opts.HashedAccounts && opts.DataDir != "" {
		store, err := identity.OpenFileAccountStore(opts.DataDir)
		if err == nil {
			return store
		}
	}
	if opts.AccountStore != nil {
		return opts.AccountStore
	}
	if opts.Accounts != nil && opts.Accounts.Accounts != nil {
		return opts.Accounts.Accounts
	}
	return nil
}

func serverConfig(addr string, opts Options) runtime.ServerConfig {
	cfg := runtime.DefaultConfig()
	cfg.Addr = addr
	cfg.DataDir = opts.DataDir
	cfg.HostKeyPath = opts.HostKeyPath
	cfg.IdleTimeout = opts.IdleTimeout
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
	acct := resolveAccountStore(opts)
	if opts.Accounts != nil {
		cfg.Accounts = opts.Accounts
		if cfg.Accounts.Accounts == nil {
			cfg.Accounts.Accounts = acct
		}
	} else if acct != nil {
		cfg.Accounts = &runtime.AccountConfig{Enabled: true, Accounts: acct}
	}
	return cfg
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
