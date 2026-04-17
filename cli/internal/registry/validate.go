package registry

import (
	"fmt"

	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
	"github.com/jjfantini/humblSKILLS/cli/internal/resolver"
)

// IssueKind classifies a dependency problem surfaced by ValidateDeps.
type IssueKind string

const (
	IssueParse       IssueKind = "parse"
	IssueUnknown     IssueKind = "unknown"
	IssueUnsatisfied IssueKind = "unsatisfied"
	IssueCycle       IssueKind = "cycle"
)

// Issue is a single dep problem. Cycle issues set Skill to the cycle path and
// leave Dep empty.
type Issue struct {
	Kind  IssueKind
	Skill string
	Dep   string
	Msg   string
}

func (i Issue) Error() string {
	if i.Kind == IssueCycle {
		return fmt.Sprintf("%s: %s", i.Kind, i.Msg)
	}
	return fmt.Sprintf("%s: skill %q dep %q: %s", i.Kind, i.Skill, i.Dep, i.Msg)
}

// ValidateDeps re-checks every skill's `requires` against the registry
// contents. CI gates this at registry-build time; doctor runs it against a
// downloaded registry as a belt-and-suspenders check so users notice if a
// broken registry somehow gets published.
func ValidateDeps(r *Registry) []Issue {
	if r == nil {
		return nil
	}

	known := make(map[string]string, len(r.Skills))
	for _, s := range r.Skills {
		known[s.Name] = s.Version
	}

	var issues []Issue
	g := resolver.New()

	for _, s := range r.Skills {
		g.AddNode(s.Name)
		for _, raw := range s.Requires {
			dep, err := frontmatter.ParseDep(raw)
			if err != nil {
				issues = append(issues, Issue{
					Kind:  IssueParse,
					Skill: s.Name,
					Dep:   raw,
					Msg:   err.Error(),
				})
				continue
			}
			registered, ok := known[dep.Name]
			if !ok {
				issues = append(issues, Issue{
					Kind:  IssueUnknown,
					Skill: s.Name,
					Dep:   raw,
					Msg:   fmt.Sprintf("no skill named %q in registry", dep.Name),
				})
				continue
			}
			if !dep.Satisfies(registered) {
				issues = append(issues, Issue{
					Kind:  IssueUnsatisfied,
					Skill: s.Name,
					Dep:   raw,
					Msg:   fmt.Sprintf("registry has %s, which doesn't satisfy %s", registered, raw),
				})
				continue
			}
			g.AddEdge(s.Name, dep.Name)
		}
	}

	if _, err := g.TopoSort(); err != nil {
		issues = append(issues, Issue{
			Kind: IssueCycle,
			Msg:  err.Error(),
		})
	}

	return issues
}
