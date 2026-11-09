package config

import (
	"encoding/json"
	"os"
	"time"
)

// File is a simple JSON config for operators.
type File struct {
	Listen        string        `json:"listen"`
	DataDir       string        `json:"data_dir"`
	LogLevel      string        `json:"log_level"`
	InspectAddr   string        `json:"inspect_addr"`
	StatusAddr    string        `json:"status_addr"`
	IdleTimeout   time.Duration `json:"idle_timeout"`
	MaxConns      int           `json:"max_conns"`
	MaxConnsPerIP int           `json:"max_conns_per_ip"`
	AllowKeys     []string      `json:"allow_keys"`
	DenyKeys      []string      `json:"deny_keys"`
	ActionRate    float64       `json:"action_rate"`
	ActionBurst   float64       `json:"action_burst"`
	AssetDir        string        `json:"asset_dir"`
	AssetPollEvery  time.Duration `json:"asset_poll_interval"`
	Accounts        bool          `json:"accounts_enabled"`
}

// Load reads path if present; returns zero value and nil error when missing.
func Load(path string) (File, error) {
	var f File
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return f, nil
	}
	if err != nil {
		return f, err
	}
	err = json.Unmarshal(b, &f)
	return f, err
}
