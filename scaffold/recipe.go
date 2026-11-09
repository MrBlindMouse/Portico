package scaffold

import "github.com/MrBlindMouse/Portico/experience"

// RecipeCommunity merges guestbook, chat, and landing for a small community site.
func RecipeCommunity() *experience.Builder {
	return experience.NewBuilder().
		Merge(Guestbook()).
		Merge(Chat()).
		Merge(Landing(map[string]experience.SceneID{
			"guestbook": "guestbook",
			"chat":      "chat",
		}))
}

// RecipeAccountsSite merges registration, nickname, and landing.
func RecipeAccountsSite() *experience.Builder {
	return experience.NewBuilder().
		Merge(Registration()).
		Merge(NicknameCommands()).
		Merge(Landing(map[string]experience.SceneID{
			"register": "register",
		})).
		BindInput("landing", "nick", "reserve-nick").
		BindInput("register", "nick", "reserve-nick")
}
