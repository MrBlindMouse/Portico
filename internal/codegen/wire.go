package codegen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/installation"
)

// FromWire emits a Go main-package skeleton from wire JSON.
func FromWire(w installation.Wire, packageName string) string {
	if packageName == "" {
		packageName = "main"
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated from wire; fill Render and Run funcs.\npackage %s\n\n", packageName)
	b.WriteString("import (\n\tportico \"github.com/MrBlindMouse/Portico\"\n\t\"github.com/MrBlindMouse/Portico/installation\"\n)\n\n")
	b.WriteString("func build() (*installation.Installation, error) {\n\treturn portico.New().\n")
	if w.Root != "" {
		fmt.Fprintf(&b, "\t\tRoot(%q).\n", w.Root)
	}
	sceneNames := make([]string, 0, len(w.Scenes))
	for name := range w.Scenes {
		sceneNames = append(sceneNames, name)
	}
	sort.Strings(sceneNames)
	for _, name := range sceneNames {
		fmt.Fprintf(&b, "\t\tScene(%q, components.Scene{\n\t\t\tRender: func(components.RenderContext) string { return %q },\n", name, name)
		sc := w.Scenes[name]
		if len(sc.Inputs) > 0 {
			b.WriteString("\t\t\tInputs: map[string]components.ActionID{\n")
			keys := make([]string, 0, len(sc.Inputs))
			for k := range sc.Inputs {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(&b, "\t\t\t\t%q: %q,\n", k, sc.Inputs[k])
			}
			b.WriteString("\t\t\t},\n")
		}
		b.WriteString("\t\t}).\n")
	}
	actionNames := append([]string(nil), w.Actions...)
	sort.Strings(actionNames)
	for _, name := range actionNames {
		fmt.Fprintf(&b, "\t\tAction(%q, components.Action{\n\t\t\tRun: func(components.ActionContext) []components.Effect { return nil },\n\t\t}).\n", name)
	}
	spaceNames := append([]string(nil), w.Spaces...)
	sort.Strings(spaceNames)
	for _, name := range spaceNames {
		fmt.Fprintf(&b, "\t\tSpace(%q, components.Space{}).\n", name)
	}
	b.WriteString("\t\tBuild()\n}\n")
	return strings.TrimSpace(b.String()) + "\n"
}
