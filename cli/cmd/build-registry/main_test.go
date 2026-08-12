package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/registry"
)

// validSkillMD is a minimal SKILL.md that satisfies frontmatter.Validate.
func validSkillMD(name, version string) string {
	return "---\n" +
		"name: " + name + "\n" +
		"description: " + name + " description line long enough\n" +
		"version: " + version + "\n" +
		"metadata:\n" +
		"  category: development\n" +
		"---\n\n# " + name + "\n\nBody.\n"
}

func writeSkill(t *testing.T, skillsDir, name, version string) {
	t.Helper()
	d := filepath.Join(skillsDir, name)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(validSkillMD(name, version)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRun_BuildsRegistryForValidSkills(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "alpha", "1.0.0")
	writeSkill(t, skillsDir, "beta", "0.9.0")

	out := filepath.Join(root, "registry.json")
	if err := run(skillsDir, out, "github.com/x/y", "main", "deadbeef", false); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var reg registry.Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if reg.SchemaVersion != registry.SchemaVersion {
		t.Errorf("schema = %d", reg.SchemaVersion)
	}
	if len(reg.Skills) != 2 {
		t.Errorf("skills = %d", len(reg.Skills))
	}
	// Sorted by name: alpha before beta.
	if reg.Skills[0].Name != "alpha" || reg.Skills[1].Name != "beta" {
		t.Errorf("not sorted: %v", []string{reg.Skills[0].Name, reg.Skills[1].Name})
	}
	// DirSHA must be set for each skill.
	for _, s := range reg.Skills {
		if s.DirSHA == "" {
			t.Errorf("%s: empty DirSHA", s.Name)
		}
		if s.Category != "development" {
			t.Errorf("%s: category = %q, want development", s.Name, s.Category)
		}
	}
}

func TestRun_MissingCategory_Rejected(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	d := filepath.Join(skillsDir, "nocategory")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nname: nocategory\ndescription: missing category on purpose\nversion: 1.0.0\n---\nBody.\n"
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	err := run(skillsDir, filepath.Join(root, "registry.json"), "r", "main", "sha", false)
	if err == nil || !strings.Contains(err.Error(), "category is required") {
		t.Fatalf("expected missing-category validation error, got %v", err)
	}
}

func TestRun_UnknownCategory_Rejected(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	d := filepath.Join(skillsDir, "badcategory")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nname: badcategory\ndescription: unknown category on purpose\nversion: 1.0.0\nmetadata:\n  category: astrology\n---\nBody.\n"
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	err := run(skillsDir, filepath.Join(root, "registry.json"), "r", "main", "sha", false)
	if err == nil || !strings.Contains(err.Error(), `unknown category "astrology"`) {
		t.Fatalf("expected unknown-category validation error, got %v", err)
	}
}

func TestRun_DuplicateSkillName(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "one", "1.0.0")
	// Second skill directory claims same name in its frontmatter.
	dir2 := filepath.Join(skillsDir, "other-dir")
	if err := os.MkdirAll(dir2, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "SKILL.md"), []byte(validSkillMD("one", "2.0.0")), 0o644); err != nil {
		t.Fatal(err)
	}

	err := run(skillsDir, filepath.Join(root, "registry.json"), "r", "main", "sha", false)
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
	if !strings.Contains(err.Error(), "duplicate skill name") {
		t.Errorf("err = %v", err)
	}
}

func TestRun_EmptySkillsDir(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	err := run(skillsDir, filepath.Join(root, "registry.json"), "r", "main", "sha", false)
	if err == nil {
		t.Fatal("expected 'no skills found' error")
	}
}

func TestRun_CheckMode_NoDriftIsClean(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "alpha", "1.0.0")

	out := filepath.Join(root, "registry.json")
	if err := run(skillsDir, out, "r", "main", "sha", false); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Same inputs → --check should pass.
	if err := run(skillsDir, out, "r", "main", "sha", true); err != nil {
		t.Errorf("--check false positive: %v", err)
	}
}

func TestRun_CheckMode_DriftFailsLoudly(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "alpha", "1.0.0")

	out := filepath.Join(root, "registry.json")
	if err := run(skillsDir, out, "r", "main", "sha", false); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Mutate a skill so registry should drift.
	writeSkill(t, skillsDir, "alpha", "2.0.0")

	err := run(skillsDir, out, "r", "main", "sha", true)
	if err == nil {
		t.Fatal("expected drift error in --check mode")
	}
	if !strings.Contains(err.Error(), "out of date") {
		t.Errorf("err = %v", err)
	}
}

