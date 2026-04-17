// Package install orchestrates resolving, fetching, and placing skills onto
// agent platforms.
package install

import (
	"fmt"
	"sort"

	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/resolver"
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
	if reg == nil {
		return nil, fmt.Errorf("plan: nil registry")
	}

	index := make(map[string]registry.Skill, len(reg.Skills))
	for _, s := range reg.Skills {
		index[s.Name] = s
	}
	if _, ok := index[root]; !ok {
		return nil, fmt.Errorf("skill %q not in registry", root)
	}

	g := resolver.New()
	visited := make(map[string]bool)
	if err := walk(root, index, g, visited); err != nil {
		return nil, err
	}

	order, err := g.TopoSort()
	if err != nil {
		return nil, fmt.Errorf("plan: %w", err)
	}

	out := make([]Step, 0, len(order))
	for _, name := range order {
		s := index[name]
		out = append(out, Step{Skill: s, IsDep: name != root})
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
