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

func TestModel_FilterFooterShowsEscHintAndCount(t *testing.T) {
	m := newTestListDetail([]Item{
		testItem{name: "alpha", filter: "alpha"},
		testItem{name: "beta", filter: "beta"},
	}, nil)
	m.width, m.height = 80, 24
	m.resize()

	// Open the filter; footer should switch to filter-mode hints + a count.
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	mm := out.(Model)
	v := mm.View()
	for _, want := range []string{"clear filter", "apply", "2 matches"} {
		if !strings.Contains(v, want) {
			t.Errorf("filter footer missing %q:\n%s", want, v)
		}
	}

	// Narrow it to a single match -> singular noun.
	for _, r := range "alph" {
		out, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		mm = out.(Model)
	}
	if v := mm.View(); !strings.Contains(v, "1 match") {
		t.Errorf("expected singular '1 match':\n%s", v)
	}
}

func TestModel_ScrollIndicator(t *testing.T) {
	m := newTestListDetail([]Item{testItem{name: "a", filter: "a"}}, nil)
	m.width, m.height = 80, 12
	m.resize()

	// Short content: nothing overflows, so no indicator.
	if ind := m.scrollIndicator(m.cfg.Theme); ind != "" {
		t.Errorf("expected no indicator for short content, got %q", ind)
	}

	// Tall content: indicator appears with a percent and a down arrow (we're
	// at the top, so more content lies below).
	m.preview.SetContent(strings.Repeat("line\n", 100))
	m.preview.GotoTop()
	ind := m.scrollIndicator(m.cfg.Theme)
	if ind == "" {
		t.Fatal("expected scroll indicator for overflowing content")
	}
	if !strings.Contains(ind, "%") || !strings.Contains(ind, "▼") {
		t.Errorf("indicator should show percent and a down arrow: %q", ind)
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