func TestRun_IgnoresNonSkillDirs(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "good", "1.0.0")
	// Hidden dir gets ignored.
	_ = os.MkdirAll(filepath.Join(skillsDir, ".hidden"), 0o755)
	// Non-directory entry gets ignored.
	_ = os.WriteFile(filepath.Join(skillsDir, "README.md"), []byte("x"), 0o644)
	// Dir without SKILL.md gets silently skipped.
	_ = os.MkdirAll(filepath.Join(skillsDir, "incomplete"), 0o755)

	out := filepath.Join(root, "registry.json")
	if err := run(skillsDir, out, "r", "main", "sha", false); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, _ := os.ReadFile(out)
	var reg registry.Registry
	_ = json.Unmarshal(data, &reg)
	if len(reg.Skills) != 1 || reg.Skills[0].Name != "good" {
		t.Errorf("skills = %+v", reg.Skills)
	}
}

func TestRun_CycleIsRejected(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	// a → b → a (cycle)
	for _, x := range [][2]string{{"a", "b"}, {"b", "a"}} {
		name, dep := x[0], x[1]
		body := "---\n" +
			"name: " + name + "\n" +
			"description: cycle test skill body long enough\n" +
			"version: 1.0.0\n" +
			"metadata:\n" +
			"  category: development\n" +
			"  requires:\n" +
			"    - " + dep + "@1.0.0\n" +
			"---\n# " + name + "\nBody.\n"
		d := filepath.Join(skillsDir, name)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	err := run(skillsDir, filepath.Join(root, "registry.json"), "r", "main", "sha", false)
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestRun_DirSHADeterministicAcrossRuns(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "alpha", "1.0.0")

	out1 := filepath.Join(root, "reg1.json")
	out2 := filepath.Join(root, "reg2.json")
	if err := run(skillsDir, out1, "r", "main", "sha", false); err != nil {
		t.Fatal(err)
	}
	if err := run(skillsDir, out2, "r", "main", "sha", false); err != nil {
		t.Fatal(err)
	}

	var r1, r2 registry.Registry
	d1, _ := os.ReadFile(out1)
	d2, _ := os.ReadFile(out2)
	_ = json.Unmarshal(d1, &r1)
	_ = json.Unmarshal(d2, &r2)
	if r1.Skills[0].DirSHA != r2.Skills[0].DirSHA {
		t.Errorf("DirSHA non-deterministic: %s vs %s", r1.Skills[0].DirSHA, r2.Skills[0].DirSHA)
	}
}

// A rebuild from a different branch and commit must leave the file untouched
// when no skill changed. Before this, every run rewrote generated_at and
// source.ref/sha, so the CI auto-fix committed on every push and each branch's
// registry.json conflicted with every other branch's.
func TestRun_ByteStableWhenNothingChanged(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "alpha", "1.0.0")
	out := filepath.Join(root, "registry.json")

	if err := run(skillsDir, out, "r", "main", "sha-one", false); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	// Same skills, different branch and commit — the shape a second CI run takes.
	if err := run(skillsDir, out, "r", "some-feature-branch", "sha-two", false); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("rebuild rewrote the file with no skill change:\nfirst:  %s\nsecond: %s", first, second)
	}

	// A real change must still be written.
	writeSkill(t, skillsDir, "beta", "1.0.0")
	if err := run(skillsDir, out, "r", "main", "sha-three", false); err != nil {
		t.Fatal(err)
	}
	third, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(second, third) {
		t.Error("adding a skill did not update registry.json")
	}
}

func TestMarshalStable_NoHTMLEscape(t *testing.T) {
	reg := registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Skills: []registry.Skill{
			{Name: "a", Description: "uses > and <", Version: "1.0.0"},
		},
	}
	data, err := marshalStable(reg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `\u003c`) || strings.Contains(string(data), `\u003e`) {
		t.Errorf("HTML escape leaked: %s", data)
	}
}

func TestSemanticDiff_IgnoresSourceAndGeneratedAt(t *testing.T) {
	a := []byte(`{"schema_version":1,"generated_at":"t1","source":{"repo":"x","ref":"main","sha":"a"},"skills":[]}`)
	b := []byte(`{"schema_version":1,"generated_at":"t2","source":{"repo":"x","ref":"main","sha":"b"},"skills":[]}`)
	diff, err := semanticDiff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if diff {
		t.Error("semantic diff should ignore generated_at / source")
	}
}

func TestSemanticDiff_DetectsSkillChange(t *testing.T) {
	a := []byte(`{"schema_version":1,"skills":[{"name":"x","version":"1.0.0"}]}`)
	b := []byte(`{"schema_version":1,"skills":[{"name":"x","version":"2.0.0"}]}`)
	diff, err := semanticDiff(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !diff {
		t.Error("expected diff detected")
	}
}

func TestRun_RoleFlowsIntoRegistry(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	d := filepath.Join(skillsDir, "fdeskill")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nname: fdeskill\ndescription: role-tagged on purpose\nversion: 1.0.0\nmetadata:\n  category: development\n  role: fde\n---\nBody.\n"
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(root, "registry.json")
	if err := run(skillsDir, out, "r", "main", "sha", false); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var reg registry.Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatal(err)
	}
	if len(reg.Skills) != 1 || reg.Skills[0].Role != "fde" {
		t.Errorf("skills = %+v, want one skill with role fde", reg.Skills)
	}
}

