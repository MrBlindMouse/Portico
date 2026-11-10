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
// Root is "form". "form-menu" is a second scene (command go-form) for a site
// that wants a door in front of the form. Merge landing and point it at
// "form" if the menu scene is unused.
//
// The walk itself is installation.StartForm / AdvanceFormValue / CancelForm.
// Those helpers only SetLocal. Local state dies when the SSH session ends.
// persistFinishedForm adds Persist for name and message when the last field
// is accepted and the identity has a fingerprint. Render shows that saved
// pair through Persisted, which reads the identity store, not session local.
// Add a field by extending the slice passed to StartForm; persistFinishedForm
// only stores keys "name" and "message" unless you extend it too.

// Form is a two-field walk: name, then message.
func Form() *installation.Builder {
	b := installation.NewBuilder()
	b.Scene("form", components.Scene{
		Render: func(ctx components.RenderContext) string {
			var body strings.Builder
			if flash, ok := ctx.Local("flash"); ok {
				if s, ok := flash.(string); ok && s != "" {
					body.WriteString(s)
					body.WriteString("\n\n")
				}
			}
			if name, ok := ctx.Persisted("form_name"); ok {
				body.WriteString("Saved name: ")
				body.WriteString(fmt.Sprint(name))
				body.WriteByte('\n')
				if msg, ok := ctx.Persisted("form_message"); ok {
					body.WriteString("Saved message: ")
					body.WriteString(fmt.Sprint(msg))
					body.WriteByte('\n')
				}
				body.WriteByte('\n')
			}
			body.WriteString("Commands: start | next <value> | cancel\n\n")
			body.WriteString(installation.FormText(ctx))
			return render.Box("Contact", strings.TrimRight(body.String(), "\n"), 48)
		},
		Inputs: map[string]components.ActionID{
			"start":  "form-start",
			"next":   "form-next",
			"cancel": "form-cancel",
		},
	})
	b.Action("form-start", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			return installation.StartForm([]string{"name", "message"})
		},
	})
	b.Action("form-next", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			line := strings.TrimSpace(ctx.InputLine())
			rest := strings.TrimSpace(strings.TrimPrefix(line, "next"))
			if rest == "" {
				return []components.Effect{
					components.SetLocalEffect("flash", "usage: next <value>"),
				}
			}
			// "next hi there" keeps the rest of the line, including spaces.
			return persistFinishedForm(ctx, installation.AdvanceFormValue(ctx, rest))
		},
	})
	b.Action("form-cancel", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			_ = ctx
			return installation.CancelForm() // clears local form keys only, not Persisted values
		},
	})
	b.Scene("form-menu", components.Scene{
		Render: func(ctx components.RenderContext) string {
			_ = ctx
			return render.Box("Form demo", "Commands:\n  go-form — open the form", 48)
		},
		Inputs: map[string]components.ActionID{"go-form": "go-form"},
	})
	b.Action("go-form", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			_ = ctx
			return []components.Effect{components.TransitionEffect("form")}
		},
	})
	b.Root("form")
	return b
}

// persistFinishedForm appends Persist effects when the helper finishes the form
// (local step -1). Guests are skipped; Persist would fail for them.
func persistFinishedForm(ctx components.ActionContext, effects []components.Effect) []components.Effect {
	if ctx.Identity().Fingerprint() == "" {
		return effects
	}
	var vals map[string]string
	finished := false
	for _, e := range effects {
		sl, ok := e.(components.SetLocal)
		if !ok {
			continue
		}
		if sl.Key == installation.LocalFormValues {
			if m, ok := sl.Value.(map[string]string); ok {
				vals = m
			}
		}
		if sl.Key == installation.LocalFormStep && sl.Value == -1 {
			finished = true
		}
	}
	if !finished || vals == nil {
		return effects
	}
	out := append([]components.Effect{}, effects...)
	if name, ok := vals["name"]; ok {
		out = append(out, components.PersistEffect("form_name", name))
	}
	if msg, ok := vals["message"]; ok {
		out = append(out, components.PersistEffect("form_message", msg))
	}
	return out
}
