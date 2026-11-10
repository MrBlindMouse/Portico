package plans

import (
	"fmt"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// How to change this plan
//
// go run github.com/MrBlindMouse/Portico/cmd/portico generate copies this
// file into the site. Edit the copy. Do not import internal/plans.
//
// Shape: Space, Actions, one Scene, Root. An Inputs entry is the word the
// visitor types; the value is the action id registered above. Both must
// exist or Build fails. Render only reads the context. The action returns
// effects; it does not write session or space fields.
//
// Shared lines live in the space (AppendSpaceList + Broadcast). A nickname
// is identity (ReserveNicknameEffect), not a space key. LeaveSpace drops
// this session; OnLeave then redraws whoever is still in the room.
// Broadcast does not draw the session that just left.
//
// nick needs a fingerprint (SSH key or account). A guest has none, so the
// action returns a flash instead of an effect the store will reject.
// restAfter, in rest.go, is the text after the command word.

const chatSpace = components.SpaceID("chat")
const chatMessagesKey = "messages"

// Chat is one shared room. Root scene is "chat".
func Chat() *installation.Builder {
	b := installation.NewBuilder()
	// OnJoin / OnLeave run inside JoinSpace / LeaveSpace, after membership changes.
	b.Space(chatSpace, components.Space{
		OnJoin:  []components.Effect{components.BroadcastEffect(chatSpace)},
		OnLeave: []components.Effect{components.BroadcastEffect(chatSpace)},
	})
	b.Action("join-chat", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			return []components.Effect{
				components.SetLocalEffect("flash", ""),
				components.JoinSpaceEffect(chatSpace),
			}
		},
	})
	b.Action("leave-chat", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			if !ctx.Space().In(chatSpace) {
				return []components.Effect{components.SetLocalEffect("flash", "not in chat")}
			}
			// LeaveSpace is a no-op when the session is already out, so check first
			// if the flash should say they were not a member.
			return []components.Effect{
				components.SetLocalEffect("flash", "left chat"),
				components.LeaveSpaceEffect(chatSpace),
			}
		},
	})
	b.Action("nick", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			name := restAfter(ctx.InputLine(), "nick")
			if ctx.Identity().Fingerprint() == "" {
				return []components.Effect{components.SetLocalEffect("flash", "nickname needs a key or account")}
			}
			if err := identity.ValidateNickname(name); err != nil {
				return []components.Effect{components.SetLocalEffect("flash", "usage: nick <name>")}
			}
			// Identity.String() prefers the nickname on the next render and on later sends.
			return []components.Effect{
				components.SetLocalEffect("flash", "you are "+name),
				components.ReserveNicknameEffect(name),
			}
		},
	})
	b.Action("send", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			text := restAfter(ctx.InputLine(), "send")
			if text == "" {
				return nil // bare "send" posts nothing and is not an error
			}
			// send also joins, so a first message does not require join-chat.
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
			body.WriteString("Commands: join-chat | leave | nick <name> | send <message>\n\n")
			// flash is session-local. It is not part of the shared message list.
			if flash, ok := ctx.Local("flash"); ok {
				if s, ok := flash.(string); ok && s != "" {
					body.WriteString(s)
					body.WriteString("\n\n")
				}
			}
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
			"leave":     "leave-chat",
			"nick":      "nick",
			"send":      "send",
		},
	})
	b.Root("chat")
	return b
}
