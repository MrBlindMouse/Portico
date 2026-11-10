package plans

import (
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// Lobby adds a multiplayer-ready lobby space.
func Lobby() *installation.Builder {
	b := installation.NewBuilder()
	b.Space("lobby", components.Space{
		OnJoin: []components.Effect{
			components.BroadcastEffect("lobby"),
		},
	})
	b.Scene("lobby", components.Scene{
		Render: func(ctx components.RenderContext) string {
			members := ctx.Members("lobby")
			ready, _ := ctx.SpaceLocal("lobby", "ready")
			var body strings.Builder
			body.WriteString("Commands: join | ready | unready | start\n\n")
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
			"ready":   "lobby-ready",
			"unready": "lobby-unready",
			"start":   "lobby-start",
		},
	})
	b.Action("lobby-join", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			return []components.Effect{components.JoinSpaceEffect("lobby")}
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

func lobbySetReady(ctx components.ActionContext, ready bool) []components.Effect {
	// ponytail: last write of the whole ready map wins; concurrent ready clicks can drop one update. Keys are Identity.String, so two sessions with the same SSH login (or both "guest") share one entry. Upgrade path is SpaceCAS plus a stable per-session id.
	m := readyMap(ctx)
	m[ctx.Identity().String()] = ready
	return []components.Effect{
		components.SetLocalEffect("flash", ""),
		components.SpaceSetEffect("lobby", "ready", m),
		components.BroadcastEffect("lobby"),
	}
}

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
