package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/ui"
)

type testItem struct {
	name, filter string
}

func (t testItem) Key() string                                  { return t.name }
func (t testItem) FilterValue() string                          { return t.filter }
func (t testItem) Row(_ *ui.Theme, _ int, selected bool) string { return t.name }
func (t testItem) Detail(_ *ui.Theme, _ int) string             { return "detail:" + t.name }

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

func TestModel_HelpOverlayTogglesAndDismisses(t *testing.T) {
	m := newTestListDetail(
		[]Item{testItem{name: "a", filter: "a"}},
		[]ActionSpec{{Key: "i", Label: "install", Action: "install"}},
	)
	m.width, m.height = 80, 24
	m.resize()

	// ? opens the overlay.
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	mm := out.(Model)
	if !mm.helpOn {
		t.Fatal("expected help overlay on after ?")
	}

	// The overlay body lists the caller's action and the general keys.
	v := mm.View()
	for _, want := range []string{"KEYBINDINGS", "NAVIGATE", "install", "toggle this help", "close help"} {
		if !strings.Contains(v, want) {
			t.Errorf("help view missing %q:\n%s", want, v)
		}
	}

	// Any key (here: a bare rune) dismisses the overlay without firing actions.
	out, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	mm = out.(Model)
	if mm.helpOn {
		t.Error("expected help overlay dismissed after key press")
	}
	if mm.result.Action != "" || mm.result.Quit {
		t.Errorf("dismiss key should not trigger action/quit: %+v", mm.result)
	}
}

func TestModel_HelpOverlayCtrlCQuits(t *testing.T) {
	m := newTestListDetail([]Item{testItem{name: "a", filter: "a"}}, nil)
	m.width, m.height = 80, 24
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	mm := out.(Model)
	out, cmd := mm.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	mm = out.(Model)
	if !mm.result.Quit {
		t.Error("ctrl+c from help should quit")
	}
	if cmd == nil {
		t.Error("expected Quit cmd from ctrl+c")
	}
}

