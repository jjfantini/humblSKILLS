// Package install orchestrates resolving, fetching, and placing skills onto
// agent platforms.
package install

import (
	"fmt"
	"sort"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/frontmatter"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/resolver"
)

// Step is one entry in an install plan: a skill that needs fetching and
// placing, along with whether it was requested directly or pulled in as a
// transitive dependency.
type Step struct {
	Skill registry.Skill
	IsDep bool
}

// Plan returns the topo-sorted list of skills required to install `root`,
// with dependencies first and the root itself last. Missing or unsatisfiable
// deps surface as errors.
func Plan(reg *registry.Registry, root string) ([]Step, error) {
	return PlanAll(reg, []string{root})
}

// PlanAll is Plan for several roots at once: one topo-sorted plan covering
// every root and its transitive deps, deps before dependents.
//
// Walking all roots into a single graph is what makes a batch install correct
// rather than just convenient. Planning each root separately and concatenating
// would fetch a shared dep once per root that needs it, and could order a dep
// after a dependent that a later root pulled in — one graph plus one topo sort
// gives each skill exactly one Step in a globally valid order.
//
// IsDep marks a skill that no caller asked for by name: a root stays a root
// even when another root also depends on it.
func PlanAll(reg *registry.Registry, roots []string) ([]Step, error) {
	if reg == nil {
		return nil, fmt.Errorf("plan: nil registry")
	}
	if len(roots) == 0 {
		return nil, nil
	}

	index := make(map[string]registry.Skill, len(reg.Skills))
	for _, s := range reg.Skills {
		index[s.Name] = s
	}

	isRoot := make(map[string]bool, len(roots))
	g := resolver.New()
	visited := make(map[string]bool)
	for _, root := range roots {
		if _, ok := index[root]; !ok {
			return nil, fmt.Errorf("skill %q not in registry", root)
		}
		// A name repeated by the caller is not an error, just already handled.
		if isRoot[root] {
			continue
		}
		isRoot[root] = true
		if err := walk(root, index, g, visited); err != nil {
			return nil, err
		}
	}

	order, err := g.TopoSort()
	if err != nil {
		return nil, fmt.Errorf("plan: %w", err)
	}

	out := make([]Step, 0, len(order))
	for _, name := range order {
		s := index[name]
		out = append(out, Step{Skill: s, IsDep: !isRoot[name]})
	}
	return out, nil
}

func walk(name string, index map[string]registry.Skill, g *resolver.Graph, visited map[string]bool) error {
	if visited[name] {
		return nil
	}
	visited[name] = true

	s, ok := index[name]
	if !ok {
		return fmt.Errorf("dep %q not in registry", name)
	}
	g.AddNode(name)

	// Sort requires for deterministic graph shape.
	reqs := append([]string(nil), s.Requires...)
	sort.Strings(reqs)
	for _, raw := range reqs {
		dep, err := frontmatter.ParseDep(raw)
		if err != nil {
			return fmt.Errorf("%s: parse dep %q: %w", name, raw, err)
		}
		depSkill, ok := index[dep.Name]
		if !ok {
			return fmt.Errorf("%s: dep %q not in registry", name, raw)
		}
		if !dep.Satisfies(depSkill.Version) {
			return fmt.Errorf("%s: dep %q unsatisfied (registry has %s)", name, raw, depSkill.Version)
		}
		g.AddEdge(name, dep.Name)
		if err := walk(dep.Name, index, g, visited); err != nil {
			return err
		}
	}
	return nil
}
