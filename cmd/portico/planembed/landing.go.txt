package plans

import (
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// Landing returns a root menu scene and navigation actions.
func Landing(destinations map[string]components.SceneID) *installation.Builder {
	b := installation.NewBuilder()
	inputs := map[string]components.ActionID{}
	for name, scene := range destinations {
		sceneID := scene
		actionID := components.ActionID("go-" + name)
		b.Action(actionID, components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{components.TransitionEffect(sceneID)}
			},
		})
		inputs[name] = actionID
	}
	b.Root("landing")
	b.Scene("landing", components.Scene{
		Render: func(ctx components.RenderContext) string {
			cmds := make([]string, 0, len(inputs))
			for cmd := range inputs {
				cmds = append(cmds, cmd)
			}
			sort.Strings(cmds)
			var body strings.Builder
			if flash, ok := ctx.Local("flash"); ok {
				if s, ok := flash.(string); ok && s != "" {
					body.WriteString(s)
					body.WriteString("\n\n")
				}
			}
			body.WriteString("Welcome, ")
			body.WriteString(ctx.Identity().String())
			body.WriteString("\n\nCommands:\n")
			for _, cmd := range cmds {
				body.WriteString("  ")
				body.WriteString(cmd)
				body.WriteByte('\n')
			}
			return render.Box("Portico", strings.TrimRight(body.String(), "\n"), 48)
		},
		Inputs: inputs,
	})
	return b
}
