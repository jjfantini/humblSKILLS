package scenarios_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
)

func writeJSON(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_MissingDirErrors(t *testing.T) {
	if _, err := scenarios.Load(filepath.Join(t.TempDir(), "evals")); err == nil {
		t.Fatal("expected error for missing evals dir")
	}
}

func TestLoad_PrefersScenariosOverEvals(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "scenarios.json"), `{
		"skill_name": "foo",
		"schema_version": 1,
		"scenarios": [
			{"id":"s1","sessions":[{"n":1,"prompt":"hi","assertions":[{"text":"t1"}]}]}
		]
	}`)
	// evals.json exists but should be ignored.
	writeJSON(t, filepath.Join(dir, "evals.json"), `{"skill_name":"foo","evals":[]}`)

	f, err := scenarios.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if f.Scenarios[0].ID != "s1" {
		t.Errorf("got %q", f.Scenarios[0].ID)
	}
}

func TestLoad_LegacyEvalsJSONLifted(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "evals.json"), `{
		"skill_name":"foo",
		"evals":[
			{"id":1,"prompt":"do a","assertions":[{"text":"t"}]},
			{"id":2,"prompt":"do b"}
		]
	}`)
	f, err := scenarios.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(f.Scenarios) != 2 {
		t.Fatalf("scenarios = %d", len(f.Scenarios))
	}
	if f.Scenarios[0].ID != "eval-1" || f.Scenarios[0].Family != "generic" {
		t.Errorf("lift mismatch: %+v", f.Scenarios[0])
	}
	// applyDefaults fills in configurations.
	if len(f.Configurations) != 3 {
		t.Errorf("configurations = %v", f.Configurations)
	}
	if f.RunsPerConfiguration != 1 {
		t.Errorf("RunsPerConfiguration = %d", f.RunsPerConfiguration)
	}
}

func TestLoad_MalformedScenariosJSON(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "scenarios.json"), "{not json")
	if _, err := scenarios.Load(dir); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoad_MalformedEvalsJSON(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "evals.json"), "{not json")
	if _, err := scenarios.Load(dir); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadFromSkill_ResolvesEvalsSubdir(t *testing.T) {
	skill := t.TempDir()
	writeJSON(t, filepath.Join(skill, "evals", "evals.json"), `{
		"skill_name":"x",
		"evals":[{"id":1,"prompt":"hi"}]
	}`)
	f, err := scenarios.LoadFromSkill(skill)
	if err != nil {
		t.Fatalf("LoadFromSkill: %v", err)
	}
	if len(f.Scenarios) != 1 {
		t.Error("scenario not lifted")
	}
}

func TestValidate_RejectsUnknownSchemaVersion(t *testing.T) {
	f := &scenarios.File{SchemaVersion: 42, SkillName: "x"}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected schema version error")
	}
}

func TestValidate_MissingSkillName(t *testing.T) {
	f := &scenarios.File{SchemaVersion: scenarios.SchemaVersion}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected skill_name error")
	}
}

func TestValidate_UnknownConfiguration(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion:  scenarios.SchemaVersion,
		SkillName:      "x",
		Configurations: []string{"bogus"},
	}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected unknown configuration error")
	}
}

func TestValidate_DuplicateScenarioID(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios: []scenarios.Scenario{
			{ID: "dup", Sessions: []scenarios.Session{{N: 1, Prompt: "p"}}},
			{ID: "dup", Sessions: []scenarios.Session{{N: 1, Prompt: "p"}}},
		},
	}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected duplicate ID error")
	}
}

func TestValidate_MissingScenarioID(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios:     []scenarios.Scenario{{Sessions: []scenarios.Session{{N: 1, Prompt: "p"}}}},
	}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected id required error")
	}
}

func TestValidate_ZeroSessions(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios:     []scenarios.Scenario{{ID: "s1"}},
	}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected at least one session error")
	}
}

func TestValidate_SessionNOutOfOrder(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios: []scenarios.Scenario{{
			ID: "s1",
			Sessions: []scenarios.Session{
				{N: 1, Prompt: "p"},
				{N: 3, Prompt: "p"}, // should be 2
			},
		}},
	}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected session numbering error")
	}
}

func TestValidate_SessionWithEmptyPrompt(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios: []scenarios.Scenario{{
			ID:       "s1",
			Sessions: []scenarios.Session{{N: 1, Prompt: "   "}},
		}},
	}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected empty prompt error")
	}
}

