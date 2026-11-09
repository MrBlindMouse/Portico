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

func openStore(dataDir string) (identity.Store, error) {
	if dataDir != "" {
		return identity.OpenFileStore(dataDir)
	}
	return identity.NewMemoryStore(), nil
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
	if opts.MaxConns > 0 {
		cfg.MaxConns = opts.MaxConns
	}
	if opts.MaxConnsPerIP > 0 {
		cfg.MaxConnsPerIP = opts.MaxConnsPerIP
	}
	if opts.Accounts != nil {
		cfg.Accounts = opts.Accounts
	} else if opts.AccountStore != nil {
		cfg.Accounts = &runtime.AccountConfig{Enabled: true, Accounts: opts.AccountStore}
	}
	return cfg
}

func startAuxiliaryHTTP(opts Options, sshAddr string, eng *runtime.Engine, actions *atomic.Uint64) {
	if opts.InspectAddr != "" {
		go func() {
			opts.Logger.Info("inspect listening", "addr", opts.InspectAddr)
			_ = http.ListenAndServe(opts.InspectAddr, inspect.NewHandler(eng))
		}()
	}
	if opts.StatusAddr != "" {
		go func() {
			opts.Logger.Info("status listening", "addr", opts.StatusAddr)
			_ = http.ListenAndServe(opts.StatusAddr, statusHandler(sshAddr, actions))
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
