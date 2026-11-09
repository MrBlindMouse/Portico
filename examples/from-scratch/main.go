// Minimal custom experience in this module (no scaffolds). Same program as README "Or write it" and portico init -style scratch.
// External sites import the module; they are not added under examples/.
package main

import (
	"log"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/experience"
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
	return portico.New().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string {
				msg := ""
				if flash, ok := ctx.Local("flash"); ok {
					if s, ok := flash.(string); ok {
						msg = s + "\n\n"
					}
				}
				return msg + "Hello, " + ctx.Identity().String() + "\n\nType: greet\n"
			},
			Inputs: map[string]experience.ActionID{"greet": "greet"},
		}).
		Action("greet", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{
					experience.SetLocalEffect("flash", "you waved"),
				}
			},
		}).
		Build()
}
