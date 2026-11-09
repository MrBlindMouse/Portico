package scaffold

import (
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/render"
)

// Registration adds scenes for account registration when password auth is enabled.
// Works with plaintext dev stores and identity.FileAccountStore (bcrypt). Mail verification is an external extension point.
func Registration() *experience.Builder {
	b := experience.NewBuilder()
	inputs := map[string]experience.ActionID{
		"register": "register-submit",
	}
	b.Scene("register", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
			return render.Box("Register", registerBody(ctx, inputs), 48)
		},
		Inputs: inputs,
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
			if err := identity.ValidateUsername(user); err != nil {
				return []experience.Effect{
					experience.SetLocalEffect("flash", "username: letters, digits, _ or -"),
				}
			}
			return []experience.Effect{
				experience.SetLocalEffect("flash", ""),
				experience.RegisterAccountEffect(user, pass),
				experience.SetLocalEffect("flash", "account created — reconnect to sign in"),
			}
		},
	})
	return b
}

func registerBody(ctx experience.RenderContext, inputs map[string]experience.ActionID) string {
	var b strings.Builder
	if flash, ok := ctx.Local("flash"); ok {
		if s, ok := flash.(string); ok && s != "" {
			b.WriteString(s)
			b.WriteString("\n\n")
		}
	}
	cmds := make([]string, 0, len(inputs))
	for c := range inputs {
		cmds = append(cmds, c)
	}
	sort.Strings(cmds)
	b.WriteString("Commands:\n")
	for _, c := range cmds {
		switch c {
		case "register":
			b.WriteString("  register <username> <password>\n")
		case "nick":
			b.WriteString("  nick <name>\n")
		default:
			b.WriteString("  ")
			b.WriteString(c)
			b.WriteByte('\n')
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func parseTwoFields(s string) (string, string) {
	a, b, found := strings.Cut(s, " ")
	if !found {
		return strings.TrimSpace(s), ""
	}
	return strings.TrimSpace(a), strings.TrimSpace(b)
}
