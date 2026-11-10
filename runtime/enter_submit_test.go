package runtime

import (
	"bufio"
	"github.com/MrBlindMouse/Portico/components"
	"io"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
)

// scriptTerminal feeds the real inputReader (same path as SSHTerminal).
type scriptTerminal struct {
	ir     *inputReader
	frames []string
}

func newScriptTerminal(r io.Reader) *scriptTerminal {
	return &scriptTerminal{ir: newInputReader(bufio.NewReader(r))}
}

func (t *scriptTerminal) WriteFrame(text string) error {
	t.frames = append(t.frames, text)
	return nil
}

func (t *scriptTerminal) ReadInput() (string, error) { return t.ir.readUserCommand() }

func (t *scriptTerminal) OnResize(func(int, int)) {}

func (t *scriptTerminal) Close() error { return nil }

func (t *scriptTerminal) SetKeys(keys []string) { t.ir.setKeys(keys) }

func TestRunLoopKeyScene(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string { return "home" },
			Keys:   true,
			Inputs: map[string]components.ActionID{"n": "goto-other", "up": "goto-other"},
		}).
		Scene("other", components.Scene{
			Render: func(ctx components.RenderContext) string { return "other" },
		}).
		Action("goto-other", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{components.TransitionEffect("other")}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := NewEngine(exp, NewApplier(nil, identity.NewMemoryStore()))
	sess := NewSessionForTest("t", identity.Anonymous(), "home")
	term := newScriptTerminal(strings.NewReader("xn\x1b[A"))
	if err := eng.RunLoop(sess, term, nil); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if len(term.frames) < 2 || !strings.Contains(term.frames[len(term.frames)-1], "other") {
		t.Fatalf("frames=%v", term.frames)
	}
}

func TestRunLoopEnterSequences(t *testing.T) {
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(ctx components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"go": "goto-other"},
		}).
		Scene("other", components.Scene{
			Render: func(ctx components.RenderContext) string { return "other" },
		}).
		Action("goto-other", components.Action{
			Run: func(ctx components.ActionContext) []components.Effect {
				return []components.Effect{components.TransitionEffect("other")}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []string{"go\x1b[27;13;13~", "go\x1b[13u"} {
		t.Run(in, func(t *testing.T) {
			eng := NewEngine(exp, NewApplier(nil, identity.NewMemoryStore()))
			sess := NewSessionForTest("t", identity.Anonymous(), "home")
			term := newScriptTerminal(strings.NewReader(in))
			if err := eng.RunLoop(sess, term, nil); err != nil && err != io.EOF {
				t.Fatal(err)
			}
			if len(term.frames) == 0 || !strings.Contains(term.frames[len(term.frames)-1], "other") {
				t.Fatalf("frames=%v", term.frames)
			}
		})
	}
}
