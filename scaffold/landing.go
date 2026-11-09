package scaffold

import (
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/render"
)

// Landing returns a root menu scene and navigation actions.
func Landing(destinations map[string]experience.SceneID) *experience.Builder {
	b := experience.NewBuilder()
	inputs := map[string]experience.ActionID{}
	for name, scene := range destinations {
		sceneID := scene
		actionID := experience.ActionID("go-" + name)
		b.Action(actionID, experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{experience.TransitionEffect(sceneID)}
			},
		})
		inputs[name] = actionID
	}
	b.Root("landing")
	b.Scene("landing", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
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
