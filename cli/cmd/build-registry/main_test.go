package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
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

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "b"); got != "b" {
		t.Errorf("got %q", got)
	}
	if got := firstNonEmpty(); got != "" {
		t.Errorf("got %q", got)
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
