package resolver

import (
	"errors"
	"reflect"
	"testing"
)

func TestTopoSort_Empty(t *testing.T) {
	g := New()
	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(order) != 0 {
		t.Errorf("expected empty order, got %v", order)
	}
}

func TestTopoSort_SingleNode(t *testing.T) {
	g := New()
	g.AddNode("a")
	order, err := g.TopoSort()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"a"}) {
		t.Errorf("got %v", order)
	}
}

func TestTopoSort_Linear(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	order, err := g.TopoSort()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"c", "b", "a"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("got %v, want %v", order, want)
	}
}

func TestTopoSort_Diamond(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "d")
	order, err := g.TopoSort()
	if err != nil {
		t.Fatal(err)
	}
	pos := make(map[string]int, len(order))
	for i, n := range order {
		pos[n] = i
	}
	if pos["d"] > pos["b"] || pos["d"] > pos["c"] {
		t.Errorf("d should precede b,c: %v", order)
	}
	if pos["b"] > pos["a"] || pos["c"] > pos["a"] {
		t.Errorf("b,c should precede a: %v", order)
	}
}

func TestTopoSort_Cycle(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	g.AddEdge("c", "a")
	_, err := g.TopoSort()
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("expected CycleError, got %T: %v", err, err)
	}
	if len(ce.Path) < 2 || ce.Path[0] != ce.Path[len(ce.Path)-1] {
		t.Errorf("cycle path malformed: %v", ce.Path)
	}
}

func TestTopoSort_SelfLoop(t *testing.T) {
	g := New()
	g.AddEdge("a", "a")
	_, err := g.TopoSort()
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("expected CycleError for self-loop, got %v", err)
	}
}

func TestTopoSort_Disconnected(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddNode("x")
	order, err := g.TopoSort()
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 3 {
		t.Errorf("expected 3 nodes, got %v", order)
	}
}

func TestTopoSort_UnknownDepIgnored(t *testing.T) {
	g := New()
	g.AddNode("a")
	// Manually register a dep to a node that was never AddNode'd; the deps
	// field won't add it because AddEdge always inserts both endpoints — so
	// build it by hand to simulate malformed input.
	g.deps["a"] = []string{"ghost"}
	order, err := g.TopoSort()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"a"}) {
		t.Errorf("got %v", order)
	}
}

func TestTopoSort_Deterministic(t *testing.T) {
	// Run repeatedly; output must be byte-for-byte identical.
	make := func() []string {
		g := New()
		g.AddEdge("a", "b")
		g.AddEdge("a", "c")
		g.AddEdge("x", "y")
		g.AddNode("z")
		out, err := g.TopoSort()
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	first := make()
	for i := 0; i < 20; i++ {
		if !reflect.DeepEqual(first, make()) {
			t.Fatalf("non-deterministic output")
		}
	}
}
