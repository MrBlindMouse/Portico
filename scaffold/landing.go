package scaffold

import (
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/render"
)

// Landing returns a root menu scene and navigation actions.
func Landing(destinations map[string]experience.SceneID) *experience.Builder {
	b := experience.NewBuilder()
	inputs := map[string]experience.ActionID{}
	for name, scene := range destinations {
		aid := experience.ActionID("go-" + name)
		target := scene
		sceneID := target
		actionID := aid
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
			var body string
			body = "Welcome, " + ctx.Identity().String() + "\n\nCommands:\n"
			for cmd := range inputs {
				body += "  " + cmd + "\n"
			}
			return render.Box("Portico", body, 48)
		},
		Inputs: inputs,
	})
	return b
}
