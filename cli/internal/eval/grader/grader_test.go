package grader

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
)

type stubJudge struct {
	results []ExpectationResult
	err     error
}

func (s *stubJudge) Grade(ctx context.Context, prompt string, transcript []byte, outputs string,
	assertions []scenarios.Assertion) ([]ExpectationResult, error) {
	return s.results, s.err
}

func TestScriptedChecks(t *testing.T) {
	workDir := t.TempDir()
	outputDir := filepath.Join(workDir, "outputs")
	_ = os.MkdirAll(outputDir, 0o755)
	_ = os.WriteFile(filepath.Join(outputDir, "report.md"), []byte("# Hello\nworld\n"), 0o644)
	_ = os.WriteFile(filepath.Join(outputDir, "data.json"), []byte(`{"k":1}`), 0o644)

	req := Request{
		EvalPrompt: "demo",
		Assertions: []scenarios.Assertion{
			{Text: "report exists", Check: "path_exists:report.md"},
			{Text: "regex hit", Check: "regex:report.md:^# Hello$"},
			{Text: "json ok", Check: "json_valid:data.json"},
			{Text: "exec ok", Check: "exec:true"},
			{Text: "missing", Check: "path_exists:nope.txt"},
		},
		OutputDir: outputDir,
		WorkDir:   workDir,
	}
	g, err := Grade(context.Background(), req)
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	if g.Summary.Total != 5 {
		t.Fatalf("total: %d", g.Summary.Total)
	}
	if g.Summary.Passed != 4 {
		t.Fatalf("passed: %d (want 4)", g.Summary.Passed)
	}
	if g.Summary.PassRate != 0.8 {
		t.Fatalf("pass_rate: %v (want 0.8)", g.Summary.PassRate)
	}
}

func TestLLMAssertionsWithoutJudgeAutoFail(t *testing.T) {
	req := Request{
		Assertions: []scenarios.Assertion{
			{Text: "vibe check", Check: "llm"},
		},
		OutputDir: t.TempDir(),
		WorkDir:   t.TempDir(),
	}
	g, _ := Grade(context.Background(), req)
	if g.Summary.Failed != 1 {
		t.Fatalf("expected 1 fail when judge absent, got %d", g.Summary.Failed)
	}
}

func TestLLMAssertionsUseJudge(t *testing.T) {
	req := Request{
		Assertions: []scenarios.Assertion{
			{Text: "it's good", Check: "llm"},
		},
		OutputDir: t.TempDir(),
		WorkDir:   t.TempDir(),
		LLMJudge: &stubJudge{results: []ExpectationResult{
			{Text: "it's good", Passed: true, Evidence: "looks fine"},
		}},
	}
	g, err := Grade(context.Background(), req)
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	if g.Summary.Passed != 1 {
		t.Fatalf("expected 1 pass, got %d", g.Summary.Passed)
	}
	if g.Expectations[0].Evidence != "looks fine" {
		t.Fatalf("evidence not carried through")
	}
}

func TestGradingJSONRoundTrip(t *testing.T) {
	g := &Grading{
		Expectations: []ExpectationResult{{Text: "x", Passed: true, Evidence: "ok"}},
		Summary:      Summary{Passed: 1, Total: 1, PassRate: 1},
	}
	path := filepath.Join(t.TempDir(), "grading.json")
	if err := Write(path, g); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Fatalf("empty grading.json")
	}
}
