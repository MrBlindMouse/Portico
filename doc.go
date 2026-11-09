// Package portico is the public entry for SSH sites built on Portico.
//
// Site authors typically:
//
//   - Build an experience with portico.New() (scenes, actions, spaces, effects).
//   - Call Serve or ServeDev with portico.Options (data dir, limits, inspect HTTP).
//   - Call NewHost when you need HotReload or OnReady before ListenAndServe.
//
// See examples/hello for a full site. DESIGN.md describes architecture; README.md is the framework guide.
package portico
