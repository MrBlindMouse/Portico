package portico

import (
	"github.com/MrBlindMouse/Portico/asset"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/runtime"
)

// HotReload validates b and swaps the live experience on the server.
func HotReload(srv *runtime.Server, b *experience.Builder) error {
	return srv.ReloadExperience(b)
}

// MergeAssetsIntoBuilder loads dir into builder.
func MergeAssetsIntoBuilder(b *experience.Builder, dir string) error {
	return asset.NewLoader(dir).Apply(b)
}