func TestRun_UnknownRole_Rejected(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	d := filepath.Join(skillsDir, "badrole")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nname: badrole\ndescription: unknown role on purpose\nversion: 1.0.0\nmetadata:\n  category: development\n  role: astronaut\n---\nBody.\n"
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	err := run(skillsDir, filepath.Join(root, "registry.json"), "r", "main", "sha", false)
	if err == nil || !strings.Contains(err.Error(), `unknown role "astronaut"`) {
		t.Fatalf("expected unknown-role validation error, got %v", err)
	}
}

// --- source.sha staleness -------------------------------------------------
//
// Regression cover for the v2.47.0 bug: a registry.json whose source.sha
// predates a skill it lists installs as `no files found under skills/<name>
// in tarball`, and semanticDiff (which zeroes the source block) made that
// state permanent.

// fakeTree substitutes gitPathsAt with an in-memory sha -> paths map for the
// duration of a test. A sha absent from the map behaves like an object git
// cannot resolve, which is the "do not draw conclusions" case.
func fakeTree(t *testing.T, trees map[string][]string) {
	t.Helper()
	prev := gitPathsAt
	t.Cleanup(func() { gitPathsAt = prev })
	gitPathsAt = func(sha string, paths []string) (map[string]bool, error) {
		have, ok := trees[sha]
		if !ok {
			return nil, fmt.Errorf("git ls-tree %s: not a valid object name", sha)
		}
		set := make(map[string]bool, len(have))
		for _, p := range have {
			set[p] = true
		}
		out := make(map[string]bool, len(paths))
		for _, p := range paths {
			if set[p] {
				out[p] = true
			}
		}
		return out, nil
	}
}

func registryBytes(t *testing.T, sha string, paths ...string) []byte {
	t.Helper()
	reg := registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "r", Ref: "main", SHA: sha},
	}
	for _, p := range paths {
		reg.Skills = append(reg.Skills, registry.Skill{Name: filepath.Base(p), Path: p})
	}
	b, err := json.Marshal(reg)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSourceSHAVerdict(t *testing.T) {
	const (
		old = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // pre-dates skills/beta
		new = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" // contains both
	)
	both := []string{"skills/alpha", "skills/beta"}
	skills := []registry.Skill{{Name: "alpha", Path: "skills/alpha"}, {Name: "beta", Path: "skills/beta"}}

	tests := []struct {
		name        string
		trees       map[string][]string
		existingSHA string
		newSHA      string
		wantRewrite bool
		wantNote    string // substring; "" means no note
	}{
		{
			name:        "recorded sha contains every skill",
			trees:       map[string][]string{old: both, new: both},
			existingSHA: old, newSHA: new,
			wantRewrite: false,
		},
		{
			name:        "recorded sha predates a skill and the new one fixes it",
			trees:       map[string][]string{old: {"skills/alpha"}, new: both},
			existingSHA: old, newSHA: new,
			wantRewrite: true, wantNote: "predates skills/beta",
		},
		{
			// Both broken: rewriting cannot help, and doing it anyway would
			// make every run dirty the file -> the workflow pushes forever.
			name:        "neither sha contains the skill",
			trees:       map[string][]string{old: {"skills/alpha"}, new: {"skills/alpha"}},
			existingSHA: old, newSHA: new,
			wantRewrite: false, wantNote: "cannot be installed",
		},
		{
			name:        "git cannot resolve the recorded sha",
			trees:       map[string][]string{new: both},
			existingSHA: old, newSHA: new,
			wantRewrite: false,
		},
		{
			name:        "same sha is never rewritten",
			trees:       map[string][]string{old: {"skills/alpha"}},
			existingSHA: old, newSHA: old,
			wantRewrite: false,
		},
		{
			name:        "empty recorded sha is left alone",
			trees:       map[string][]string{new: both},
			existingSHA: "", newSHA: new,
			wantRewrite: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeTree(t, tc.trees)
			rewrite, note := sourceSHAVerdict(registryBytes(t, tc.existingSHA, both...), tc.newSHA, skills)
			if rewrite != tc.wantRewrite {
				t.Errorf("rewrite = %v, want %v (note: %q)", rewrite, tc.wantRewrite, note)
			}
			if tc.wantNote == "" && note != "" {
				t.Errorf("note = %q, want none", note)
			}
			if tc.wantNote != "" && !strings.Contains(note, tc.wantNote) {
				t.Errorf("note = %q, want substring %q", note, tc.wantNote)
			}
		})
	}
}

