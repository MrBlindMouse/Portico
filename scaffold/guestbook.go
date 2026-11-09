package scaffold

import (
	"fmt"
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/render"
)

const guestbookSpace = experience.SpaceID("guestbook")
const guestbookListKey = "entries"

// Guestbook adds a guestbook scene and space.
func Guestbook() *experience.Builder {
	b := experience.NewBuilder()
	b.Space(guestbookSpace, experience.Space{})
	b.OnSessionStart(experience.JoinSpaceEffect(guestbookSpace))
	b.Root("guestbook")
	b.Action("sign", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			text := restAfter(ctx.InputLine(), "sign")
			if text == "" {
				text = "was here"
			}
			return []experience.Effect{
				experience.JoinSpaceEffect(guestbookSpace),
				experience.AppendSpaceListEffect(guestbookSpace, guestbookListKey, GuestbookSignMessage(ctx.Identity().String(), text)),
				experience.BroadcastEffect(guestbookSpace),
			}
		},
	})
	b.Scene("guestbook", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
			var lines []string
			for _, e := range ctx.SpaceList(guestbookSpace, guestbookListKey) {
				lines = append(lines, fmt.Sprint(e))
			}
			if len(lines) == 0 {
				lines = append(lines, "(no signatures yet)")
			}
			lines = append(lines, "", "Commands: sign <message>")
			return render.Box("Guestbook", strings.Join(lines, "\n"), 52)
		},
		Inputs: map[string]experience.ActionID{
			"sign": "sign",
		},
	})
	return b
}

// GuestbookSignMessage formats a signature line.
func GuestbookSignMessage(nick, msg string) string {
	return fmt.Sprintf("%s: %s", nick, msg)
}
