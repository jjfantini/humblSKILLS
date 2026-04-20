package resolver_test

import (
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/resolver"
)

// FuzzTopoSort feeds arbitrary edge sequences into the graph and
// ensures TopoSort never panics. Cycles are a legitimate outcome
// surfaced as *CycleError; any other error means the caller has a
// bigger bug (none today — the function only returns *CycleError or nil).
func FuzzTopoSort(f *testing.F) {
	// Seed with well-formed and malformed graphs.
	seeds := []string{
		"",
		"a b",          // edge a -> b
		"a b\nb c",     // chain
		"a b\nb a",     // cycle
		"a a",          // self-loop
		"a b\nc d",     // disconnected
		"a b\nb c\nc a", // longer cycle
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		g := resolver.New()
		for _, line := range strings.Split(s, "\n") {
			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}
			if len(parts) == 1 {
				g.AddNode(parts[0])
				continue
			}
			g.AddEdge(parts[0], parts[1])
		}
		_, _ = g.TopoSort()
	})
}
