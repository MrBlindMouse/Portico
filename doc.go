// Package portico is the public entry for SSH sites built on Portico.
//
// Site authors (consumer repos):
//
//   - Build an installation with New() (components: scenes, actions, spaces; effects in installation); actions return effects only.
//   - Call Serve or ServeDev with Options (data dir, limits, inspect HTTP).
//   - Call NewHost when you need HotReload or OnReady before ListenAndServe.
//
// Do not import runtime or inspect from site code. After portico init, read AGENTS.md in the site repo;
// use docs/manual and docs/site-agents.md at your go.mod tag (portico docs -tag <version>).
// DESIGN.md is vision; README.md, docs/manual, and docs/building-a-site.md are the operator guides.
package portico
