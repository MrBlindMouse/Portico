package runtime_test

import (
	"github.com/MrBlindMouse/Portico/components"
	"testing"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
)

func transcriptExp(t *testing.T) *installation.Installation {
	t.Helper()
	exp, err := installation.NewBuilder().
		Root("home").
		Scene("home", components.Scene{
			Render: func(components.RenderContext) string { return "home" },
			Inputs: map[string]components.ActionID{"stay": "stay", "go": "goto"},
		}).
		Scene("other", components.Scene{
			Render: func(components.RenderContext) string { return "other" },
			Inputs: map[string]components.ActionID{"ping": "ping"},
		}).
		Action("stay", components.Action{Run: func(components.ActionContext) []components.Effect { return nil }}).
		Action("ping", components.Action{Run: func(components.ActionContext) []components.Effect { return nil }}).
		Action("goto", components.Action{
			Run: func(components.ActionContext) []components.Effect {
				return []components.Effect{components.TransitionEffect("other")}
			},
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	return exp
}

func TestTranscriptClearsOnSceneChange(t *testing.T) {
	h := testkit.New(transcriptExp(t))
	sess := h.Session("t", "home")
	if err := h.Engine.HandleInput(sess, "stay"); err != nil {
		t.Fatal(err)
	}
	if got := sess.RecentCommands(0); len(got) != 1 || got[0] != "stay" {
		t.Fatalf("commands=%v", got)
	}
	if err := h.Engine.HandleInput(sess, "go"); err != nil {
		t.Fatal(err)
	}
	if sess.CurrentScene() != "other" {
		t.Fatalf("scene=%s", sess.CurrentScene())
	}
	if got := sess.RecentCommands(0); len(got) != 0 {
		t.Fatalf("transition should drop the transcript, got %v", got)
	}
	if err := h.Engine.HandleInput(sess, "ping"); err != nil {
		t.Fatal(err)
	}
	if got := sess.RecentCommands(0); len(got) != 1 || got[0] != "ping" {
		t.Fatalf("commands=%v", got)
	}
}

func TestTranscriptKeepAndOff(t *testing.T) {
	h := testkit.New(transcriptExp(t))
	h.Engine.SetInputPolicy(false, runtime.TranscriptKeep)
	sess := h.Session("t", "home")
	if err := h.Engine.HandleInput(sess, "stay"); err != nil {
		t.Fatal(err)
	}
	if err := h.Engine.HandleInput(sess, "go"); err != nil {
		t.Fatal(err)
	}
	got := sess.RecentCommands(0)
	if len(got) != 2 || got[0] != "stay" || got[1] != "go" {
		t.Fatalf("keep=%v", got)
	}

	h.Engine.SetInputPolicy(false, runtime.TranscriptOff)
	if err := h.Engine.HandleInput(sess, "ping"); err != nil {
		t.Fatal(err)
	}
	if got := sess.RecentCommands(0); len(got) != 2 {
		t.Fatalf("off should not append, got %v", got)
	}
}
