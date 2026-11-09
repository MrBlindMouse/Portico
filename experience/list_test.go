package experience_test

import (
	"strings"
	"testing"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/render"
)

type listMem struct {
	v  map[string]any
	th render.Theme
}

func (m listMem) Identity() identity.Identity { return identity.Anonymous() }
func (m listMem) Local(k string) (any, bool) {
	v, ok := m.v[k]
	return v, ok
}
func (m listMem) Asset(string) (string, bool)                       { return "", false }
func (m listMem) Assets() map[string]string                         { return nil }
func (m listMem) Persisted(string) (any, bool)                      { return nil, false }
func (m listMem) Members(experience.SpaceID) []identity.Identity    { return nil }
func (m listMem) SpaceLocal(experience.SpaceID, string) (any, bool) { return nil, false }
func (m listMem) SpaceList(experience.SpaceID, string) []any        { return nil }
func (m listMem) Size() (int, int)                                  { return 80, 24 }
func (m listMem) Theme() render.Theme                               { return m.th }
func (m listMem) PlaceInput(render.Input)                           {}
func (m listMem) SetLocal(string, any) experience.Effect            { return nil }
func (m listMem) Transition(experience.SceneID) experience.Effect   { return nil }
func (m listMem) InputLine() string                                 { return "" }
func (m listMem) Space() experience.SpaceContext                    { return listSpace{} }
func (m listMem) Persist(string, any) experience.Effect             { return nil }

type listSpace struct{}

func (listSpace) In(experience.SpaceID) bool                     { return false }
func (listSpace) Members(experience.SpaceID) []identity.Identity { return nil }
func (listSpace) Local(experience.SpaceID, string) (any, bool)   { return nil, false }
func (listSpace) List(experience.SpaceID, string) []any          { return nil }

func TestMoveListClamps(t *testing.T) {
	m := listMem{v: map[string]any{}, th: render.Plain}
	fx := experience.MoveList(m, "", 1, 3)
	sl, ok := fx[0].(experience.SetLocal)
	if !ok || sl.Key != experience.LocalListIndex || sl.Value != 1 {
		t.Fatalf("%#v", fx)
	}
	m.v[sl.Key] = sl.Value
	fx = experience.MoveList(m, "", 10, 3)
	sl = fx[0].(experience.SetLocal)
	if sl.Value != 2 {
		t.Fatalf("clamp %#v", sl)
	}
	m.v[experience.LocalListIndex] = "nope"
	if experience.ListIndex(m, "") != 0 {
		t.Fatal("non-int index")
	}
}

func TestRenderListMarksSelection(t *testing.T) {
	m := listMem{
		v:  map[string]any{experience.LocalListIndex: 4},
		th: render.Plain,
	}
	got := experience.RenderList(m, "", []string{"a", "b", "c", "d", "e"}, 2)
	if got != "  d\n> e" {
		t.Fatalf("%q", got)
	}
	m.th = render.Default
	got = experience.RenderList(m, "", []string{"a", "b", "c", "d", "e"}, 2)
	if !strings.Contains(got, "> e") || !strings.Contains(got, "\x1b[7m") {
		t.Fatalf("%q", got)
	}
}
