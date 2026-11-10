package plans

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/render"
)

// How to change this plan
//
// go run github.com/MrBlindMouse/Portico/cmd/portico generate copies this
// file into the site. Edit the copy.
//
// Root is "blog-list". "open …" transitions to "blog-post"; "back" returns.
// Posts are assets whose names start with "post-". Load them with
// builder.Asset in main, or Options.AssetDir (files the runtime watches).
// Anything that does not start with "post-" is ignored by the list.
// The visitor types the id shown under Posts—the part after "post-" in the
// asset name ("open 1.md" loads "post-1.md"; "open 1" or "open post-1.md" miss).
// A missing asset is a render string, not an error effect.
//
// post_id is session-local so the post scene can read it after Transition.
// blog_last is Persist of that same id, only when the identity has a
// fingerprint, so the list can show "Last opened" on a later connection.

// Blog lists post assets and opens one.
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
				return render.Box("Post", blogPostBody("no post selected"), 48)
			}
			text, ok := ctx.Asset("post-" + s)
			if !ok {
				msg := fmt.Sprintf("no asset post-%s\n\nUse the id from Posts (e.g. open %s)", s, s)
				return render.Box("Post", blogPostBody(msg), 48)
			}
			return render.Box("Post: "+s, blogPostBody(text), 64)
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
					components.SetLocalEffect("flash", "usage: open <id from Posts list>"),
				}
			}
			effects := []components.Effect{
				components.SetLocalEffect("flash", ""),
				components.SetLocalEffect("post_id", id[0]),
				components.TransitionEffect("blog-post"),
			}
			if ctx.Identity().Fingerprint() != "" {
				effects = append(effects, components.PersistEffect("blog_last", id[0]))
			}
			return effects
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
	if last, ok := ctx.Persisted("blog_last"); ok {
		b.WriteString("Last opened: ")
		b.WriteString(fmt.Sprint(last))
		b.WriteString("\n\n")
	}
	b.WriteString("Posts:\n")
	var ids []string
	for name := range ctx.Assets() {
		rest, ok := strings.CutPrefix(name, "post-")
		if ok && rest != "" {
			ids = append(ids, rest)
		}
	}
	sort.Strings(ids) // map iteration order is random; the list should not be
	if len(ids) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, id := range ids {
			b.WriteString("  ")
			b.WriteString(id)
			b.WriteByte('\n')
		}
	}
	b.WriteString("\nCommands: open <id from Posts>")
	if len(ids) > 0 {
		b.WriteString("\n  e.g. open ")
		b.WriteString(ids[0])
	}
	return b.String()
}

func blogPostBody(main string) string {
	return strings.TrimRight(main, "\n") + "\n\nCommand: back — return to Posts"
}
