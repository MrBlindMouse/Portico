package plans

import (
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// How to change this plan
//
// portico generate copies this file into the site. Edit the copy.
//
// Shape: one Space, one Scene, actions for join / leave / ready / unready / start,
// Root "lobby". Inputs keys are the words typed. Render only reads.
//
// "ready" is one map on the space, not a field on the session. SpaceCAS
// writes it only when the stored value still equals what this action read.
// Copy the map before changing it (readyMap); the expect value must be the
// old map, not the one you just edited. A lost race returns an error from
// apply. Two sessions whose Identity.String() matches share one map entry
// (two guests, or the same SSH login).
//
// leave sets this identity's entry to false, then LeaveSpace. A dropped
// connection only runs LeaveSpace and OnLeave, so the map can still say
// "ready" for someone who is gone until the next ready or unready.
// "started" is a single flag; SpaceSet is enough there.

// Lobby is a room that starts when every current member is ready.
func Lobby() *installation.Builder {
	b := installation.NewBuilder()
	b.Space("lobby", components.Space{
		OnJoin:  []components.Effect{components.BroadcastEffect("lobby")},
		OnLeave: []components.Effect{components.BroadcastEffect("lobby")},
	})
	b.Scene("lobby", components.Scene{
		Render: func(ctx components.RenderContext) string {
			members := ctx.Members("lobby")
			ready, _ := ctx.SpaceLocal("lobby", "ready")
			var body strings.Builder
			body.WriteString("Commands: join | leave | ready | unready | start\n\n")
			if flash, ok := ctx.Local("flash"); ok {
				if s, ok := flash.(string); ok && s != "" {
					body.WriteString(s)
					body.WriteString("\n\n")
				}
			}
			body.WriteString(fmtMembers(members))
			body.WriteByte('\n')
			body.WriteString(fmtReady(ready))
			if started, ok := ctx.SpaceLocal("lobby", "started"); ok && started == true {
				body.WriteString("Started.\n")
			}
			return render.Box("Lobby", strings.TrimRight(body.String(), "\n"), 56)
		},
		Inputs: map[string]components.ActionID{
			"join":    "lobby-join",
			"leave":   "lobby-leave",
			"ready":   "lobby-ready",
			"unready": "lobby-unready",
			"start":   "lobby-start",
		},
	})
	b.Action("lobby-join", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			return []components.Effect{
				components.SetLocalEffect("flash", ""),
				components.JoinSpaceEffect("lobby"),
			}
		},
	})
	b.Action("lobby-leave", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			if !ctx.Space().In("lobby") {
				return []components.Effect{components.SetLocalEffect("flash", "not in the lobby")}
			}
			// lobbySetReady clears flash. The later SetLocal replaces it.
			// Apply runs the slice in order; it is not rolled back on a later error.
			effects := lobbySetReady(ctx, false)
			effects = append(effects,
				components.SetLocalEffect("flash", "left the lobby"),
				components.LeaveSpaceEffect("lobby"),
			)
			return effects
		},
	})
	b.Action("lobby-ready", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			if !ctx.Space().In("lobby") {
				return []components.Effect{components.SetLocalEffect("flash", "join the lobby first")}
			}
			return lobbySetReady(ctx, true)
		},
	})
	b.Action("lobby-unready", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			if !ctx.Space().In("lobby") {
				return []components.Effect{components.SetLocalEffect("flash", "join the lobby first")}
			}
			return lobbySetReady(ctx, false)
		},
	})
	b.Action("lobby-start", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			if !ctx.Space().In("lobby") {
				return []components.Effect{components.SetLocalEffect("flash", "join the lobby first")}
			}
			// Members, not the map, decide who must be ready. A leftover map
			// entry for someone who left does not block start.
			ready := readyMap(ctx)
			for _, m := range ctx.Space().Members("lobby") {
				if !ready[m.String()] {
					return []components.Effect{components.SetLocalEffect("flash", "not everyone is ready")}
				}
			}
			return []components.Effect{
				components.SetLocalEffect("flash", ""),
				components.SpaceSetEffect("lobby", "started", true),
				components.BroadcastEffect("lobby"),
			}
		},
	})
	b.Root("lobby")
	return b
}

// lobbySetReady publishes a new ready map if the stored map is unchanged.
// expect is nil when the key has never been set (CAS treats that as empty).
func lobbySetReady(ctx components.ActionContext, ready bool) []components.Effect {
	cur, ok := ctx.Space().Local("lobby", "ready")
	var expect any
	if ok {
		expect = cur
	}
	m := readyMap(ctx)
	m[ctx.Identity().String()] = ready
	return []components.Effect{
		components.SpaceCASEffect("lobby", "ready", expect, m),
		components.SetLocalEffect("flash", ""),
		components.BroadcastEffect("lobby"),
	}
}

// readyMap copies the space map so callers can edit the copy without changing expect.
func readyMap(ctx components.ActionContext) map[string]bool {
	out := map[string]bool{}
	v, ok := ctx.Space().Local("lobby", "ready")
	if !ok {
		return out
	}
	m, ok := v.(map[string]bool)
	if !ok {
		return out
	}
	for k, val := range m {
		out[k] = val
	}
	return out
}

func fmtReady(v any) string {
	m, ok := v.(map[string]bool)
	if !ok {
		return ""
	}
	var names []string
	for n, on := range m {
		if on {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "Ready: (none)\n"
	}
	return "Ready: " + strings.Join(names, ", ") + "\n"
}

func fmtMembers(m []identity.Identity) string {
	if len(m) == 0 {
		return "No one in lobby."
	}
	var names []string
	for _, id := range m {
		names = append(names, id.String())
	}
	return "Present: " + strings.Join(names, ", ")
}
