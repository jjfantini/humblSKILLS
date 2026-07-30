package mirrors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
)

// writeSkill lays down a minimal mirrored skill on disk.
func writeSkill(t *testing.T, root, name, upstreamBlock, preservedRel, preservedBody string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(preservedRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	skill := "---\nname: " + name + "\ndescription: test\n" + upstreamBlock +
		"metadata:\n  version: \"1.0.0\"\n  category: design\n---\n\nbody\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(skill), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, preservedRel), []byte(preservedBody), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func block(preserved string) string {
	return "upstream:\n  name: up\n  source: acme/skills\n  fetch: https://example.invalid/x.md\n" +
		"  preserved: " + preserved + "\n  synced: 2026-01-01\n  deltas:\n    - \"adds a thing\"\n"
}

func fixedFetcher(body string) Fetcher {
	return func(context.Context, frontmatter.Upstream) ([]byte, error) { return []byte(body), nil }
}

func TestCheckCurrentIgnoresWhitespaceOnlyDifference(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "m", block("references/raw/up.md"), "references/raw/up.md", "# A\n\ntext\n")

	// CRLF + trailing blank line: byte-different, semantically identical.
	res, err := Check(context.Background(), root, fixedFetcher("# A\r\n\r\ntext\r\n\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 {
		t.Fatalf("want 1 result, got %d", len(res))
	}
	if res[0].Status != StatusCurrent {
		t.Fatalf("line-ending noise must not read as drift; got %s (%s)", res[0].Status, res[0].Reason)
	}
	if res[0].Stale() {
		t.Error("current mirror must not be Stale()")
	}
}

func TestCheckDistinguishesEditFromRewrite(t *testing.T) {
	preserved := "# One\n\n## Two\n\n## Three\n\n## Four\n"
	tests := []struct {
		name    string
		current string
		want    Status
	}{
		{
			// Same structure, changed prose: a spot check, not a re-distill.
			name:    "edited body keeps headings",
			current: "# One\n\nnew prose\n\n## Two\n\n## Three\n\n## Four\n",
			want:    StatusDrifted,
		},
		{
			// Nothing survives: concepts have to be re-derived.
			name:    "restructured",
			current: "# Alpha\n\n## Beta\n\n## Gamma\n",
			want:    StatusRewritten,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeSkill(t, root, "m", block("references/raw/up.md"), "references/raw/up.md", preserved)
			res, err := Check(context.Background(), root, fixedFetcher(tc.current))
			if err != nil {
				t.Fatal(err)
			}
			if res[0].Status != tc.want {
				t.Fatalf("got %s (%s), want %s", res[0].Status, res[0].Reason, tc.want)
			}
			if !res[0].Stale() {
				t.Error("drifted/rewritten must be Stale()")
			}
		})
	}
}

func TestCheckUnreachableUpstreamIsUnknownNotCurrent(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "m", block("references/raw/up.md"), "references/raw/up.md", "# A\n")

	failing := func(context.Context, frontmatter.Upstream) ([]byte, error) {
		return nil, os.ErrDeadlineExceeded
	}
	res, err := Check(context.Background(), root, failing)
	if err != nil {
		t.Fatalf("one unreachable upstream must not fail the whole check: %v", err)
	}
	if res[0].Status != StatusUnknown {
		t.Fatalf("got %s, want unknown — a failed fetch must never read as current", res[0].Status)
	}
	if res[0].Stale() {
		t.Error("unknown is not stale, but Summary must still surface it")
	}
	if !strings.Contains(Summary(res), "could not be checked") {
		t.Errorf("Summary must surface uncheckable mirrors, got %q", Summary(res))
	}
}

func TestCheckSkipsNonMirrors(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "plain")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nname: plain\ndescription: d\nmetadata:\n  version: \"1.0.0\"\n---\n\nx\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Check(context.Background(), root, fixedFetcher("anything"))
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Fatalf("skills without an upstream block are not mirrors; got %d results", len(res))
	}
	if Summary(res) != "" {
		t.Error("Summary must be empty when there is nothing to report")
	}
}

func TestBaselineFileResolvesDirectoryPreserve(t *testing.T) {
	if got := BaselineFile("references/raw/"); got != "references/raw/SKILL.md" {
		t.Errorf("directory preserve should baseline on the upstream SKILL.md, got %q", got)
	}
	if got := BaselineFile("references/raw/x.md"); got != "references/raw/x.md" {
		t.Errorf("file preserve should pass through, got %q", got)
	}
}

func TestAffectedComputesBlastRadiusFromSources(t *testing.T) {
	root := t.TempDir()
	dir := writeSkill(t, root, "m", block("references/raw/up.md"), "references/raw/up.md", "# A\n")

	mk := func(rel, source string) {
		full := filepath.Join(dir, "references", "wiki", rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		body := "---\ntitle: t\nsources:\n  - \"" + source + "\"\n---\n\nbody\n"
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("a/b/hit.md", "references/raw/up.md")
	mk("a/b/miss.md", "references/raw/other.md")

	got, err := Affected(dir, "references/raw/up.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "references/wiki/a/b/hit.md" {
		t.Fatalf("blast radius should be exactly the citing concept, got %v", got)
	}

	// A directory preserve means every file under it is upstream material,
	// so concepts citing companions are in the blast radius too.
	all, err := Affected(dir, "references/raw/")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("directory preserve should span companion sources, got %v", all)
	}
}

func TestWritePlanCarriesTheFourSections(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "m", block("references/raw/up.md"), "references/raw/up.md", "# One\n\n## Two\n")

	res, err := Check(context.Background(), root, fixedFetcher("# Alpha\n\n## Beta\n"))
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, ".mirror-sync")
	a, err := WritePlan(res[0], out)
	if err != nil {
		t.Fatal(err)
	}

	planBytes, err := os.ReadFile(a.PlanPath)
	if err != nil {
		t.Fatal(err)
	}
	plan := string(planBytes)
	for _, want := range []string{
		"## 1. Diff these two files",
		"## 2. Blast radius",
		"## 3. Declared deltas",
		"## 4. Completion checklist",
		"adds a thing", // the declared delta must be reproduced as a guardrail
		"`+` alpha",    // heading delta
		"`-` one",
	} {
		if !strings.Contains(plan, want) {
			t.Errorf("plan missing %q", want)
		}
	}

	// The incoming upstream must land on disk so the executing agent can diff
	// at full fidelity rather than trusting a rendered summary.
	got, err := os.ReadFile(a.IncomingPath)
	if err != nil {
		t.Fatalf("incoming upstream not written: %v", err)
	}
	if string(got) != "# Alpha\n\n## Beta\n" {
		t.Errorf("incoming upstream not written verbatim, got %q", got)
	}
}

func TestWritePlanRefusesWithoutUpstream(t *testing.T) {
	if _, err := WritePlan(Result{Skill: "x"}, t.TempDir()); err == nil {
		t.Fatal("want error when the result has no upstream block")
	}
}

func TestShouldCheckRespectsTTL(t *testing.T) {
	stamp := filepath.Join(t.TempDir(), "nested", "stamp")
	if !ShouldCheck(stamp) {
		t.Error("a never-checked stamp must be due")
	}
	if err := Stamp(stamp); err != nil {
		t.Fatal(err)
	}
	if ShouldCheck(stamp) {
		t.Error("a just-written stamp must not be due again inside the TTL")
	}
}
