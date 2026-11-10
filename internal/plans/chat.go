package plans

import (
	"fmt"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

const chatSpace = components.SpaceID("chat")
const chatMessagesKey = "messages"

// Chat adds a multi-user chat space and scene.
func Chat() *installation.Builder {
	b := installation.NewBuilder()
	b.Space(chatSpace, components.Space{
		OnJoin: []components.Effect{components.BroadcastEffect(chatSpace)},
	})
	b.Action("join-chat", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			return []components.Effect{components.JoinSpaceEffect(chatSpace)}
		},
	})
	b.Action("send", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			text := restAfter(ctx.InputLine(), "send")
			if text == "" {
				return nil
			}
			msg := ctx.Identity().String() + ": " + text
			return []components.Effect{
				components.JoinSpaceEffect(chatSpace),
				components.AppendSpaceListEffect(chatSpace, chatMessagesKey, msg),
				components.BroadcastEffect(chatSpace),
			}
		},
	})
	b.Scene("chat", components.Scene{
		Render: func(ctx components.RenderContext) string {
			var body strings.Builder
			body.WriteString("Commands: join-chat, send <message>\n\n")
			for _, msg := range ctx.SpaceList(chatSpace, chatMessagesKey) {
				body.WriteString("  ")
				body.WriteString(fmt.Sprint(msg))
				body.WriteByte('\n')
			}
			members := ctx.Members(chatSpace)
			if len(members) > 0 {
				body.WriteString("\nPresent:\n")
				for _, m := range members {
					body.WriteString("  - " + m.String() + "\n")
				}
			}
			return render.Box("Chat", strings.TrimRight(body.String(), "\n"), 52)
		},
		Inputs: map[string]components.ActionID{
			"join-chat": "join-chat",
			"send":      "send",
		},
	})
	b.Root("chat")
	return b
}
