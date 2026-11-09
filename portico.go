package portico

import "github.com/MrBlindMouse/Portico/experience"

// Builder constructs an experience (alias for agent discoverability).
type Builder = experience.Builder

// New starts a new experience builder.
func New() *Builder { return experience.NewBuilder() }
