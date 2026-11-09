package runtime_test

import (
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
)

func transcriptExp(t *testing.T) *experience.Experience {
	t.Helper()
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "home" },
			Inputs: map[string]experience.ActionID{"stay": "stay", "go": "goto"},
		}).
		Scene("other", experience.Scene{
			Render: func(experience.RenderContext) string { return "other" },
			Inputs: map[string]experience.ActionID{"ping": "ping"},
		}).
		Action("stay", experience.Action{Run: func(experience.ActionContext) []experience.Effect { return nil }}).
		Action("ping", experience.Action{Run: func(experience.ActionContext) []experience.Effect { return nil }}).
		Action("goto", experience.Action{
			Run: func(experience.ActionContext) []experience.Effect {
				return []experience.Effect{experience.TransitionEffect("other")}
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
