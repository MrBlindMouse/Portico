package plans

import (
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// How to change this plan
//
// go run github.com/MrBlindMouse/Portico/cmd/portico generate copies this
// file into the site. Edit the copy.
// The call site stays in main: Landing(map[string]SceneID{ "chat": "chat", ... }).
// The map key is the command. The value is the scene id to open. Build does
// not inspect effects inside action functions, so a missing destination is
// an error when the visitor types the command, not when the site starts.
// Register that scene (usually by merging another plan) before Serve.
//
// Merge landing last if this menu should be the installation root.
// Each command is one action id "go-<command>" returning TransitionEffect.
// There is no command back to the menu; the destination scene adds one if
// you want it (another TransitionEffect to "landing").
// An empty map still builds: a welcome box and no commands.

// Landing is the entrance menu. Root scene is "landing".
func Landing(destinations map[string]components.SceneID) *installation.Builder {
	b := installation.NewBuilder()
	inputs := map[string]components.ActionID{}
	for name, scene := range destinations {
		// Copy the loop variables. Each action closes over its own command's scene.
		sceneID := scene
		command := name
		actionID := components.ActionID("go-" + command)
		b.Action(actionID, components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				_ = ctx
				return []components.Effect{components.TransitionEffect(sceneID)}
			},
		})
		inputs[command] = actionID
	}
	b.Root("landing")
	b.Scene("landing", components.Scene{
		Render: func(ctx components.RenderContext) string {
			cmds := make([]string, 0, len(inputs))
			for cmd := range inputs {
				cmds = append(cmds, cmd)
			}
			sort.Strings(cmds) // map order is random; the menu should not be
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
