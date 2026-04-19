package install

import (
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// UpdatePlan describes one skill whose manifest entry has drifted from the
// registry. Targets lists every (platform, scope) the skill is currently
// installed onto.
type UpdatePlan struct {
	Skill       string          `json:"skill"`
	FromVersion string          `json:"from_version"`
	ToVersion   string          `json:"to_version"`
	FromSHA     string          `json:"from_source_sha,omitempty"`
	ToSHA       string          `json:"to_source_sha,omitempty"`
	FromDirSHA  string          `json:"from_dir_sha,omitempty"`
	ToDirSHA    string          `json:"to_dir_sha,omitempty"`
	Targets     []ManifestEntry `json:"targets"`
}

// ManifestEntry mirrors the subset of manifest.Installation a caller needs to
// decide what to update.
type ManifestEntry struct {
	Platform string `json:"platform"`
	Scope    string `json:"scope"`
	Path     string `json:"path"`
}

// PlanUpdates compares every manifest entry against the registry and returns
// one UpdatePlan per skill with drift. A skill whose registry entry is
// missing is skipped (there's nothing to upgrade to). When `only` is
// non-empty, only those skill names are considered.
func PlanUpdates(reg *registry.Registry, m *manifest.Manifest, only []string) []UpdatePlan {
	if reg == nil || m == nil {
		return nil
	}
	regIndex := make(map[string]registry.Skill, len(reg.Skills))
	for _, s := range reg.Skills {
		regIndex[s.Name] = s
	}

	filter := map[string]struct{}{}
	for _, n := range only {
		filter[n] = struct{}{}
	}

	// Group manifest entries by skill.
	bySkill := map[string][]manifest.Installation{}
	for _, inst := range m.Installations {
		if len(filter) > 0 {
			if _, ok := filter[inst.Skill]; !ok {
				continue
			}
		}
		bySkill[inst.Skill] = append(bySkill[inst.Skill], inst)
	}

	var out []UpdatePlan
	for name, insts := range bySkill {
		regSkill, ok := regIndex[name]
		if !ok {
			continue
		}
		// Any target that's drifted triggers an UpdatePlan for the skill.
		//
		// Drift is keyed on the per-skill signals: version and DirSHA
		// (RegistryRef). The repo-wide Source.SHA is NOT consulted — it
		// advances on every commit to the humblSKILLS repo, including
		// commits that don't touch this skill, which would flag every
		// installation as drifted after each CLI release. Source.SHA is
		// kept in the manifest purely as install-time metadata.
		drifted := false
		for _, i := range insts {
			if i.Version != regSkill.Version ||
				i.RegistryRef != regSkill.DirSHA {
				drifted = true
				break
			}
		}
		if !drifted {
			continue
		}

		first := insts[0]
		plan := UpdatePlan{
			Skill:       name,
			FromVersion: first.Version,
			ToVersion:   regSkill.Version,
			FromSHA:     first.SourceSHA,
			ToSHA:       reg.Source.SHA,
			FromDirSHA:  first.RegistryRef,
			ToDirSHA:    regSkill.DirSHA,
		}
		for _, i := range insts {
			plan.Targets = append(plan.Targets, ManifestEntry{
				Platform: i.Platform, Scope: i.Scope, Path: i.Path,
			})
		}
		out = append(out, plan)
	}
	return out
}
