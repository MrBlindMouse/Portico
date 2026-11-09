package scaffold

import (
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/render"
)

// Form adds a two-field demo form (name, message).
func Form() *experience.Builder {
	b := experience.NewBuilder()
	b.Scene("form", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
			return experience.RenderForm(ctx, "Contact")
		},
		Inputs: map[string]experience.ActionID{
			"start": "form-start",
			"next":  "form-next",
			"cancel": "form-cancel",
		},
	})
	b.Action("form-start", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			return experience.StartForm([]string{"name", "message"})
		},
	})
	b.Action("form-next", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			line := strings.TrimSpace(ctx.InputLine())
			rest := strings.TrimSpace(strings.TrimPrefix(line, "next"))
			if rest == "" {
				return []experience.Effect{
					experience.SetLocalEffect("flash", "usage: next <value>"),
				}
			}
			return experience.AdvanceFormValue(ctx, rest)
		},
	})
	b.Action("form-cancel", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			_ = ctx
			return experience.CancelForm()
		},
	})
	b.Scene("form-menu", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
			_ = ctx
			return render.Box("Form demo", "Commands:\n  start — begin form\n  (in form scene) next <value>\n  cancel", 48)
		},
		Inputs: map[string]experience.ActionID{"go-form": "go-form"},
	})
	b.Action("go-form", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			_ = ctx
			return []experience.Effect{experience.TransitionEffect("form")}
		},
	})
	return b
}
