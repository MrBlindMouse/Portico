# TODO

DESIGN.md describes the scaffold. Section 16 of that file is the list of systems a site or another program builds. They are not framework tasks. Section 15 records undecided future research only (not shipped).

What is still missing inside the scaffold:

- **Submit effects without a keypress.** DESIGN §7. A timer or other signal in the site's process should be able to hand effects to Portico. The site owns the clock. There is no bus to add.
- **Testkit assertions.** The kit already runs the real apply path in memory. Helpers for the current scene, rendered text, persisted values, and broadcasts are the rest of the testing block in DESIGN §7. A starter test is useful only when a site copies a scaffold.
