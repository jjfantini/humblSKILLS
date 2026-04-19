// Package grader evaluates assertions against run outputs.
//
// Two pathways:
//
//  1. Scripted: path_exists / exec / script / regex / json_valid are checked
//     in-process (or by shelling out to a provided command). Deterministic
//     and reusable - preferred over LLM judgment per Anthropic's guide.
//  2. LLM judge: "llm" assertions are batched into a single call to a grader
//     model. The grader receives the eval prompt, the transcript, the output
//     file tree, and the list of LLM assertions. It returns per-assertion
//     {text, passed, evidence} per Anthropic's agents/grader.md contract.
//
// Grading output mirrors the Anthropic schema so downstream tooling (report
// viewer, aggregate scripts) works unchanged.
package grader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
)

// Grading is the top-level grading.json shape. Matches Anthropic exactly.
type Grading struct {
	Expectations     []ExpectationResult `json:"expectations"`
	Summary          Summary             `json:"summary"`
	ExecutionMetrics *ExecutionMetrics   `json:"execution_metrics,omitempty"`
	Timing           *TimingBlock        `json:"timing,omitempty"`
}

// ExpectationResult is one graded assertion. Field names match Anthropic.
type ExpectationResult struct {
	Text     string `json:"text"`
	Passed   bool   `json:"passed"`
	Evidence string `json:"evidence"`
}

// Summary rolls up per-grading pass counts.
type Summary struct {
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Total    int     `json:"total"`
	PassRate float64 `json:"pass_rate"`
}

// ExecutionMetrics is carried through from the runner's transcript.
type ExecutionMetrics struct {
	ToolCalls      map[string]int `json:"tool_calls,omitempty"`
	TotalToolCalls int            `json:"total_tool_calls,omitempty"`
	TotalSteps     int            `json:"total_steps,omitempty"`
	Errors         int            `json:"errors_encountered,omitempty"`
	OutputChars    int            `json:"output_chars,omitempty"`
	TranscriptChars int           `json:"transcript_chars,omitempty"`
}

// TimingBlock is copied straight from the run's timing.json.
type TimingBlock struct {
	ExecutorDurationSec float64 `json:"executor_duration_seconds,omitempty"`
	GraderDurationSec   float64 `json:"grader_duration_seconds,omitempty"`
	TotalDurationSec    float64 `json:"total_duration_seconds,omitempty"`
}

// Request is what the caller hands Grade. WorkDir is where scripted
// assertions resolve relative paths against (typically the session's
// OutputDir's parent so `outputs/foo.txt` style references work).
type Request struct {
	EvalPrompt  string
	Assertions  []scenarios.Assertion
	OutputDir   string        // where Execute wrote output files (outputs/)
	WorkDir     string        // resolution base for scripted checks (session dir)
	Transcript  []byte
	LLMJudge    LLMJudge      // nil -> "llm" assertions auto-fail with a clear evidence note
	ExecTimeout time.Duration // per scripted check; 0 = 30s
	ToolCalls   map[string]int
}

// LLMJudge abstracts the grader LLM so unit tests can substitute a stub.
// The real implementation lives in the Anthropic runner (it reuses the
// same SDK credentials).
type LLMJudge interface {
	Grade(ctx context.Context, prompt string, transcript []byte, outputs string,
		assertions []scenarios.Assertion) ([]ExpectationResult, error)
}

// Grade runs every assertion, returning a single Grading.
func Grade(ctx context.Context, r Request) (*Grading, error) {
	if r.ExecTimeout == 0 {
		r.ExecTimeout = 30 * time.Second
	}
	graderStart := time.Now()

	scripted := make([]int, 0, len(r.Assertions))
	llmIdx := make([]int, 0, len(r.Assertions))
	for i, a := range r.Assertions {
		k, _ := scenarios.ParseCheck(a.Check)
		if k == scenarios.CheckLLM {
			llmIdx = append(llmIdx, i)
		} else {
			scripted = append(scripted, i)
		}
	}

	results := make([]ExpectationResult, len(r.Assertions))
	for _, i := range scripted {
		results[i] = gradeScripted(ctx, r, r.Assertions[i])
	}
	if len(llmIdx) > 0 {
		if r.LLMJudge == nil {
			for _, i := range llmIdx {
				results[i] = ExpectationResult{
					Text:     r.Assertions[i].Text,
					Passed:   false,
					Evidence: "LLM judge not configured (set ANTHROPIC_API_KEY or use scripted checks).",
				}
			}
		} else {
			picks := make([]scenarios.Assertion, 0, len(llmIdx))
			for _, i := range llmIdx {
				picks = append(picks, r.Assertions[i])
			}
			outputsTree := describeOutputs(r.OutputDir)
			llmRes, err := r.LLMJudge.Grade(ctx, r.EvalPrompt, r.Transcript, outputsTree, picks)
			if err != nil {
				for _, i := range llmIdx {
					results[i] = ExpectationResult{
						Text:     r.Assertions[i].Text,
						Passed:   false,
						Evidence: "LLM grader error: " + err.Error(),
					}
				}
			} else if len(llmRes) != len(llmIdx) {
				return nil, fmt.Errorf("LLM returned %d results, expected %d", len(llmRes), len(llmIdx))
			} else {
				for k, i := range llmIdx {
					results[i] = llmRes[k]
				}
			}
		}
	}

	g := &Grading{Expectations: results}
	for _, e := range results {
		g.Summary.Total++
		if e.Passed {
			g.Summary.Passed++
		} else {
			g.Summary.Failed++
		}
	}
	if g.Summary.Total > 0 {
		g.Summary.PassRate = float64(g.Summary.Passed) / float64(g.Summary.Total)
	}
	g.ExecutionMetrics = &ExecutionMetrics{
		ToolCalls:       r.ToolCalls,
		TotalToolCalls:  sum(r.ToolCalls),
		TranscriptChars: len(r.Transcript),
		OutputChars:     int(outputsSize(r.OutputDir)),
	}
	g.Timing = &TimingBlock{
		GraderDurationSec: time.Since(graderStart).Seconds(),
	}
	return g, nil
}

