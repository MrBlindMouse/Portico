# runtime/

SSH server, session lifecycle, input dispatch, and effect application. **Site modules do not import this package.** They call `portico.Serve`, `ServeDev`, or `NewHost` / `Host.HotReload`.

## Invariants

- Only `apply.go` mutates `Session` and space registry state (`session.go` updates identity nickname and render cache).
- Session constructor is unexported (`newSession`).
- Space hooks `OnJoin` / `OnLeave` run as effect lists through `Apply`.
- `SetInstallation` refuses definitions that orphan live sessions (current scene or a joined space).
- A failed effect leaves earlier effects in that slice applied. Apply is not a transaction.
- `Engine.Submit`, `SubmitToSpace`, and `SubmitAll` apply effects and paint (`Render` plus `RestoreInputLine`). They skip action hooks, the rate limiter, and `onAction`. `Host` refuses them while the server is draining.

## Prefer / avoid

- Prefer `Terminal` interface over direct SSH imports in loop code.
- Terminal size is `Session.setSize` from the SSH window (`sshwire.go`, `server.go`). It is render metadata, like `setLastRender`, and is not an effect. Unset size reads as 80×24. `Engine.SetTheme` is the session theme (`RenderContext.Theme`).
- A placed caret (`render.Screen.Input`) is painted inside its rectangle. No placement keeps `>` on the row under the frame. After a successful command the runtime redraws and prints a fresh `>`. `ScrollbackEcho` reprints `> line` instead. `RecentCommands` is the session transcript (cleared on scene change unless `TranscriptKeep`; `TranscriptOff` records nothing). Editor up/down recall clears when the scene changes. `Keys` and a placement fail in `renderScene`.
- Input path: byte→line (`terminalinput.go`), transport line filter, then `InputEligible` in `dispatcher.go`. Wire already sitting in the read buffer (Kitty `KITTY_*`, long base64, escape-shaped lines, iTerm `-- BEGIN CONDUCTOR --`) is not echoed. From `KITTY_DATA_START` through `KITTY_DATA_END`, every line is dropped, including `OK`; a line that contains a space ends that window and is kept. Do not ban short tokens like `OK` at transport outside that window—register on a scene if they are commands. CSI Enter keys in `csiEnterKeys`. The runtime line editor echoes input, moves within the line (arrows, Home, End), erases (backspace, Ctrl-U, Ctrl-W), recalls history (up/down), completes the scene's input keys (Tab), redraws on Ctrl-L, treats bracketed paste as text, restores the half-typed line after resize, and moves `> ` to the next line on Enter. Ctrl-C ends the session, as does Ctrl-D at an empty prompt. Ctrl-Z and Ctrl-\ are swallowed. An unmatched line prints `no such command`. Debug `line_end` / `submit_csi` / `skipped_csi` on `input_cycle`. `Scene.Keys` reads one bound key (a single character, or `enter`, `tab`, `backspace`, `delete`, `home`, `end`, `up`, `down`, `left`, `right`) and does not prompt. `enter` is raw Return, CSI Enter (`csiEnterKeys`), or SS3 `OM`. `backspace` is BS or DEL. `tab` is Tab. `delete` is CSI `3~`. `home` and `end` are CSI/SS3 `H`/`F` and the `~` codes line mode already uses. Not a VT emulator: no mouse; extend submit keys only in `terminalinput.go`.
- Action rate limit runs after successful `ResolveInput`.
- Input-path debug (`input_protocol`, `input_ineligible`, `input_cycle`) requires logger level **debug**; action timing fields are always at info when a logger is set.
- Avoid importing site-specific packages or embedding domain rules.
