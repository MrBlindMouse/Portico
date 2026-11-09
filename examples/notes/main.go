// Copied from scaffold/guestbook.go and edited in this module.
// The command is note, and each line is stored as "noted: <text>".
// External sites do this in their own repo; they do not edit scaffold/.
package main

import (
	"fmt"
	"log"
	"strings"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/render"
)

func main() {
	exp, err := build()
	if err != nil {
		log.Fatal(err)
	}
	if err := portico.ServeDev(exp, portico.Options{DataDir: ".portico-data"}); err != nil {
		log.Fatal(err)
	}
}

func build() (*experience.Experience, error) {
	const space = experience.SpaceID("notes")
	const listKey = "entries"
	return portico.New().
		Root("notes").
		Space(space, experience.Space{}).
		Action("note", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				text := restAfter(ctx.InputLine(), "note")
				if text == "" {
					text = "was here"
				}
				line := fmt.Sprintf("noted: %s — %s", ctx.Identity().String(), text)
				return []experience.Effect{
					experience.JoinSpaceEffect(space),
					experience.AppendSpaceListEffect(space, listKey, line),
					experience.BroadcastEffect(space),
				}
			},
		}).
		Scene("notes", experience.Scene{
			Render: func(ctx experience.RenderContext) string {
				var lines []string
				for _, e := range ctx.SpaceList(space, listKey) {
					lines = append(lines, fmt.Sprint(e))
				}
				if len(lines) == 0 {
					lines = append(lines, "(nothing noted yet)")
				}
				lines = append(lines, "", "Commands: note <message>")
				return render.Box("Notes", strings.Join(lines, "\n"), 52)
			},
			Inputs: map[string]experience.ActionID{"note": "note"},
		}).
		Build()
}

func restAfter(line, cmd string) string {
	line = strings.TrimSpace(line)
	head, rest, found := strings.Cut(line, " ")
	if head != cmd || !found {
		return ""
	}
	return strings.TrimSpace(rest)
}
