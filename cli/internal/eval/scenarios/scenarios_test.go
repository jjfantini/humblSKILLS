package scenarios

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadScenariosHappyPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "scenarios.json", `{
	  "skill_name": "demo",
	  "schema_version": 1,
	  "scenarios": [
	    {"id": "s1", "family": "create", "sessions": [
	      {"n": 1, "prompt": "hello", "assertions": [
	        {"text": "output exists", "check": "path_exists:out.txt"},
	        {"text": "looks good", "check": "llm"}
	      ]},
	      {"n": 2, "prompt": "and again"}
	    ]}
	  ]
	}`)
	f, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if f.SkillName != "demo" {
		t.Fatalf("SkillName: %q", f.SkillName)
	}
	if len(f.Configurations) != 4 {
		t.Fatalf("expected 4 default configurations, got %v", f.Configurations)
	}
	if f.RunsPerConfiguration != 1 {
		t.Fatalf("RunsPerConfiguration default: %d", f.RunsPerConfiguration)
	}
	if f.TotalSessions() != 2 {
		t.Fatalf("TotalSessions: %d", f.TotalSessions())
	}
}

func TestLegacyEvalsJSONLift(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "evals.json", `{
	  "skill_name": "legacy",
	  "evals": [
	    {"id": 1, "prompt": "pick top 3", "expected_output": "bar chart",
	     "assertions": [{"text": "has bar chart"}]}
	  ]
	}`)
	f, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(f.Scenarios) != 1 || f.Scenarios[0].Family != "generic" {
		t.Fatalf("bad lift: %+v", f.Scenarios)
	}
	if len(f.Scenarios[0].Sessions) != 1 {
		t.Fatalf("expected one session per legacy eval")
	}
}

func TestValidateRejectsDuplicateIDs(t *testing.T) {
	f := &File{
		SkillName:     "x",
		SchemaVersion: 1,
		Scenarios: []Scenario{
			{ID: "dup", Sessions: []Session{{N: 1, Prompt: "a"}}},
			{ID: "dup", Sessions: []Session{{N: 1, Prompt: "b"}}},
		},
	}
	applyDefaults(f)
	if err := Validate(f); err == nil {
		t.Fatalf("expected error for duplicate ids")
	}
}

func TestValidateRejectsOutOfOrderSessions(t *testing.T) {
	f := &File{
		SkillName:     "x",
		SchemaVersion: 1,
		Scenarios: []Scenario{
			{ID: "s", Sessions: []Session{
				{N: 1, Prompt: "a"},
				{N: 3, Prompt: "b"},
			}},
		},
	}
	applyDefaults(f)
	if err := Validate(f); err == nil {
		t.Fatalf("expected error for out-of-order sessions")
	}
}

func TestParseCheck(t *testing.T) {
	tests := []struct {
		in   string
		kind CheckKind
		arg  string
	}{
		{"", CheckLLM, ""},
		{"llm", CheckLLM, ""},
		{"path_exists:foo/bar", CheckPathExists, "foo/bar"},
		{"exec:bash scripts/lint.sh", CheckExec, "bash scripts/lint.sh"},
		{"regex:out.md:^Hello", CheckRegex, "out.md:^Hello"},
	}
	for _, tc := range tests {
		k, a := ParseCheck(tc.in)
		if k != tc.kind || a != tc.arg {
			t.Errorf("ParseCheck(%q) = (%s,%q), want (%s,%q)", tc.in, k, a, tc.kind, tc.arg)
		}
	}
}
