package plans

import (
	"fmt"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

const guestbookSpace = components.SpaceID("guestbook")
const guestbookListKey = "entries"

// Guestbook adds a guestbook scene and space.
func Guestbook() *installation.Builder {
	b := installation.NewBuilder()
	b.Space(guestbookSpace, components.Space{})
	b.OnSessionStart(components.JoinSpaceEffect(guestbookSpace))
	b.Root("guestbook")
	b.Action("sign", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			text := restAfter(ctx.InputLine(), "sign")
			if text == "" {
				text = "was here"
			}
			return []components.Effect{
				components.JoinSpaceEffect(guestbookSpace),
				components.AppendSpaceListEffect(guestbookSpace, guestbookListKey, fmt.Sprintf("%s: %s", ctx.Identity().String(), text)),
				components.BroadcastEffect(guestbookSpace),
			}
		},
	})
	b.Scene("guestbook", components.Scene{
		Render: func(ctx components.RenderContext) string {
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
		Inputs: map[string]components.ActionID{
			"sign": "sign",
		},
	})
	return b
}
