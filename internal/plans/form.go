package plans

import (
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// Form adds a two-field demo form (name, message).
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
			return installation.AdvanceFormValue(ctx, rest)
		},
	})
	b.Action("form-cancel", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			_ = ctx
			return installation.CancelForm()
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
