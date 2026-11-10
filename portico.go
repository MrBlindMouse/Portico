package portico

import "github.com/MrBlindMouse/Portico/installation"

// Builder constructs an installation (alias for agent discoverability).
type Builder = installation.Builder

// New starts a new installation builder.
func New() *Builder { return installation.NewBuilder() }
