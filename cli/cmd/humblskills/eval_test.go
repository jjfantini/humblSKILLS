package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

// --- pure helpers -----------------------------------------------------------

func TestUnionKeys(t *testing.T) {
	a := map[string]int{"x": 1, "y": 2}
	b := map[string]int{"y": 3, "z": 4}
	got := unionKeys(a, b)
	want := []string{"x", "y", "z"} // sorted, deduped
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unionKeys = %v, want %v", got, want)
	}
}

func TestSortedStrKeys(t *testing.T) {
	got := sortedStrKeys(map[string]struct{}{"b": {}, "a": {}, "c": {}})
	if !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("sortedStrKeys = %v", got)
	}
}

func TestAbbreviate(t *testing.T) {
	if got := abbreviate("line one\nline two", 100); got != "line one line two" {
		t.Errorf("newlines should collapse to spaces, got %q", got)
	}
	if got := abbreviate("abcdefgh", 3); got != "abc..." {
		t.Errorf("abbreviate = %q, want abc...", got)
	}
}

func TestSkillBasename(t *testing.T) {
	if got := skillBasename(filepath.Join("a", "b", "my-skill")); got != "my-skill" {
		t.Errorf("skillBasename = %q", got)
	}
}

func TestKnownProviders(t *testing.T) {
	got := knownProviders()
	if len(got) == 0 {
		t.Fatal("expected at least one known provider")
	}
	if !contains(got, "anthropic") {
		t.Errorf("expected anthropic in providers, got %v", got)
	}
}

func TestScaffoldEvalsDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "evals")
	if err := scaffoldEvalsDir(dir, "demo-skill"); err != nil {
		t.Fatalf("scaffoldEvalsDir: %v", err)
	}
	for _, name := range []string{"scenarios.json", "evals.json", "README.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to be created: %v", name, err)
		}
	}
	for _, sub := range []string{"files", "assertions"} {
		if fi, err := os.Stat(filepath.Join(dir, sub)); err != nil || !fi.IsDir() {
			t.Errorf("expected %s/ dir: %v", sub, err)
		}
	}
	// Re-scaffolding into an existing dir must refuse rather than clobber.
	if err := scaffoldEvalsDir(dir, "demo-skill"); err == nil {
		t.Error("expected scaffoldEvalsDir to refuse an existing directory")
	}
}

// --- CLI layer --------------------------------------------------------------

func TestEvalRunners_JSONListsMock(t *testing.T) {
	testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t, "eval", "runners", "--json")
	if res.RunErr != nil {
		t.Fatalf("eval runners --json: %v", res.RunErr)
	}
	// The mock runner is always available and last in the default lineup.
	assertContains(t, res.Out, "mock")
}

func TestEvalWhere_PrintsWorkspacePath(t *testing.T) {
	testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t, "eval", "where")
	if res.RunErr != nil {
		t.Fatalf("eval where: %v", res.RunErr)
	}
	// Default workspace lives under <state>/humblskills/evals.
	assertContains(t, res.Out, "evals")
}

func TestEvalInit_ScaffoldsFromLocalRoot(t *testing.T) {
	s := testutil.NewSandbox(t)
	// Point skill resolution at a sandboxed root containing skills/demo/.
	skillDir := filepath.Join(s.Root, "skills", "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	s.Setenv(t, "HUMBLSKILLS_ROOT", s.Root)

	res := runCLI(t, "eval", "init", "demo")
	if res.RunErr != nil {
		t.Fatalf("eval init demo: %v", res.RunErr)
	}
	if _, err := os.Stat(filepath.Join(skillDir, "evals", "scenarios.json")); err != nil {
		t.Errorf("expected evals/scenarios.json to be scaffolded: %v", err)
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