// Write emits g as prettified JSON to path. Atomic via temp+rename.
func Write(path string, g *Grading) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// --- scripted assertions ----------------------------------------------------

func gradeScripted(ctx context.Context, r Request, a scenarios.Assertion) ExpectationResult {
	kind, arg := scenarios.ParseCheck(a.Check)
	switch kind {
	case scenarios.CheckPathExists:
		target := resolveRel(r, arg)
		if _, err := os.Stat(target); err == nil {
			return pass(a.Text, "path exists: "+target)
		}
		return fail(a.Text, "path not found: "+target)
	case scenarios.CheckExec:
		return gradeExec(ctx, r, a, arg)
	case scenarios.CheckScript:
		return gradeExec(ctx, r, a, "bash "+arg)
	case scenarios.CheckRegex:
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			return fail(a.Text, "malformed regex check: "+arg)
		}
		target := resolveRel(r, parts[0])
		body, err := os.ReadFile(target)
		if err != nil {
			return fail(a.Text, "read target: "+err.Error())
		}
		// Default to multi-line mode so "^...$" matches any line, which is
		// what users intuitively expect for document-level regex checks.
		pat := parts[1]
		if !strings.HasPrefix(pat, "(?") {
			pat = "(?m)" + pat
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return fail(a.Text, "invalid regex: "+err.Error())
		}
		if re.Match(body) {
			return pass(a.Text, "matched in "+target)
		}
		return fail(a.Text, "no match in "+target)
	case scenarios.CheckJSONValid:
		target := resolveRel(r, arg)
		body, err := os.ReadFile(target)
		if err != nil {
			return fail(a.Text, "read: "+err.Error())
		}
		var v any
		if err := json.Unmarshal(body, &v); err != nil {
			return fail(a.Text, "invalid JSON: "+err.Error())
		}
		return pass(a.Text, "valid JSON in "+target)
	default:
		return fail(a.Text, "unknown check kind: "+string(kind))
	}
}

func gradeExec(ctx context.Context, r Request, a scenarios.Assertion, cmdline string) ExpectationResult {
	ctx2, cancel := context.WithTimeout(ctx, r.ExecTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx2, "sh", "-c", cmdline)
	cmd.Dir = r.WorkDir
	// Make the OUTPUT_DIR available to scripts.
	cmd.Env = append(os.Environ(), "EVAL_OUTPUT_DIR="+r.OutputDir)
	out, err := cmd.CombinedOutput()
	tail := lastLines(string(out), 3)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fail(a.Text, "timed out: "+cmdline)
		}
		return fail(a.Text, fmt.Sprintf("exit != 0 (%v) — %s", err, tail))
	}
	return pass(a.Text, fmt.Sprintf("exit 0 — %s", tail))
}

// --- helpers ----------------------------------------------------------------

func resolveRel(r Request, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	// Prefer OutputDir so "outputs/foo" and "foo" both work.
	out := filepath.Join(r.OutputDir, rel)
	if _, err := os.Stat(out); err == nil {
		return out
	}
	return filepath.Join(r.WorkDir, rel)
}

func pass(text, ev string) ExpectationResult {
	return ExpectationResult{Text: text, Passed: true, Evidence: ev}
}

func fail(text, ev string) ExpectationResult {
	return ExpectationResult{Text: text, Passed: false, Evidence: ev}
}

func sum(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.TrimSpace(strings.Join(lines, " | "))
}

func outputsSize(dir string) int64 {
	var total int64
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

func describeOutputs(dir string) string {
	var sb strings.Builder
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		info, err := d.Info()
		if err != nil {
			return nil
		}
		fmt.Fprintf(&sb, "%s (%d bytes)\n", rel, info.Size())
		return nil
	})
	return sb.String()
}
