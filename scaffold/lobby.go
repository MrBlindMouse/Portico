package scaffold

import (
	"fmt"
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
			body := "Commands: join | ready | unready | start\n\n"
			body += fmtMembers(members) + "\n"
			if ready != nil {
				body += fmt.Sprintf("Ready map: %v\n", ready)
			}
			return render.Box("Lobby", body, 56)
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
			name := ctx.Identity().String()
			return lobbySetReady(name, true)
		},
	})
	b.Action("lobby-unready", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			name := ctx.Identity().String()
			return lobbySetReady(name, false)
		},
	})
	b.Action("lobby-start", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			return []experience.Effect{
				experience.SpaceSetEffect("lobby", "started", true),
				experience.BroadcastEffect("lobby"),
			}
		},
	})
	return b
}

func lobbySetReady(name string, ready bool) []experience.Effect {
	// ponytail: LWW merge of ready map without per-session read in action — host sites should CAS for production.
	return []experience.Effect{
		experience.SpaceSetEffect("lobby", "ready:"+name, ready),
		experience.BroadcastEffect("lobby"),
	}
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
