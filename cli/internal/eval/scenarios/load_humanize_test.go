package scenarios

import (
	"os"
	"path/filepath"
	"testing"
)

// Regression: bundled skills in this repo must keep a valid scenarios.json.
func TestLoadHumanizeSkillScenarios(t *testing.T) {
	t.Parallel()
	// From cli/internal/eval/scenarios -> repo root is ../../../../
	rel := filepath.Join("..", "..", "..", "..", "skills", "use-smart-humanize-text")
	p, err := filepath.Abs(rel)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(p, "evals", "scenarios.json")); err != nil {
		t.Skip("repo skill fixture not present:", err)
	}
	f, err := LoadFromSkill(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Scenarios) == 0 {
		t.Fatal("no scenarios")
	}
}
