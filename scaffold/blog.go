package scaffold

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/render"
)

// Blog adds blog-list and blog-post scenes. Load assets named post-<id>.
func Blog() *experience.Builder {
	b := experience.NewBuilder()
	b.Scene("blog-list", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
			return render.Box("Blog", blogListBody(ctx), 48)
		},
		Inputs: map[string]experience.ActionID{
			"open": "blog-open",
		},
	})
	b.Scene("blog-post", experience.Scene{
		Render: func(ctx experience.RenderContext) string {
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
		Inputs: map[string]experience.ActionID{
			"back": "blog-back",
		},
	})
	b.Action("blog-open", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			line := strings.TrimSpace(ctx.InputLine())
			rest := strings.TrimSpace(strings.TrimPrefix(line, "open"))
			id := strings.Fields(rest)
			if len(id) == 0 {
				return []experience.Effect{
					experience.SetLocalEffect("flash", "usage: open <post-id>"),
				}
			}
			return []experience.Effect{
				experience.SetLocalEffect("post_id", id[0]),
				experience.TransitionEffect("blog-post"),
			}
		},
	})
	b.Action("blog-back", experience.Action{
		Run: func(ctx experience.ActionContext) []experience.Effect {
			_ = ctx
			return []experience.Effect{experience.TransitionEffect("blog-list")}
		},
	})
	return b
}

func blogListBody(ctx experience.RenderContext) string {
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
