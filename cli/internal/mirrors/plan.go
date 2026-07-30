package mirrors

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// PlanArtifacts are the files a [WritePlan] call produced.
type PlanArtifacts struct {
	Dir      string
	PlanPath string
	// IncomingPath is the current upstream written to disk so the executing
	// agent can diff it against the preserved copy at full fidelity. Embedding
	// a rendered diff in the plan would only lose information.
	IncomingPath string
	// Affected are the wiki concepts that cite the preserved copy - the blast
	// radius, computed from `sources:` frontmatter rather than guessed.
	Affected []string
}

// wikiFrontmatter is the subset of a wiki concept's frontmatter needed to
// compute blast radius. Wiki concepts use a different shape from SKILL.md.
type wikiFrontmatter struct {
	Title   string   `yaml:"title"`
	Sources []string `yaml:"sources"`
}

// Affected returns the wiki concepts under dir whose `sources:` cite the
// preserved copy, sorted. This is the blast radius of an upstream change: the
// exact set of files a re-sync has to revisit.
func Affected(dir, preserved string) ([]string, error) {
	wikiRoot := filepath.Join(dir, "references", "wiki")
	var hits []string
	err := filepath.WalkDir(wikiRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil //nolint:nilerr // a missing wiki tree is not an error
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		fm, ok := splitYAML(raw)
		if !ok {
			return nil
		}
		var w wikiFrontmatter
		if err := yaml.Unmarshal(fm, &w); err != nil {
			return nil
		}
		for _, s := range w.Sources {
			if citesPreserved(strings.TrimSpace(s), preserved) {
				rel, relErr := filepath.Rel(dir, path)
				if relErr != nil {
					rel = path
				}
				hits = append(hits, filepath.ToSlash(rel))
				break
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	sort.Strings(hits)
	return hits, nil
}

// citesPreserved reports whether a wiki concept's `sources:` entry refers to
// upstream material. When Preserved names a directory (trailing slash) every
// file beneath it is upstream, so any concept citing one of them is in the
// blast radius - not just those citing the SKILL.md baseline.
func citesPreserved(source, preserved string) bool {
	if strings.HasSuffix(preserved, "/") {
		return strings.HasPrefix(source, preserved)
	}
	return source == preserved
}

func splitYAML(raw []byte) ([]byte, bool) {
	s := string(raw)
	if !strings.HasPrefix(s, "---\n") {
		return nil, false
	}
	end := strings.Index(s[4:], "\n---")
	if end < 0 {
		return nil, false
	}
	return []byte(s[4 : 4+end]), true
}

// headingDelta reports headings added and removed between preserved and
// current. This is the cheap high-signal summary; the full diff is left to the
// two files on disk.
func headingDelta(preserved, current []byte) (added, removed []string) {
	pre, cur := headings(preserved), headings(current)
	inPre := map[string]bool{}
	for _, h := range pre {
		inPre[h] = true
	}
	inCur := map[string]bool{}
	for _, h := range cur {
		inCur[h] = true
	}
	for _, h := range cur {
		if !inPre[h] {
			added = append(added, h)
		}
	}
	for _, h := range pre {
		if !inCur[h] {
			removed = append(removed, h)
		}
	}
	return added, removed
}

// WritePlan emits a re-sync work order for a drifted mirror into outDir.
//
// The plan is deliberately a prompt, not a patch. It carries the four things a
// re-sync needs and that are otherwise reconstructed by hand every time: the
// two files to diff, the blast radius, the deltas that must survive, and the
// completion checklist.
func WritePlan(r Result, outDir string) (PlanArtifacts, error) {
	var a PlanArtifacts
	if r.Upstream == nil {
		return a, fmt.Errorf("%s: no upstream block", r.Skill)
	}
	dir := filepath.Join(outDir, r.Skill)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return a, err
	}

	affected, err := Affected(r.Dir, r.Upstream.Preserved)
	if err != nil {
		return a, err
	}

	incoming := filepath.Join(dir, "upstream-incoming"+filepath.Ext(BaselineFile(r.Upstream.Preserved)))
	if len(r.Current) > 0 {
		if err := os.WriteFile(incoming, r.Current, 0o644); err != nil {
			return a, err
		}
	}

	added, removed := headingDelta(r.Preserved, r.Current)
	planPath := filepath.Join(dir, "PLAN.md")
	if err := os.WriteFile(planPath, []byte(renderPlan(r, incoming, affected, added, removed)), 0o644); err != nil {
		return a, err
	}

	a = PlanArtifacts{Dir: dir, PlanPath: planPath, IncomingPath: incoming, Affected: affected}
	return a, nil
}

func renderPlan(r Result, incoming string, affected, added, removed []string) string {
	u := r.Upstream
	var b strings.Builder

	fmt.Fprintf(&b, "# Re-sync plan: %s\n\n", r.Skill)
	fmt.Fprintf(&b, "Status: **%s** - %s\n\n", r.Status, r.Reason)
	fmt.Fprintf(&b, "Upstream `%s` (%s), last synced %s.\n\n",
		u.Name, u.Source, orDash(u.Synced))
	b.WriteString("This is a work order for a human or agent. Detection is automated; " +
		"the distillation is not, and must not be.\n\n")

	b.WriteString("## 1. Diff these two files\n\n")
	fmt.Fprintf(&b, "- Preserved baseline: `%s`\n", filepath.Join(r.Dir, BaselineFile(u.Preserved)))
	fmt.Fprintf(&b, "- Incoming upstream:  `%s`\n\n", incoming)
	if len(added) > 0 || len(removed) > 0 {
		b.WriteString("Heading-level summary:\n\n")
		for _, h := range added {
			fmt.Fprintf(&b, "- `+` %s\n", h)
		}
		for _, h := range removed {
			fmt.Fprintf(&b, "- `-` %s\n", h)
		}
		b.WriteString("\n")
	}
	if r.Status == StatusRewritten {
		b.WriteString("> Upstream was restructured rather than edited. Prefer re-distilling the " +
			"affected concepts over patching them, and check whether new upstream material " +
			"needs concepts that do not exist yet.\n\n")
	}

	b.WriteString("## 2. Blast radius\n\n")
	if len(affected) == 0 {
		b.WriteString("_No wiki concept cites the preserved copy. Verify `sources:` is populated - " +
			"an empty blast radius usually means broken provenance, not zero impact._\n\n")
	} else {
		fmt.Fprintf(&b, "These %d concept(s) cite `%s` and must be revisited:\n\n", len(affected), u.Preserved)
		for _, f := range affected {
			fmt.Fprintf(&b, "- [ ] `%s`\n", f)
		}
		b.WriteString("\n")
	}

	b.WriteString("## 3. Declared deltas - do NOT \"fix\" these\n\n")
	if len(u.Deltas) == 0 {
		b.WriteString("_None declared._\n\n")
	} else {
		for _, d := range u.Deltas {
			fmt.Fprintf(&b, "- %s\n", d)
		}
		b.WriteString("\nThese are intentional divergences. Re-introducing upstream's version of " +
			"any of them undoes a past decision. If one is genuinely obsolete, remove it from " +
			"the `upstream.deltas` block and say why in `decisions.md`.\n\n")
	}

	b.WriteString("## 4. Completion checklist\n\n")
	b.WriteString("- [ ] Re-distill each affected concept above against the new upstream\n")
	b.WriteString("- [ ] Add concepts for upstream material that has no home yet\n")
	fmt.Fprintf(&b, "- [ ] Replace the preserved copy with the incoming file (`%s`)\n", BaselineFile(u.Preserved))
	b.WriteString("- [ ] Bump `upstream.synced` to today\n")
	b.WriteString("- [ ] Update `upstream.deltas` if the intentional differences changed\n")
	b.WriteString("- [ ] Regenerate `ATTRIBUTION.md` from the block\n")
	b.WriteString("- [ ] Bump `metadata.version`\n")
	b.WriteString("- [ ] `bash scripts/lint.sh` exits 0\n")
	b.WriteString("- [ ] Append a `log.md` entry; record non-obvious calls in `decisions.md`\n\n")
	b.WriteString("Replace the preserved copy **last**. Until it is replaced it is the only " +
		"record of what upstream looked like when the distillation was written.\n")

	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
