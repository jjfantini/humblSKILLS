package resolver_test

import (
	"fmt"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/resolver"
)

// BenchmarkTopoSort_100 tests a graph sized like a realistic skill
// dependency set (100 nodes, ~3 edges each).
func BenchmarkTopoSort_100(b *testing.B) {
	g := resolver.New()
	for i := 0; i < 100; i++ {
		for j := 1; j <= 3 && i-j >= 0; j++ {
			g.AddEdge(fmt.Sprintf("n%d", i), fmt.Sprintf("n%d", i-j))
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := g.TopoSort(); err != nil {
			b.Fatal(err)
		}
	}
}
