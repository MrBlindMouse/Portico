package components

import (
	"fmt"
	"strings"
)

// SceneID identifies a scene within an installation.
type SceneID string

// ActionID identifies an action within an installation.
type ActionID string

// SpaceID identifies a shared space within an installation.
type SpaceID string

func validateID(kind, value string) error {
	if value == "" {
		return fmt.Errorf("portico: %s id must not be empty", kind)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("portico: %s id %q has surrounding whitespace", kind, value)
	}
	return nil
}

func (id SceneID) Validate() error  { return validateID("scene", string(id)) }
func (id ActionID) Validate() error { return validateID("action", string(id)) }
func (id SpaceID) Validate() error  { return validateID("space", string(id)) }
