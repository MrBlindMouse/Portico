// Package installation assembles components into a validated Installation.
//
// Scenes, actions, spaces, effects, and contexts live in package components.
// Build() checks references before Serve. The runtime apply step is the only
// mutator for session and space state. See docs/site-agents.md.
package installation
