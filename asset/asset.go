package asset

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MrBlindMouse/Portico/experience"
)

// Loader loads named files into a Builder.
type Loader struct {
	dir string
}

// NewLoader reads assets from a directory.
func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

// Apply loads all regular files in dir into builder assets (basename → content).
func (l *Loader) Apply(b *experience.Builder) error {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(l.dir, e.Name()))
		if err != nil {
			return err
		}
		b.Asset(e.Name(), string(content))
	}
	return nil
}

// Reload reads all assets from dir.
func (l *Loader) Reload() (map[string]string, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(l.dir, e.Name()))
		if err != nil {
			return nil, err
		}
		out[e.Name()] = string(b)
	}
	return out, nil
}

// MergeAssets applies a content map onto a builder (overwriting names).
func MergeAssets(b *experience.Builder, assets map[string]string) {
	for name, content := range assets {
		b.Asset(name, content)
	}
}

// ValidateNames ensures asset names are safe basenames.
func ValidateNames(assets map[string]string) error {
	for name := range assets {
		if name == "" || name != filepath.Base(name) {
			return fmt.Errorf("portico: invalid asset name %q", name)
		}
	}
	return nil
}
