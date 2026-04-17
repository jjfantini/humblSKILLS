// Package resolver provides a small DAG + topological sort used to validate
// skill dependency graphs and to order installs so deps are installed first.
package resolver

import (
	"sort"
	"strings"
)

// Graph is a directed graph of skill dependencies. An edge from A -> B means
// "A depends on B" (so B must be installed before A).
type Graph struct {
	nodes map[string]struct{}
	deps  map[string][]string
}

// New returns an empty graph.
func New() *Graph {
	return &Graph{
		nodes: map[string]struct{}{},
		deps:  map[string][]string{},
	}
}

// AddNode inserts a node, no-op if it already exists.
func (g *Graph) AddNode(n string) {
	g.nodes[n] = struct{}{}
}

// AddEdge records that `from` depends on `to`. Both nodes are inserted if
// they don't already exist.
func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	g.deps[from] = append(g.deps[from], to)
}

// CycleError is returned by TopoSort when the graph contains a cycle. Path is
// the offending cycle with the starting node repeated at the end, e.g.
// ["a", "b", "c", "a"].
type CycleError struct {
	Path []string
}

func (e *CycleError) Error() string {
	return "dependency cycle: " + strings.Join(e.Path, " -> ")
}

// TopoSort returns nodes in deps-first install order. If the graph contains a
// cycle, TopoSort returns a *CycleError. Deps referencing unknown nodes are
// ignored — the caller is expected to validate those separately.
func (g *Graph) TopoSort() ([]string, error) {
	names := make([]string, 0, len(g.nodes))
	for n := range g.nodes {
		names = append(names, n)
	}
	sort.Strings(names)

	const (
		white = 0
		grey  = 1
		black = 2
	)
	color := make(map[string]int, len(g.nodes))
	order := make([]string, 0, len(g.nodes))
	stack := make([]string, 0, len(g.nodes))

	var visit func(n string) error
	visit = func(n string) error {
		switch color[n] {
		case black:
			return nil
		case grey:
			i := 0
			for ; i < len(stack); i++ {
				if stack[i] == n {
					break
				}
			}
			cycle := append([]string{}, stack[i:]...)
			cycle = append(cycle, n)
			return &CycleError{Path: cycle}
		}

		color[n] = grey
		stack = append(stack, n)

		deps := append([]string(nil), g.deps[n]...)
		sort.Strings(deps)
		for _, d := range deps {
			if _, known := g.nodes[d]; !known {
				continue
			}
			if err := visit(d); err != nil {
				return err
			}
		}

		stack = stack[:len(stack)-1]
		color[n] = black
		order = append(order, n)
		return nil
	}

	for _, n := range names {
		if err := visit(n); err != nil {
			return nil, err
		}
	}
	return order, nil
}
