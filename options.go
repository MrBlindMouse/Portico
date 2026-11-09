package portico

import (
	"log/slog"
	"time"

	"github.com/MrBlindMouse/Portico/config"
	"github.com/MrBlindMouse/Portico/runtime"
)

// Options configures Serve.
type Options struct {
	DataDir        string
	HostKeyPath    string
	InspectAddr    string
	StatusAddr     string
	IdleTimeout    time.Duration
	MaxConns       int
	MaxConnsPerIP  int
	AllowKeys      []string
	DenyKeys       []string
	ActionRate     float64
	ActionBurst    float64
	AssetDir       string
	AssetPollEvery time.Duration
	Logger         *slog.Logger
	LogLevel       string
	AccountStore   *runtime.AccountStore
	Accounts       *runtime.AccountConfig
	// OnReady runs after the host is wired, before Serve returns (and before ListenAndServe in NewHost-only flows).
	OnReady func(*Host)
}

// MergeConfig overlays non-zero fields from a config file onto opts.
func (o Options) MergeConfig(cfg config.File) Options {
	if cfg.Listen != "" {
		// listen address is passed to Serve(addr), not stored on Options
	}
	if cfg.DataDir != "" {
		o.DataDir = cfg.DataDir
	}
	if cfg.InspectAddr != "" {
		o.InspectAddr = cfg.InspectAddr
	}
	if cfg.StatusAddr != "" {
		o.StatusAddr = cfg.StatusAddr
	}
	if cfg.IdleTimeout > 0 {
		o.IdleTimeout = cfg.IdleTimeout
	}
	if cfg.MaxConns > 0 {
		o.MaxConns = cfg.MaxConns
	}
	if cfg.MaxConnsPerIP > 0 {
		o.MaxConnsPerIP = cfg.MaxConnsPerIP
	}
	if len(cfg.AllowKeys) > 0 {
		o.AllowKeys = cfg.AllowKeys
	}
	if len(cfg.DenyKeys) > 0 {
		o.DenyKeys = cfg.DenyKeys
	}
	if cfg.ActionRate > 0 {
		o.ActionRate = cfg.ActionRate
	}
	if cfg.ActionBurst > 0 {
		o.ActionBurst = cfg.ActionBurst
	}
	if cfg.AssetDir != "" {
		o.AssetDir = cfg.AssetDir
	}
	if cfg.AssetPollEvery > 0 {
		o.AssetPollEvery = cfg.AssetPollEvery
	}
	return o
}
