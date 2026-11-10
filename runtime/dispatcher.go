package runtime

import (
	"fmt"
	"github.com/MrBlindMouse/Portico/components"
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/installation"
)

// InputEligible reports whether input could dispatch in scene (exact key or first token).
// Lines that fail this check are not user commands for the current scene—for example
// terminal acknowledgements such as "OK" when the scene only defines greet/go/….
func InputEligible(exp *installation.Installation, scene components.SceneID, input string) bool {
	sc, ok := exp.Scene(scene)
	if !ok {
		return false
	}
	if _, ok := sc.Inputs[input]; ok {
		return true
	}
	if head, _, found := strings.Cut(input, " "); found {
		if _, ok := sc.Inputs[head]; ok {
			return true
		}
	}
	return false
}

// ResolveInput maps a raw input token to an action id for the current scene.
func ResolveInput(exp *installation.Installation, scene components.SceneID, input string) (components.ActionID, error) {
	if !InputEligible(exp, scene, input) {
		return "", fmt.Errorf("portico: no action for input %q in scene %q", input, scene)
	}
	sc, _ := exp.Scene(scene)
	if aid, ok := sc.Inputs[input]; ok {
		return aid, nil
	}
	head, _, _ := strings.Cut(input, " ")
	return sc.Inputs[head], nil
}

// completeSceneInput extends a command prefix using the scene's input keys.
// An empty prefix, a line that already has an argument, or an ambiguous key
// with no longer shared prefix is returned unchanged.
func completeSceneInput(exp *installation.Installation, scene components.SceneID, line string) string {
	if exp == nil || line == "" || strings.ContainsRune(line, ' ') {
		return line
	}
	sc, ok := exp.Scene(scene)
	if !ok {
		return line
	}
	var matches []string
	for key := range sc.Inputs {
		if strings.HasPrefix(key, line) {
			matches = append(matches, key)
		}
	}
	if len(matches) == 0 {
		return line
	}
	sort.Strings(matches)
	if len(matches) == 1 {
		return matches[0]
	}
	prefix := matches[0]
	for _, m := range matches[1:] {
		for !strings.HasPrefix(m, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" || len(prefix) < len(line) {
				return line
			}
		}
	}
	if len(prefix) <= len(line) {
		return line
	}
	return prefix
}
