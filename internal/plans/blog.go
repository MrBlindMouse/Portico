package plans

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// Blog adds blog-list and blog-post scenes. Load assets named post-<id>.
func Blog() *installation.Builder {
	b := installation.NewBuilder()
	b.Scene("blog-list", components.Scene{
		Render: func(ctx components.RenderContext) string {
			return render.Box("Blog", blogListBody(ctx), 48)
		},
		Inputs: map[string]components.ActionID{
			"open": "blog-open",
		},
	})
	b.Scene("blog-post", components.Scene{
		Render: func(ctx components.RenderContext) string {
			id, _ := ctx.Local("post_id")
			s, _ := id.(string)
			if s == "" {
				return render.Box("Post", "no post selected", 48)
			}
			text, ok := ctx.Asset("post-" + s)
			if !ok {
				return render.Box("Post", fmt.Sprintf("missing asset post-%s", s), 48)
			}
			return render.Box("Post: "+s, text, 64)
		},
		Inputs: map[string]components.ActionID{
			"back": "blog-back",
		},
	})
	b.Action("blog-open", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			line := strings.TrimSpace(ctx.InputLine())
			rest := strings.TrimSpace(strings.TrimPrefix(line, "open"))
			id := strings.Fields(rest)
			if len(id) == 0 {
				return []components.Effect{
					components.SetLocalEffect("flash", "usage: open <post-id>"),
				}
			}
			return []components.Effect{
				components.SetLocalEffect("post_id", id[0]),
				components.TransitionEffect("blog-post"),
			}
		},
	})
	b.Action("blog-back", components.Action{
		Run: func(ctx components.ActionContext) []components.Effect {
			_ = ctx
			return []components.Effect{components.TransitionEffect("blog-list")}
		},
	})
	b.Root("blog-list")
	return b
}

func blogListBody(ctx components.RenderContext) string {
	var b strings.Builder
	if flash, ok := ctx.Local("flash"); ok {
		if s, ok := flash.(string); ok && s != "" {
			b.WriteString(s)
			b.WriteString("\n\n")
		}
	}
	b.WriteString("Posts:\n")
	var ids []string
	for name := range ctx.Assets() {
		rest, ok := strings.CutPrefix(name, "post-")
		if ok && rest != "" {
			ids = append(ids, rest)
		}
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, id := range ids {
			b.WriteString("  ")
			b.WriteString(id)
			b.WriteByte('\n')
		}
	}
	b.WriteString("\nCommands: open <id>")
	return b.String()
}
