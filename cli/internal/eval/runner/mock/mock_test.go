package mock_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/mock"
)

func TestName(t *testing.T) {
	if got := mock.New().Name(); got != "mock" {
		t.Errorf("got %q", got)
	}
}

func TestCapabilities_ReportsDefaultModel(t *testing.T) {
	caps := mock.New().Capabilities()
	if caps.DefaultModel == "" {
		t.Error("DefaultModel empty")
	}
	if len(caps.SupportsTools) == 0 {
		t.Error("SupportsTools empty")
	}
	if !caps.SupportsParallel {
		t.Error("SupportsParallel should be true")
	}
}

func TestDoctorCheck_AlwaysAvailable(t *testing.T) {
	got := mock.New().DoctorCheck(context.Background())
	if !got.Available {
		t.Error("mock should always be available")
	}
}

func TestExecute_WritesStubOutputAndTranscript(t *testing.T) {
	outDir := t.TempDir()
	r := mock.New()
	req := runner.Request{
		Prompt:    "hello world",
		OutputDir: outDir,
	}
	res, err := r.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.TotalTokens <= 0 {
		t.Errorf("TotalTokens = %d", res.TotalTokens)
	}
	if len(res.Transcript) == 0 {
		t.Error("empty transcript")
	}
	if len(res.OutputFiles) != 1 || res.OutputFiles[0] != "mock-output.txt" {
		t.Errorf("OutputFiles = %v", res.OutputFiles)
	}
	// File written to OutputDir.
	body, err := os.ReadFile(filepath.Join(outDir, "mock-output.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "hello world") {
		t.Errorf("output body missing prompt: %s", body)
	}
	// ToolCalls populated.
	if res.ToolCalls["Write"] != 1 {
		t.Errorf("Write tool count = %d", res.ToolCalls["Write"])
	}
	if res.ToolCalls["Bash"] != 1 {
		t.Errorf("Bash tool count = %d", res.ToolCalls["Bash"])
	}
	// metrics.json sidecar written.
	if _, err := os.Stat(filepath.Join(outDir, "metrics.json")); err != nil {
		t.Errorf("metrics.json missing: %v", err)
	}
}

func TestExecute_InputFilesBumpReadCount(t *testing.T) {
	workDir := t.TempDir()
	file1 := filepath.Join(workDir, "in1.md")
	file2 := filepath.Join(workDir, "in2.md")
	for _, p := range []string{file1, file2} {
		_ = os.WriteFile(p, []byte("content"), 0o644)
	}

	r := mock.New()
	res, err := r.Execute(context.Background(), runner.Request{
		Prompt:     "p",
		OutputDir:  t.TempDir(),
		InputFiles: []string{file1, file2},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.ToolCalls["Read"] != 2 {
		t.Errorf("Read tool count = %d, want 2", res.ToolCalls["Read"])
	}
}

func TestExecute_SmartSkillBoostsBrainReadsAndReducesCompletion(t *testing.T) {
	skillDir := t.TempDir()
	refsDir := filepath.Join(skillDir, "references")
	if err := os.MkdirAll(refsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Give mock all four brain files so it reports max brain_reads.
	for _, name := range []string{"_index.md", "patterns.md", "decisions.md", "log.md"} {
		_ = os.WriteFile(filepath.Join(refsDir, name), []byte("x"), 0o644)
	}

	withSkill, err := mock.New().Execute(context.Background(), runner.Request{
		Prompt:    "p",
		SkillDir:  skillDir,
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	withoutSkill, err := mock.New().Execute(context.Background(), runner.Request{
		Prompt:    "p",
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}

	// With brain: prompt tokens larger, completion tokens smaller.
	if withSkill.PromptTokens <= withoutSkill.PromptTokens {
		t.Errorf("smart skill should use more prompt tokens: with=%d without=%d",
			withSkill.PromptTokens, withoutSkill.PromptTokens)
	}
	if withSkill.CompletionTokens >= withoutSkill.CompletionTokens {
		t.Errorf("smart skill should use fewer completion tokens: with=%d without=%d",
			withSkill.CompletionTokens, withoutSkill.CompletionTokens)
	}
	// Read tool count must include brain reads.
	if withSkill.ToolCalls["Read"] < 4 {
		t.Errorf("Read count %d doesn't include brain reads", withSkill.ToolCalls["Read"])
	}
}