// TestRun_StaleSourceSHA_IsRewrittenThenConverges is the end-to-end shape of
// the bug: content identical, source.sha too old. The first rebuild must fix
// the SHA; the second must take the early exit, or the Registry workflow
// commits and pushes on every trigger.
func TestRun_StaleSourceSHA_IsRewrittenThenConverges(t *testing.T) {
	const (
		beforeBeta = "1111111111111111111111111111111111111111"
		afterBeta  = "2222222222222222222222222222222222222222"
	)
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "alpha", "1.0.0")
	writeSkill(t, skillsDir, "beta", "1.0.0")
	out := filepath.Join(root, "registry.json")

	// Seed the broken state: both skills listed, SHA from before beta existed.
	fakeTree(t, map[string][]string{
		beforeBeta: {"skills/alpha"},
		afterBeta:  {"skills/alpha", "skills/beta"},
	})
	if err := run(skillsDir, out, "r", "main", beforeBeta, false); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	if got := readSourceSHA(t, out); got != beforeBeta {
		t.Fatalf("seed sha = %s, want %s", got, beforeBeta)
	}

	// Rebuild at a commit that does contain beta: nothing semantic changed,
	// but the SHA must still be replaced.
	if err := run(skillsDir, out, "r", "main", afterBeta, false); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if got := readSourceSHA(t, out); got != afterBeta {
		t.Fatalf("sha = %s, want it rewritten to %s", got, afterBeta)
	}

	// Convergence: a second rebuild at a third good commit must be a no-op,
	// since the recorded SHA now resolves every skill.
	const laterStillGood = "3333333333333333333333333333333333333333"
	fakeTree(t, map[string][]string{
		afterBeta:      {"skills/alpha", "skills/beta"},
		laterStillGood: {"skills/alpha", "skills/beta"},
	})
	before, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := run(skillsDir, out, "r", "main", laterStillGood, false); err != nil {
		t.Fatalf("converge run: %v", err)
	}
	after, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("registry.json changed on a run with a healthy source.sha; the Registry workflow would push on every trigger")
	}
}

// TestRun_OutsideGitRepo_KeepsEarlyExit guards the fallback: when git cannot
// answer, behaviour must be exactly what it was before this check existed.
func TestRun_OutsideGitRepo_KeepsEarlyExit(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "skills")
	writeSkill(t, skillsDir, "alpha", "1.0.0")
	out := filepath.Join(root, "registry.json")

	fakeTree(t, nil) // every lookup errors, as outside a repository
	if err := run(skillsDir, out, "r", "main", "1111111111111111111111111111111111111111", false); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := run(skillsDir, out, "r", "main", "9999999999999999999999999999999999999999", false); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("file was rewritten despite git being unable to verify the recorded sha")
	}
}

func readSourceSHA(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var reg registry.Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatal(err)
	}
	return reg.Source.SHA
}

// TestGitLsTree exercises the real git plumbing the fake stands in for: a
// directory path resolves as a tree entry, an absent path is simply missing
// from the output, and an unknown revision is an error rather than "absent".
func TestGitLsTree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	gitRun := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	gitRun("init", "-q")
	gitRun("config", "user.email", "t@example.com")
	gitRun("config", "user.name", "t")
	writeSkill(t, filepath.Join(repo, "skills"), "alpha", "1.0.0")
	gitRun("add", "-A")
	gitRun("commit", "-qm", "add alpha")

	shaOut, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(string(shaOut))

	// gitLsTree shells out to git in the process CWD, so run from the repo.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	present, err := gitLsTree(sha, []string{"skills/alpha", "skills/beta"})
	if err != nil {
		t.Fatalf("gitLsTree: %v", err)
	}
	if !present["skills/alpha"] {
		t.Error("skills/alpha should resolve as a tree entry at HEAD")
	}
	if present["skills/beta"] {
		t.Error("skills/beta was never committed and must not resolve")
	}

	if _, err := gitLsTree("0000000000000000000000000000000000000000", []string{"skills/alpha"}); err == nil {
		t.Error("an unresolvable revision must error, not report the path as absent")
	}

	// Paths recorded in the registry are repo-root-relative, but the Makefile
	// invokes this binary as `go -C cli run ...`, so the process CWD is a
	// subdirectory. Without --full-tree git would resolve "skills/alpha" as
	// "<subdir>/skills/alpha" and report every skill as missing — which reads
	// as "the recorded SHA is broken" for a registry that is perfectly fine.
	sub := filepath.Join(repo, "cli")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}
	fromSub, err := gitLsTree(sha, []string{"skills/alpha"})
	if err != nil {
		t.Fatalf("gitLsTree from subdirectory: %v", err)
	}
	if !fromSub["skills/alpha"] {
		t.Error("skills/alpha must resolve from a subdirectory; pathspec is repo-root-relative")
	}
}
