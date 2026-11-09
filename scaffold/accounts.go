package scaffold

import (
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/render"
)

// Registration adds scenes for account registration when password auth is enabled.
// Mail verification remains an external extension point.
func Registration() *experience.Builder {
	b := experience.NewBuilder()
	b.Scene("register", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
			flash, _ := ctx.Local("flash")
			body := "Commands:\n  register <username> <password>"
			if s, ok := flash.(string); ok && s != "" {
				body = s + "\n\n" + body
			}
			return render.Box("Register", body, 48)
		},
		Inputs: map[string]experience.ActionID{
			"register": "register-submit",
		},
	})
	b.Action("register-submit", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			line := strings.TrimSpace(ctx.InputLine())
			rest := strings.TrimSpace(strings.TrimPrefix(line, "register"))
			user, pass := parseTwoFields(rest)
			if user == "" || pass == "" {
				return []experience.Effect{
					experience.SetLocalEffect("flash", "usage: register <username> <password>"),
				}
			}
			return []experience.Effect{
				experience.RegisterAccountEffect(user, pass),
				experience.SetLocalEffect("flash", "account created"),
			}
		},
	})
	return b
}

func parseTwoFields(s string) (string, string) {
	a, b, found := strings.Cut(s, " ")
	if !found {
		return strings.TrimSpace(s), ""
	}
	return strings.TrimSpace(a), strings.TrimSpace(b)
}
