package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/ui"
)

type testCollapse struct {
	k       string
	parents []string
}

func (c testCollapse) Key() string                     { return c.k }
func (c testCollapse) Row(*ui.Theme, int, bool) string { return c.k }
func (c testCollapse) Detail(*ui.Theme, int) string    { return c.k }
func (c testCollapse) FilterValue() string             { return "" }
func (c testCollapse) IsCollapsible() bool             { return true }
func (c testCollapse) CollapseKey() string             { return c.k }
func (c testCollapse) ParentCollapseKeys() []string    { return c.parents }
func (c testCollapse) RowCollapsed(th *ui.Theme, width int, selected, collapsed bool) string {
	_ = th
	_ = width
	_ = selected
	if collapsed {
		return "▸ " + c.k
	}
	return "▾ " + c.k
}

type testNestedRow struct {
	k       string
	parents []string
}

func (r testNestedRow) Key() string                     { return r.k }
func (r testNestedRow) Row(*ui.Theme, int, bool) string { return r.k }
func (r testNestedRow) Detail(*ui.Theme, int) string    { return "" }
func (r testNestedRow) FilterValue() string             { return r.k }
func (r testNestedRow) ParentCollapseKeys() []string    { return r.parents }

func TestListDetail_CollapsibleNavAndToggle(t *testing.T) {
	m := NewListDetail(Config{Items: []Item{
		testHeader{"reg"},
		testCollapse{k: "cat"},
		testNestedRow{k: "a", parents: []string{"cat"}},
		testCollapse{k: "role", parents: []string{"cat"}},
		testNestedRow{k: "b", parents: []string{"cat", "role"}},
	}})

	// Registry header skipped; land on category.
	if m.cursor != 1 || m.items[m.cursor].Key() != "cat" {
		t.Fatalf("cursor=%d key=%q, want category", m.cursor, m.items[m.cursor].Key())
	}
	// Category is navigable (not skipped).
	if got := m.nextSelectable(1); got != 2 {
		t.Fatalf("next from cat = %d, want 2 (skill a)", got)
	}

	// Collapse category → hide a, role, b.
	m.cursor = 1
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	keys := visibleKeys(m)
	want := []string{"reg", "cat"}
	if len(keys) != len(want) {
		t.Fatalf("after collapse visible = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("after collapse visible = %v, want %v", keys, want)
		}
	}
	if m.cursor != 1 || m.items[m.cursor].Key() != "cat" {
		t.Fatalf("cursor should stay on category, got %d %q", m.cursor, m.items[m.cursor].Key())
	}

	// Expand again.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	keys = visibleKeys(m)
	want = []string{"reg", "cat", "a", "role", "b"}
	if len(keys) != len(want) {
		t.Fatalf("after expand visible = %v, want %v", keys, want)
	}
}

func TestListDetail_CollapseRoleOnly(t *testing.T) {
	m := NewListDetail(Config{Items: []Item{
		testCollapse{k: "cat"},
		testNestedRow{k: "a", parents: []string{"cat"}},
		testCollapse{k: "role", parents: []string{"cat"}},
		testNestedRow{k: "b", parents: []string{"cat", "role"}},
	}})
	// Move to role header.
	m.cursor = 2
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	keys := visibleKeys(m)
	want := []string{"cat", "a", "role"}
	if len(keys) != len(want) {
		t.Fatalf("visible = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("visible = %v, want %v", keys, want)
		}
	}
}

func TestListDetail_FilterHidesCollapsibleHeaders(t *testing.T) {
	m := NewListDetail(Config{Items: []Item{
		testCollapse{k: "cat"},
		testNestedRow{k: "alpha", parents: []string{"cat"}},
		testNestedRow{k: "beta", parents: []string{"cat"}},
	}})
	m.filter.SetValue("alp")
	m.rebuildVisible()
	keys := visibleKeys(m)
	if len(keys) != 1 || keys[0] != "alpha" {
		t.Fatalf("filter visible = %v, want [alpha]", keys)
	}
}

func visibleKeys(m Model) []string {
	out := make([]string, 0, len(m.items))
	for _, it := range m.items {
		out = append(out, it.Key())
	}
	return out
}
