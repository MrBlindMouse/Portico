package plans

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// How to change this plan
//
// go run github.com/MrBlindMouse/Portico/cmd/portico generate copies this
// file into the site. Edit the copy.
// The call site stays in main: Landing(map[string]SceneID{ "chat": "chat", ... }).
// The map key is the destination label. The value is the scene id to open. Build does
// not inspect effects inside action functions, so a missing destination is
// an error when the visitor picks the entry, not when the site starts.
// Register that scene (usually by merging another plan) before Serve.
//
// Merge landing last if this menu should be the installation root.
// Each destination is one action id "go-<label>" returning TransitionEffect.
// The menu uses Scene.Keys: press the shown letter to open a row.
// There is no command back to the menu; the destination scene adds one if
// you want it (another TransitionEffect to "landing").
// An empty map still builds: a welcome box and no menu rows.

// Landing is the entrance menu. Root scene is "landing".
func Landing(destinations map[string]components.SceneID) *installation.Builder {
	type entry struct {
		label string
		scene components.SceneID
		key   string
	}
	rows := make([]entry, 0, len(destinations))
	for label, scene := range destinations {
		r, _ := utf8.DecodeRuneInString(label)
		rows = append(rows, entry{label: label, scene: scene, key: string(r)})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].label < rows[j].label })

	b := installation.NewBuilder()
	inputs := map[string]components.ActionID{}
	for _, row := range rows {
		if !components.SingleKeyInput(row.key) {
			panic(fmt.Sprintf("plans.Landing: illegal menu key %q for %q", row.key, row.label))
		}
		if _, dup := inputs[row.key]; dup {
			panic(fmt.Sprintf("plans.Landing: duplicate menu key %q", row.key))
		}
		sceneID := row.scene
		label := row.label
		actionID := components.ActionID("go-" + label)
		b.Action(actionID, components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				_ = ctx
				return []components.Effect{components.TransitionEffect(sceneID)}
			},
		})
		inputs[row.key] = actionID
	}
	b.Root("landing")
	sc := components.Scene{
		Render: func(ctx components.RenderContext) string {
			var body strings.Builder
			if flash, ok := ctx.Local("flash"); ok {
				if s, ok := flash.(string); ok && s != "" {
					body.WriteString(s)
					body.WriteString("\n\n")
				}
			}
			body.WriteString("Welcome, ")
			body.WriteString(ctx.Identity().String())
			body.WriteString("\n\nMenu:\n")
			for _, row := range rows {
				body.WriteString("  ")
				body.WriteString(row.key)
				body.WriteString("  ")
				body.WriteString(row.label)
				body.WriteByte('\n')
			}
			const boxW = 48
			panel := render.Box("Portico", strings.TrimRight(body.String(), "\n"), boxW)
			w, h := ctx.Size()
			lines := strings.Count(panel, "\n") + 1
			col := 0
			if w > boxW {
				col = (w - boxW) / 2
			}
			row := 0
			if h > lines {
				row = (h - lines) / 2
			}
			s := render.NewScreen(w, h)
			s.Draw(row, col, panel)
			return s.String()
		},
		Inputs: inputs,
	}
	if len(inputs) > 0 {
		sc.Keys = true
	}
	b.Scene("landing", sc)
	return b
}
