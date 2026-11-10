package plans

import (
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// How to change this plan
//
// portico generate copies this file into the site. Edit the copy.
//
// Root scene is "register". The command is "register <username> <password>".
// RegisterAccountEffect writes the account store configured in portico.Options
// (dev AccountStore, or HashedAccounts + DataDir for bcrypt files). If no
// store is set, apply returns an error and the flash in this slice may
// already have been cleared: effects are not a transaction.
// Creating an account does not sign this SSH session in. The visitor
// reconnects with that username and password. Checking mail is not this plan.
// ValidateUsername allows letters, digits, "_" and "-". The password is the
// rest of the line after the first space (it may contain spaces).

// Registration creates a password account.
func Registration() *installation.Builder {
	b := installation.NewBuilder()
	inputs := map[string]components.ActionID{
		"register": "register-submit",
	}
	b.Scene("register", components.Scene{
		Render: func(ctx components.RenderContext) string {
			return render.Box("Register", registerBody(ctx, inputs), 48)
		},
		Inputs: inputs,
	})
	b.Action("register-submit", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			line := strings.TrimSpace(ctx.InputLine())
			rest := strings.TrimSpace(strings.TrimPrefix(line, "register"))
			user, pass := parseTwoFields(rest)
			if user == "" || pass == "" {
				return []components.Effect{
					components.SetLocalEffect("flash", "usage: register <username> <password>"),
				}
			}
			if err := identity.ValidateUsername(user); err != nil {
				return []components.Effect{
					components.SetLocalEffect("flash", "username: letters, digits, _ or -"),
				}
			}
			return []components.Effect{
				components.SetLocalEffect("flash", ""),
				components.RegisterAccountEffect(user, pass),
				// Replaces the empty flash when registration succeeds.
				components.SetLocalEffect("flash", "account created — reconnect to sign in"),
			}
		},
	})
	b.Root("register")
	return b
}

func registerBody(ctx components.RenderContext, inputs map[string]components.ActionID) string {
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
		default:
			b.WriteString("  ")
			b.WriteString(c)
			b.WriteByte('\n')
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// parseTwoFields splits "user secret words" into user and "secret words".
func parseTwoFields(s string) (string, string) {
	a, b, found := strings.Cut(s, " ")
	if !found {
		return strings.TrimSpace(s), ""
	}
	return strings.TrimSpace(a), strings.TrimSpace(b)
}
