package runtime

import (
	"fmt"
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
)

// ResolveInput maps a raw input token to an action id for the current scene.
func ResolveInput(exp *experience.Experience, scene experience.SceneID, input string) (experience.ActionID, error) {
	sc, ok := exp.Scene(scene)
	if !ok {
		return "", fmt.Errorf("portico: undefined scene %q", scene)
	}
	if aid, ok := sc.Inputs[input]; ok {
		return aid, nil
	}
	if head, _, found := strings.Cut(input, " "); found {
		if aid, ok := sc.Inputs[head]; ok {
			return aid, nil
		}
	}
	return "", fmt.Errorf("portico: no action for input %q in scene %q", input, scene)
}
