package scaffold

import (
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
)

// NicknameCommands adds a reserve-nick action (prefix command: nick <name>).
func NicknameCommands() *experience.Builder {
	b := experience.NewBuilder()
	b.Action("reserve-nick", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			line := strings.TrimSpace(ctx.InputLine())
			name := strings.TrimSpace(strings.TrimPrefix(line, "nick"))
			if name == "" {
				return []experience.Effect{experience.SetLocalEffect("flash", "usage: nick <name>")}
			}
			return []experience.Effect{
				experience.SetLocalEffect("flash", ""),
				experience.ReserveNicknameEffect(name),
				experience.SetLocalEffect("flash", "nickname reserved"),
			}
		},
	})
	return b
}
