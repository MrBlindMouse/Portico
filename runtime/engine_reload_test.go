package runtime_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
)

func TestSetExperienceRejectsStrandedSession(t *testing.T) {
	exp, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Scene("gone", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	sess := runtime.NewSessionForTest("s1", identity.Anonymous(), "gone")
	eng := runtime.NewEngine(exp, nil)
	eng.RegisterSession(sess)
	smaller, _ := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{Render: func(experience.RenderContext) string { return "" }}).
		Build()
	err := eng.SetExperience(smaller)
	if err == nil || !strings.Contains(err.Error(), "gone") {
		t.Fatalf("err=%v", err)
	}
}
