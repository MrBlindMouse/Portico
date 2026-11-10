package portico

import (
	"github.com/MrBlindMouse/Portico/asset"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/runtime"
)

// HotReload validates b and swaps the live installation on the server.
func HotReload(srv *runtime.Server, b *installation.Builder) error {
	return srv.ReloadInstallation(b)
}

// MergeAssetsIntoBuilder loads dir into builder.
func MergeAssetsIntoBuilder(b *installation.Builder, dir string) error {
	return asset.NewLoader(dir).Apply(b)
}
