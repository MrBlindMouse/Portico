package portico

import (
	"log/slog"
	"time"

	"github.com/MrBlindMouse/Portico/config"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
)

// Options configures Serve.
type Options struct {
	DataDir       string
	HostKeyPath   string
	InspectAddr   string
	StatusAddr    string
	IdleTimeout   time.Duration
	MaxConns      int
	MaxConnsPerIP int
	AllowKeys     []string
	DenyKeys      []string
	// ActionRate is actions per second per session. 0 uses the default (5/s, burst 10).
	// A negative rate disables limiting. ActionBurst 0 uses the default burst when limiting is on.
	ActionRate     float64
	ActionBurst    float64
	AssetDir       string
	AssetPollEvery time.Duration
	Logger         *slog.Logger
	LogLevel       string
	// Account stores are chosen in order: AccountBackend, then HashedAccounts with
	// DataDir (bcrypt files), then AccountStore, then Accounts.Accounts.
	AccountStore      *runtime.AccountStore
	AccountBackend    identity.AccountStore
	IdentityStore     identity.Store
	Accounts          *runtime.AccountConfig
	AllowAccounts     []string
	DenyAccounts      []string
	RequireAccount    bool
	HashedAccounts    bool
	ShutdownGrace     time.Duration
	KeepAliveInterval time.Duration
	KeepAliveCountMax int
	// ActionBefore and ActionAfter run around action dispatch without importing runtime.
	ActionBefore func(experience.SessionView, experience.ActionID, string) (string, error)
	ActionAfter  func(experience.SessionView, experience.ActionID, error)
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
	if cfg.ActionRate != 0 {
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
	if len(cfg.AllowAccounts) > 0 {
		o.AllowAccounts = cfg.AllowAccounts
	}
	if len(cfg.DenyAccounts) > 0 {
		o.DenyAccounts = cfg.DenyAccounts
	}
	if cfg.RequireAccount {
		o.RequireAccount = true
	}
	if cfg.ShutdownGrace > 0 {
		o.ShutdownGrace = cfg.ShutdownGrace
	}
	if cfg.KeepAliveInterval > 0 {
		o.KeepAliveInterval = cfg.KeepAliveInterval
	}
	if cfg.KeepAliveCountMax > 0 {
		o.KeepAliveCountMax = cfg.KeepAliveCountMax
	}
	if cfg.HashedAccounts {
		o.HashedAccounts = true
	}
	return o
}
