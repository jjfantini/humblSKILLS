package harness

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/mock"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
)

// buildSkill creates a minimal smart-skill on disk so DeriveFlat and
// brain.Snapshot have real files to work with.
func buildSkill(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: demo\ndescription: test skill\n---\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "references"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "references", "_index.md"),
		[]byte("# index\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "references", "log.md"),
		[]byte("# Log\n---\n[INGEST 2026-04-01] seed\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "references", "patterns.md"),
		[]byte("# Patterns\n---\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "references", "decisions.md"),
		[]byte("# Decisions\n---\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "references", "wiki", "a", "b"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "references", "wiki", "a", "b", "c.md"),
		[]byte("concept\n"), 0o644)
	return dir
}

func TestHarnessEndToEndMockRunner(t *testing.T) {
	skill := buildSkill(t)
	ws := t.TempDir()
	f := &scenarios.File{
		SkillName:            "demo",
		SchemaVersion:        1,
		Configurations:       []string{scenarios.ArmSmartSkill, scenarios.ArmFlatSkill, scenarios.ArmNoSkill},
		RunsPerConfiguration: 1,
		Scenarios: []scenarios.Scenario{
			{
				ID:     "s1",
				Family: "create",
				Sessions: []scenarios.Session{
					{N: 1, Prompt: "first", Assertions: []scenarios.Assertion{
						{Text: "produced an output file", Check: "path_exists:mock-output.txt"},
					}},
					{N: 2, Prompt: "second", Assertions: []scenarios.Assertion{
						{Text: "produced an output file", Check: "path_exists:mock-output.txt"},
					}},
				},
			},
		},
	}
	h, err := New(Options{
		SkillDir:      skill,
		Scenarios:     f,
		Runner:        mock.New(),
		WorkspaceRoot: ws,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Drain events in a goroutine.
	go func() {
		for range h.Events() {
		}
	}()
	res, err := h.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Iteration != 1 {
		t.Fatalf("iteration: %d", res.Iteration)
	}
	if len(res.Trajectory.Rows) != 2*3 {
		t.Fatalf("expected 6 trajectory rows (2 sessions x 3 arms), got %d", len(res.Trajectory.Rows))
	}
	// report.html must exist and be non-empty.
	info, err := os.Stat(res.ReportHTML)
	if err != nil || info.Size() == 0 {
		t.Fatalf("report.html missing or empty: %v", err)
	}
	// benchmark.json parses.
	var b any
	data, _ := os.ReadFile(filepath.Join(res.IterDir, "benchmark.json"))
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("benchmark.json not valid JSON: %v", err)
	}
	// Smart arm should have a snapshot-after directory in the trajectory.
	found := false
	_ = filepath.Walk(res.IterDir, func(p string, info os.FileInfo, err error) error {
		if err == nil && filepath.Base(p) == "brain-snapshot-after" {
			found = true
		}
		return nil
	})
	if !found {
		t.Fatalf("expected brain-snapshot-after dir for smart arm")
	}
}
