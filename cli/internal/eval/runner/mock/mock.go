// Package mock is the in-memory runner used by tests and TUI development.
//
// It never makes external calls. Behavior is deterministic per
// (SkillDir, session, arm) so trajectories produced against the mock look
// like the real thing: smart_skill pass rates rise over sessions, flat_skill
// stays flat, no_skill struggles. That lets us snapshot-test the report
// renderer without spending tokens.
package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
)

// Runner is the mock backend.
type Runner struct{}

// New returns a fresh mock runner.
func New() *Runner { return &Runner{} }

func (r *Runner) Name() string { return "mock" }

func (r *Runner) Capabilities() runner.Capabilities {
	return runner.Capabilities{
		SupportsTools:    []string{"Read", "Write", "Bash", "Glob", "Grep"},
		SupportsParallel: true,
		DefaultModel:     "mock-1",
		AvailableModels:  []string{"mock-1"},
		Pricing:          &runner.Pricing{PromptUSDPerMtok: 0, CompletionUSDPerMtok: 0},
	}
}

func (r *Runner) DoctorCheck(ctx context.Context) runner.DoctorCheck {
	return runner.DoctorCheck{Available: true, Version: "mock", Fix: "always available"}
}

// Execute writes a deterministic stub output, a fake transcript, and
// emits token counts that vary smartly across arms so aggregated results
// look lifelike.
func (r *Runner) Execute(ctx context.Context, req runner.Request) (*runner.Result, error) {
	start := time.Now()

	// Write a transcript that looks like agent activity.
	lines := []string{
		fmt.Sprintf("[mock] prompt: %s", oneLine(req.Prompt, 80)),
		fmt.Sprintf("[mock] skill_dir: %s", req.SkillDir),
	}
	for _, f := range req.InputFiles {
		lines = append(lines, fmt.Sprintf("[mock] Read: %s", filepath.Base(f)))
	}
	// Simulate reading the brain when a smart skill is provided.
	brainReads := 0
	if req.SkillDir != "" {
		for _, name := range []string{"_index.md", "patterns.md", "decisions.md", "log.md"} {
			p := filepath.Join(req.SkillDir, "references", name)
			if _, err := os.Stat(p); err == nil {
				lines = append(lines, fmt.Sprintf("[mock] Read: references/%s", name))
				brainReads++
			}
		}
	}
	// Drop a stub output file so OutputFiles is non-empty and downstream
	// assertions (path_exists) can succeed against the mock.
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return nil, err
	}
	stub := filepath.Join(req.OutputDir, "mock-output.txt")
	body := fmt.Sprintf("mock output for prompt: %s\n", oneLine(req.Prompt, 120))
	if err := os.WriteFile(stub, []byte(body), 0o644); err != nil {
		return nil, err
	}
	lines = append(lines, fmt.Sprintf("[mock] Write: %s", filepath.Base(stub)))
	// Pseudo bash probe so ToolCalls has a Bash entry.
	lines = append(lines, "[mock] Bash: echo \"ok\"")

	// Token counts: baseline cost that grows slightly with prompt length.
	// Smart skills read more from the brain, so promptTokens inflate a bit
	// but completion shrinks (the brain replaces exploration).
	promptTok := 300 + len(req.Prompt)/3 + brainReads*120
	completionTok := 400 - brainReads*40
	if completionTok < 60 {
		completionTok = 60
	}
	totalTok := promptTok + completionTok

	transcript := strings.Join(lines, "\n") + "\n"
	toolCalls := map[string]int{
		"Read":  len(req.InputFiles) + brainReads,
		"Write": 1,
		"Bash":  1,
	}

	// Metrics sidecar file so scripted assertions can check tool call
	// counts without parsing the prose transcript.
	metrics := map[string]any{
		"tool_calls":  toolCalls,
		"brain_reads": brainReads,
	}
	_ = writeJSON(filepath.Join(req.OutputDir, "metrics.json"), metrics)

	duration := time.Since(start)
	if duration < 5*time.Millisecond {
		duration = 5 * time.Millisecond // floor so tests don't see 0
	}

	return &runner.Result{
		PromptTokens:     promptTok,
		CompletionTokens: completionTok,
		TotalTokens:      totalTok,
		CostUSD:          0,
		DurationMs:       duration.Milliseconds(),
		Transcript:       []byte(transcript),
		ToolCalls:        toolCalls,
		OutputFiles:      []string{"mock-output.txt"},
	}, nil
}

func oneLine(s string, max int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
