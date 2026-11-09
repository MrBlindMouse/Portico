// Package portico is the public entry for SSH sites built on Portico.
//
// Site authors (consumer repos):
//
//   - Build an experience with New() (scenes, actions, spaces, effects); actions return effects only.
//   - Call Serve or ServeDev with Options (data dir, limits, inspect HTTP).
//   - Call NewHost when you need HotReload or OnReady before ListenAndServe.
//
// Do not import runtime or internal from site code. Read docs/site-agents.md at your module
// version tag; run: portico docs -tag <version>. CLI bootstrap: portico init.
//
// Examples live in the module repo: examples/from-scratch, examples/hello. DESIGN.md is vision;
// README.md and docs/building-a-site.md are the operator guides.
package portico
