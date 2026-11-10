package plans

import (
	"fmt"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// How to change this plan
//
// portico generate copies this file into the site. Edit the copy.
//
// Shape: one Space, session hooks, one Action, one Scene, Root "guestbook".
// The visitor types "sign"; that Inputs key must name an action Build can see.
// Render only reads. The action returns effects.
//
// The board is a space list (everyone sees the same lines). Persist is the
// signer's own last line in the identity store, which survives reconnect.
// Skip Persist when Fingerprint() is empty: the store rejects guests, and a
// failed effect still leaves earlier effects in the same slice applied.
// OnSessionStart joins; OnSessionEnd leaves. A disconnect also leaves every
// space the session was in, so the end hook is the explicit pair for the start hook.
// restAfter, in rest.go, is the text after "sign".

const guestbookSpace = components.SpaceID("guestbook")
const guestbookListKey = "entries"

// Guestbook is one shared board. Root scene is "guestbook".
func Guestbook() *installation.Builder {
	b := installation.NewBuilder()
	b.Space(guestbookSpace, components.Space{
		OnLeave: []components.Effect{components.BroadcastEffect(guestbookSpace)},
	})
	b.OnSessionStart(components.JoinSpaceEffect(guestbookSpace))
	b.OnSessionEnd(components.LeaveSpaceEffect(guestbookSpace))
	b.Root("guestbook")
	b.Action("sign", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			text := restAfter(ctx.InputLine(), "sign")
			if text == "" {
				text = "was here" // "sign" with no text still posts
			}
			line := fmt.Sprintf("%s: %s", ctx.Identity().String(), text)
			effects := []components.Effect{
				components.JoinSpaceEffect(guestbookSpace), // safe if already joined
				components.AppendSpaceListEffect(guestbookSpace, guestbookListKey, line),
				components.BroadcastEffect(guestbookSpace),
			}
			if ctx.Identity().Fingerprint() != "" {
				effects = append(effects, components.PersistEffect("last_signature", line))
			}
			return effects
		},
	})
	b.Scene("guestbook", components.Scene{
		Render: func(ctx components.RenderContext) string {
			var lines []string
			// Persisted is per identity. The list below is the shared board.
			if last, ok := ctx.Persisted("last_signature"); ok {
				lines = append(lines, "Your last: "+fmt.Sprint(last), "")
			}
			entries := ctx.SpaceList(guestbookSpace, guestbookListKey)
			for _, e := range entries {
				lines = append(lines, fmt.Sprint(e))
			}
			if len(entries) == 0 {
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
