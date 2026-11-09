package scaffold

import (
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/render"
)

// Lobby adds a multiplayer-ready lobby space.
func Lobby() *experience.Builder {
	b := experience.NewBuilder()
	b.Space("lobby", experience.Space{
		OnJoin: []experience.Effect{
			experience.BroadcastEffect("lobby"),
		},
	})
	b.Scene("lobby", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
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
		Inputs: map[string]experience.ActionID{
			"join":    "lobby-join",
			"ready":   "lobby-ready",
			"unready": "lobby-unready",
			"start":   "lobby-start",
		},
	})
	b.Action("lobby-join", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			return []experience.Effect{experience.JoinSpaceEffect("lobby")}
		},
	})
	b.Action("lobby-ready", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			if !ctx.Space().In("lobby") {
				return []experience.Effect{experience.SetLocalEffect("flash", "join the lobby first")}
			}
			return lobbySetReady(ctx, true)
		},
	})
	b.Action("lobby-unready", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			if !ctx.Space().In("lobby") {
				return []experience.Effect{experience.SetLocalEffect("flash", "join the lobby first")}
			}
			return lobbySetReady(ctx, false)
		},
	})
	b.Action("lobby-start", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			if !ctx.Space().In("lobby") {
				return []experience.Effect{experience.SetLocalEffect("flash", "join the lobby first")}
			}
			ready := readyMap(ctx)
			for _, m := range ctx.Space().Members("lobby") {
				if !ready[m.String()] {
					return []experience.Effect{experience.SetLocalEffect("flash", "not everyone is ready")}
				}
			}
			return []experience.Effect{
				experience.SetLocalEffect("flash", ""),
				experience.SpaceSetEffect("lobby", "started", true),
				experience.BroadcastEffect("lobby"),
			}
		},
	})
	return b
}

func lobbySetReady(ctx experience.ActionContext, ready bool) []experience.Effect {
	// ponytail: last write of the whole ready map wins; concurrent ready clicks can drop one update. Keys are Identity.String, so two sessions with the same SSH login (or both "guest") share one entry. Upgrade path is SpaceCAS plus a stable per-session id.
	m := readyMap(ctx)
	m[ctx.Identity().String()] = ready
	return []experience.Effect{
		experience.SetLocalEffect("flash", ""),
		experience.SpaceSetEffect("lobby", "ready", m),
		experience.BroadcastEffect("lobby"),
	}
}

func readyMap(ctx experience.ActionContext) map[string]bool {
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
