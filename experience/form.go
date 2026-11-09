package experience

import (
	"fmt"

	"github.com/MrBlindMouse/Portico/render"
)

const (
	LocalFormStep   = "form_step"
	LocalFormFields = "form_fields"
	LocalFormValues = "form_values"
	LocalFormPrompt = "form_prompt"
)

// StartForm returns effects that begin a multi-step form in session-local state.
func StartForm(fields []string) []Effect {
	return []Effect{
		SetLocalEffect(LocalFormStep, 0),
		SetLocalEffect(LocalFormFields, fields),
		SetLocalEffect(LocalFormValues, map[string]string{}),
		SetLocalEffect(LocalFormPrompt, promptFor(fields, 0)),
	}
}

// AdvanceFormValue records one field answer and advances the step (or finishes).
func AdvanceFormValue(ctx ActionContext, value string) []Effect {
	step, _ := ctx.Local(LocalFormStep)
	fields, _ := ctx.Local(LocalFormFields)
	values, _ := ctx.Local(LocalFormValues)
	i, _ := step.(int)
	names := stringFields(fields)
	vals := map[string]string{}
	if m, ok := values.(map[string]string); ok {
		vals = m
	}
	if i >= 0 && i < len(names) {
		vals[names[i]] = value
	}
	next := i + 1
	if next >= len(names) {
		return []Effect{
			SetLocalEffect(LocalFormValues, vals),
			SetLocalEffect(LocalFormPrompt, ""),
			SetLocalEffect(LocalFormStep, -1),
		}
	}
	return []Effect{
		SetLocalEffect(LocalFormValues, vals),
		SetLocalEffect(LocalFormStep, next),
		SetLocalEffect(LocalFormPrompt, promptFor(names, next)),
	}
}

// CancelForm clears form local state.
func CancelForm() []Effect {
	return []Effect{
		SetLocalEffect(LocalFormStep, nil),
		SetLocalEffect(LocalFormFields, nil),
		SetLocalEffect(LocalFormValues, nil),
		SetLocalEffect(LocalFormPrompt, nil),
	}
}

// RenderForm draws the active form prompt in a box (read-only render helper).
func RenderForm(ctx RenderContext, title string) string {
	prompt, _ := ctx.Local(LocalFormPrompt)
	if p, ok := prompt.(string); ok && p != "" {
		return render.Box(title, p, 48)
	}
	vals, _ := ctx.Local(LocalFormValues)
	if m, ok := vals.(map[string]string); ok && len(m) > 0 {
		body := "Done:\n"
		for k, v := range m {
			body += fmt.Sprintf("  %s: %s\n", k, v)
		}
		return render.Box(title, body, 48)
	}
	return render.Box(title, "(no active form)", 48)
}

func stringFields(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func promptFor(fields []string, step int) string {
	if step < 0 || step >= len(fields) {
		return ""
	}
	return fmt.Sprintf("Enter %s:", fields[step])
}
