package portico

import (
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
)

// ServeWarnings returns operator-facing hints for common misconfiguration.
// It does not block startup.
func ServeWarnings(exp *installation.Installation, opts Options) []string {
	if exp == nil {
		return nil
	}
	var out []string
	scenes := make(map[components.SceneID]struct{}, len(exp.Scenes()))
	for _, id := range exp.Scenes() {
		scenes[id] = struct{}{}
	}
	if _, ok := scenes["register"]; ok && !accountsEnabled(opts) {
		out = append(out, "installation has scene \"register\" but options enable no account store (see HashedAccounts, AccountBackend, or dev AccountStore)")
	}
	if _, ok := scenes["blog-list"]; ok && opts.AssetDir == "" && !hasPostAssets(exp) {
		out = append(out, "installation has scene \"blog-list\" but no post-* text assets and AssetDir is empty")
	}
	return out
}

func accountsEnabled(opts Options) bool {
	if opts.HashedAccounts || opts.AccountBackend != nil {
		return true
	}
	if opts.AccountStore != nil {
		return true
	}
	if opts.Accounts != nil && opts.Accounts.Enabled {
		return true
	}
	return false
}

func hasPostAssets(exp *installation.Installation) bool {
	for name := range exp.Assets() {
		if strings.HasPrefix(name, "post-") {
			return true
		}
	}
	return false
}