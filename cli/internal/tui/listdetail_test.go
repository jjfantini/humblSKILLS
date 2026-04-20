package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

type testItem struct {
	name, filter string
}

func (t testItem) Key() string                                     { return t.name }
func (t testItem) FilterValue() string                              { return t.filter }
func (t testItem) Row(_ *ui.Theme, _ int, selected bool) string    { return t.name }
func (t testItem) Detail(_ *ui.Theme, _ int) string                { return "detail:" + t.name }

func newTestListDetail(items []Item, actions []ActionSpec) Model {
	return NewListDetail(Config{
		Theme:      ui.DefaultTheme(),
		Section:    "Test",
		Version:    "v1",
		Items:      items,
		LeftTitle:  "LEFT",
		RightTitle: "RIGHT",
		Actions:    actions,
		EmptyMsg:   "empty",
	})
}

func TestNewListDetail_AppliesDefaults(t *testing.T) {
	m := NewListDetail(Config{})
	if m.cfg.LeftTitle != "ITEMS" {
		t.Errorf("LeftTitle = %q", m.cfg.LeftTitle)
	}
	if m.cfg.RightTitle != "DETAIL" {
		t.Errorf("RightTitle = %q", m.cfg.RightTitle)
	}
	if m.cfg.Theme == nil {
		t.Error("Theme should default to DefaultTheme")
	}
}

func TestModel_QuitKey(t *testing.T) {
	m := newTestListDetail([]Item{testItem{name: "a", filter: "a"}}, nil)
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	mm := out.(Model)
	if !mm.result.Quit {
		t.Error("expected Quit")
	}
	if cmd == nil {
		t.Error("expected Quit cmd")
	}
}

func TestModel_DownArrow_MovesCursor(t *testing.T) {
	m := newTestListDetail([]Item{
		testItem{name: "a", filter: "a"},
		testItem{name: "b", filter: "b"},
	}, nil)
	m.width, m.height = 80, 24
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm := out.(Model)
	if mm.cursor != 1 {
		t.Errorf("cursor = %d, want 1", mm.cursor)
	}
}

func TestModel_UpArrow_BoundedAtZero(t *testing.T) {
	m := newTestListDetail([]Item{testItem{name: "a", filter: "a"}}, nil)
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	mm := out.(Model)
	if mm.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", mm.cursor)
	}
}

func TestModel_FilterTogglesAndMatches(t *testing.T) {
	m := newTestListDetail([]Item{
		testItem{name: "alpha", filter: "alpha"},
		testItem{name: "beta", filter: "beta"},
	}, nil)
	m.width, m.height = 80, 24

	// Press / to open filter.
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	mm := out.(Model)
	if !mm.filtOn {
		t.Fatal("expected filter on")
	}
	// Type "bet".
	for _, r := range "bet" {
		out, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		mm = out.(Model)
	}
	// Only beta survives.
	if len(mm.items) != 1 || mm.items[0].Key() != "beta" {
		t.Errorf("filter failed: %+v", mm.items)
	}

	// Enter closes filter but keeps the selection.
	out, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = out.(Model)
	if mm.filtOn {
		t.Error("enter should close filter")
	}
}

func TestModel_FilterEscClears(t *testing.T) {
	m := newTestListDetail([]Item{
		testItem{name: "alpha", filter: "alpha"},
		testItem{name: "beta", filter: "beta"},
	}, nil)
	// Open filter, type, then ESC.
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	mm := out.(Model)
	out, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	mm = out.(Model)
	out, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm = out.(Model)
	if mm.filtOn {
		t.Error("esc should close filter")
	}
	if len(mm.items) != 2 {
		t.Errorf("filter not cleared: %d items", len(mm.items))
	}
}

func TestModel_EnterFiresFirstAction(t *testing.T) {
	m := newTestListDetail(
		[]Item{testItem{name: "a", filter: "a"}},
		[]ActionSpec{{Key: "i", Label: "install", Action: "install"}},
	)
	m.width, m.height = 80, 24
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := out.(Model)
	if mm.result.Action != "install" {
		t.Errorf("action = %q", mm.result.Action)
	}
	if cmd == nil {
		t.Error("expected Quit cmd")
	}
}

func TestModel_ActionKeyRunsAction(t *testing.T) {
	m := newTestListDetail(
		[]Item{testItem{name: "a", filter: "a"}},
		[]ActionSpec{{Key: "u", Label: "update", Action: "update"}},
	)
	m.width, m.height = 80, 24
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	mm := out.(Model)
	if mm.result.Action != "update" {
		t.Errorf("action = %q", mm.result.Action)
	}
}

func TestModel_Selected_DefaultState(t *testing.T) {
	m := newTestListDetail([]Item{testItem{name: "a", filter: "a"}}, nil)
	r := m.Selected()
	if r.Quit || r.Action != "" {
		t.Errorf("initial state should be zero-valued: %+v", r)
	}
}

func TestModel_ViewMentionsTitles(t *testing.T) {
	m := newTestListDetail([]Item{testItem{name: "item-a", filter: "a"}}, nil)
	m.width, m.height = 80, 24
	m.resize()
	v := m.View()
	for _, want := range []string{"LEFT", "RIGHT", "item-a"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q:\n%s", want, v)
		}
	}
}

func TestModel_EmptyItemsShowsEmptyMsg(t *testing.T) {
	m := newTestListDetail(nil, nil)
	m.width, m.height = 80, 24
	m.resize()
	v := m.View()
	if !strings.Contains(v, "empty") {
		t.Errorf("empty msg missing:\n%s", v)
	}
}
