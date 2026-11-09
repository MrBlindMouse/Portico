package scaffold

import (
	"fmt"
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/render"
)

const chatSpace = experience.SpaceID("chat")
const chatMessagesKey = "messages"

// Chat adds a multi-user chat space and scene.
func Chat() *experience.Builder {
	b := experience.NewBuilder()
	b.Space(chatSpace, experience.Space{})
	b.Action("join-chat", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			return []experience.Effect{experience.JoinSpaceEffect(chatSpace)}
		},
	})
	b.Action("send", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			line, _ := ctx.Local("line")
			text, _ := line.(string)
			if strings.TrimSpace(text) == "" {
				return nil
			}
			msg := ctx.Identity().String() + ": " + text
			return []experience.Effect{
				experience.JoinSpaceEffect(chatSpace),
				experience.AppendSpaceListEffect(chatSpace, chatMessagesKey, msg),
				experience.BroadcastEffect(chatSpace),
				experience.SetLocalEffect("line", ""),
			}
		},
	})
	b.Scene("chat", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
			var body strings.Builder
			body.WriteString("Commands: join-chat, send\n\n")
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
		Inputs: map[string]experience.ActionID{
			"join-chat": "join-chat",
			"send":      "send",
		},
	})
	return b
}
