package installation

import "github.com/MrBlindMouse/Portico/components"

// LocalFlash is the session-local key plans use for one-line user feedback.
const LocalFlash = "flash"

// SetFlash returns an effect that stores a flash message for the next render.
func SetFlash(msg string) components.Effect {
	return components.SetLocalEffect(LocalFlash, msg)
}

// ClearFlash clears the flash line.
func ClearFlash() components.Effect {
	return SetFlash("")
}

// FlashPrefix returns flash text plus a blank line when a non-empty flash is set.
func FlashPrefix(ctx components.RenderContext) string {
	flash, ok := ctx.Local(LocalFlash)
	if !ok {
		return ""
	}
	s, ok := flash.(string)
	if !ok || s == "" {
		return ""
	}
	return s + "\n\n"
}