func TestValidate_AssertionWithEmptyText(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios: []scenarios.Scenario{{
			ID: "s1",
			Sessions: []scenarios.Session{{
				N: 1, Prompt: "p",
				Assertions: []scenarios.Assertion{{Text: "", Check: "llm"}},
			}},
		}},
	}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected empty assertion text error")
	}
}

func TestValidate_UnknownAssertionKind(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios: []scenarios.Scenario{{
			ID: "s1",
			Sessions: []scenarios.Session{{
				N: 1, Prompt: "p",
				Assertions: []scenarios.Assertion{{Text: "t", Check: "bogus-kind"}},
			}},
		}},
	}
	err := scenarios.Validate(f)
	if err == nil {
		t.Fatal("expected unknown kind error")
	}
	if !strings.Contains(err.Error(), "unknown assertion kind") {
		t.Errorf("err = %v", err)
	}
}

func TestValidate_TransferFromUnknownScenario(t *testing.T) {
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios: []scenarios.Scenario{
			{
				ID: "s1", TransferFrom: []string{"ghost"},
				Sessions: []scenarios.Session{{N: 1, Prompt: "p"}},
			},
		},
	}
	if err := scenarios.Validate(f); err == nil {
		t.Fatal("expected transfer_from error")
	}
}

func TestValidate_Nil(t *testing.T) {
	if err := scenarios.Validate(nil); err == nil {
		t.Fatal("expected nil error")
	}
}

func TestValidate_StampsMissingSessionN(t *testing.T) {
	// Session N=0 with correct position is auto-stamped rather than rejected.
	f := &scenarios.File{
		SchemaVersion: scenarios.SchemaVersion,
		SkillName:     "x",
		Scenarios: []scenarios.Scenario{{
			ID: "s1",
			Sessions: []scenarios.Session{
				{N: 0, Prompt: "first"},
				{N: 0, Prompt: "second"},
			},
		}},
	}
	if err := scenarios.Validate(f); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if f.Scenarios[0].Sessions[0].N != 1 || f.Scenarios[0].Sessions[1].N != 2 {
		t.Errorf("N not stamped: %+v", f.Scenarios[0].Sessions)
	}
}

func TestParseCheck_AllKinds(t *testing.T) {
	cases := []struct {
		in       string
		kind     scenarios.CheckKind
		arg      string
	}{
		{"", scenarios.CheckLLM, ""},
		{"llm", scenarios.CheckLLM, ""},
		{"path_exists:foo/bar", scenarios.CheckPathExists, "foo/bar"},
		{"regex:out.md:^H", scenarios.CheckRegex, "out.md:^H"},
		{"exec:go test", scenarios.CheckExec, "go test"},
		{"script:check.sh", scenarios.CheckScript, "check.sh"},
		{"json_valid:out.json", scenarios.CheckJSONValid, "out.json"},
	}
	for _, tc := range cases {
		k, a := scenarios.ParseCheck(tc.in)
		if k != tc.kind || a != tc.arg {
			t.Errorf("ParseCheck(%q) = (%q,%q), want (%q,%q)", tc.in, k, a, tc.kind, tc.arg)
		}
	}
}

func TestTotalSessions(t *testing.T) {
	f := &scenarios.File{
		Scenarios: []scenarios.Scenario{
			{Sessions: []scenarios.Session{{}, {}}},
			{Sessions: []scenarios.Session{{}}},
		},
	}
	if got := f.TotalSessions(); got != 3 {
		t.Errorf("got %d", got)
	}
}

func TestFindScenario(t *testing.T) {
	f := &scenarios.File{Scenarios: []scenarios.Scenario{{ID: "a"}, {ID: "b"}}}
	if got := f.FindScenario("a"); got == nil || got.ID != "a" {
		t.Error("FindScenario(a) miss")
	}
	if got := f.FindScenario("z"); got != nil {
		t.Error("FindScenario(z) unexpectedly hit")
	}
}

func TestApplyDefaults_SessionTimeoutConvertedOnLoad(t *testing.T) {
	// Load exercises applyDefaults, which converts TimeoutSec to
	// Session.Timeout.
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "scenarios.json"), `{
		"skill_name":"x","schema_version":1,
		"scenarios":[{
			"id":"s1",
			"sessions":[{"n":1,"prompt":"p","timeout_seconds":30}]
		}]
	}`)
	f, err := scenarios.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if f.Scenarios[0].Sessions[0].Timeout.Seconds() != 30 {
		t.Errorf("Timeout = %v, want 30s", f.Scenarios[0].Sessions[0].Timeout)
	}
}