func TestModel_FooterShowsHelpHint(t *testing.T) {
	m := newTestListDetail([]Item{testItem{name: "a", filter: "a"}}, nil)
	m.width, m.height = 80, 24
	m.resize()
	v := m.View()
	if !strings.Contains(v, "help") {
		t.Errorf("footer should advertise the ? help hint:\n%s", v)
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

// --- left-pane scroll window --------------------------------------------------

func manyItems(n int) []Item {
	out := make([]Item, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("item-%02d", i)
		out = append(out, testItem{name: name, filter: name})
	}
	return out
}

// The regression this whole window exists for: a list taller than the terminal
// used to render every row, so the frame overflowed the alt-screen and the
// header scrolled off the top with no way back.
func TestModel_LongListNeverOutgrowsTheScreen(t *testing.T) {
	m := newTestListDetail(manyItems(200), nil)
	m.width, m.height = 80, 24
	m.resize()

	for _, step := range []int{0, 50, 199} {
		m.moveCursor(step)
		v := m.View()
		if got := lipgloss.Height(v); got > m.height {
			t.Fatalf("frame is %d rows tall, terminal is %d (cursor %d)", got, m.height, m.cursor)
		}
		if !strings.Contains(v, "humblskills") {
			t.Fatalf("header scrolled out of the frame at cursor %d:\n%s", m.cursor, v)
		}
	}
}

// Scrolling has to actually reveal rows that were off-screen, and drop the ones
// it scrolled past — otherwise the window is decorative.
func TestModel_ScrollingRevealsOffscreenRows(t *testing.T) {
	m := newTestListDetail(manyItems(200), nil)
	m.width, m.height = 80, 24
	m.resize()

	if v := m.View(); !strings.Contains(v, "item-00") {
		t.Fatalf("first row should be visible at rest:\n%s", v)
	}
	m.moveCursor(199) // walk to the end
	v := m.View()
	if !strings.Contains(v, "item-199") {
		t.Errorf("last row should be visible once scrolled to the bottom:\n%s", v)
	}
	if strings.Contains(v, "item-00") {
		t.Errorf("first row should have scrolled out of the window:\n%s", v)
	}
	if !strings.Contains(v, "▲") {
		t.Errorf("expected an up-arrow indicator once rows are hidden above:\n%s", v)
	}
}

// The cursor must stay inside the window at all times: a highlight the user
// can't see is worse than no highlight.
func TestModel_CursorStaysInsideWindow(t *testing.T) {
	m := newTestListDetail(manyItems(60), nil)
	m.width, m.height = 80, 24
	m.resize()

	for i := 0; i < 59; i++ {
		m.moveCursor(1)
		start, end := m.listWindow()
		if m.cursor < start || m.cursor >= end {
			t.Fatalf("cursor %d outside window [%d,%d)", m.cursor, start, end)
		}
	}
	for i := 0; i < 59; i++ {
		m.moveCursor(-1)
		start, end := m.listWindow()
		if m.cursor < start || m.cursor >= end {
			t.Fatalf("cursor %d outside window [%d,%d) scrolling back up", m.cursor, start, end)
		}
	}
	if m.cursor != 0 || m.listOff != 0 {
		t.Errorf("expected to land back at the top, cursor=%d off=%d", m.cursor, m.listOff)
	}
}

func TestModel_WheelOverLeftPaneScrollsList(t *testing.T) {
	m := newTestListDetail(manyItems(60), nil)
	m.width, m.height = 80, 24
	m.resize()

	out, _ := m.Update(tea.MouseMsg{X: 1, Y: 5, Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	mm := out.(Model)
	if mm.cursor != wheelStep {
		t.Errorf("wheel down should advance %d rows, cursor = %d", wheelStep, mm.cursor)
	}
	out, _ = mm.Update(tea.MouseMsg{X: 1, Y: 5, Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	if mm = out.(Model); mm.cursor != 0 {
		t.Errorf("wheel up should return to the top, cursor = %d", mm.cursor)
	}
}

// Shrinking the terminal must not leave the window pointing past the last row.
func TestModel_ResizeKeepsCursorVisible(t *testing.T) {
	m := newTestListDetail(manyItems(60), nil)
	m.width, m.height = 80, 40
	m.resize()
	m.moveCursor(45)

	m.height = 12 // sudden shrink
	m.resize()
	start, end := m.listWindow()
	if m.cursor < start || m.cursor >= end {
		t.Errorf("cursor %d outside window [%d,%d) after shrink", m.cursor, start, end)
	}
	if got := lipgloss.Height(m.View()); got > m.height {
		t.Errorf("frame is %d rows, terminal is %d after shrink", got, m.height)
	}
}

// Filtering down to a short list has to release the offset, or the pane renders
// past the end of the matches.
func TestModel_FilterResetsScrollOffset(t *testing.T) {
	m := newTestListDetail(manyItems(60), nil)
	m.width, m.height = 80, 24
	m.resize()
	m.moveCursor(50)
	if m.listOff == 0 {
		t.Fatal("precondition: expected a non-zero offset before filtering")
	}

	m.filtOn = true
	m.filter.SetValue("item-07")
	m.applyFilter()
	if m.listOff != 0 {
		t.Errorf("offset should reset when the filtered list fits, got %d", m.listOff)
	}
	if v := m.View(); !strings.Contains(v, "item-07") {
		t.Errorf("the single match should be visible:\n%s", v)
	}
}

// The help sheet is a static body with no viewport of its own, so it has to fit
// the frame outright — Frame's clamp would otherwise silently eat its last rows.
func TestModel_HelpOverlayFitsAStockTerminal(t *testing.T) {
	m := newTestListDetail(manyItems(200), []ActionSpec{
		{Key: "i", Label: "install", Action: "install"},
		{Key: "u", Label: "update", Action: "update"},
		{Key: "x", Label: "remove", Action: "remove"},
	})
	m.width, m.height = 80, 24
	m.resize()
	m.helpOn = true

	v := m.View()
	if got := lipgloss.Height(v); got > m.height {
		t.Errorf("help frame is %d rows, terminal is %d", got, m.height)
	}
	for _, want := range []string{"KEYBINDINGS", "NAVIGATE", "SCROLL DETAIL", "ACTIONS", "GENERAL", "toggle this help"} {
		if !strings.Contains(v, want) {
			t.Errorf("help sheet lost %q:\n%s", want, v)
		}
	}
}
