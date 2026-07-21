package tui

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

type testRow struct{ k string }

func (r testRow) Key() string                     { return r.k }
func (r testRow) Row(*ui.Theme, int, bool) string { return r.k }
func (r testRow) Detail(*ui.Theme, int) string    { return "" }
func (r testRow) FilterValue() string             { return r.k }

type testHeader struct{ k string }

func (h testHeader) Key() string                     { return h.k }
func (h testHeader) Row(*ui.Theme, int, bool) string { return h.k }
func (h testHeader) Detail(*ui.Theme, int) string    { return "" }
func (h testHeader) FilterValue() string             { return "" }
func (h testHeader) IsHeader() bool                  { return true }

// TestListDetail_HeaderSkip verifies non-selectable header rows are skipped by
// cursor seeding and navigation.
func TestListDetail_HeaderSkip(t *testing.T) {
	m := NewListDetail(Config{Items: []Item{
		testHeader{"h1"}, testRow{"a"}, testRow{"b"}, testHeader{"h2"}, testRow{"c"},
	}})

	if m.cursor != 1 {
		t.Fatalf("firstSelectable should skip leading header: cursor=%d want 1", m.cursor)
	}
	if got := m.nextSelectable(1); got != 2 {
		t.Fatalf("next from 1 = %d, want 2", got)
	}
	if got := m.nextSelectable(2); got != 4 {
		t.Fatalf("next from 2 = %d, want 4 (skip h2)", got)
	}
	if got := m.nextSelectable(4); got != -1 {
		t.Fatalf("next from 4 = %d, want -1", got)
	}
	if got := m.prevSelectable(4); got != 2 {
		t.Fatalf("prev from 4 = %d, want 2 (skip h2)", got)
	}
	if got := m.prevSelectable(1); got != -1 {
		t.Fatalf("prev from 1 = %d, want -1 (skip h1)", got)
	}
}
