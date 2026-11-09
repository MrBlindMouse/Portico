package runtime

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
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
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string { return "home" },
			Keys:   true,
			Inputs: map[string]experience.ActionID{"n": "goto-other", "up": "goto-other"},
		}).
		Scene("other", experience.Scene{
			Render: func(ctx experience.RenderContext) string { return "other" },
		}).
		Action("goto-other", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{experience.TransitionEffect("other")}
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
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string { return "home" },
			Inputs: map[string]experience.ActionID{"go": "goto-other"},
		}).
		Scene("other", experience.Scene{
			Render: func(ctx experience.RenderContext) string { return "other" },
		}).
		Action("goto-other", experience.Action{
			Run: func(ctx experience.ActionContext) []experience.Effect {
				return []experience.Effect{experience.TransitionEffect("other")}
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
