// Package runner defines the Runner interface every eval backend satisfies.
//
// The harness is runner-agnostic: it hands each backend a Request (skill path,
// prompt, input files, output dir) and expects a Result (tokens, duration,
// transcript, tool-call counts, output file list). Six backends ship:
// claudecode, cursor-agent, codex, anthropic-api, openai-api, mock.
package runner

import (
	"context"
	"time"
)

// Runner is the contract every eval backend satisfies.
type Runner interface {
	// Name is the stable ID used in config / JSON / UIs
	// (e.g. "claudecode", "anthropic-api", "mock").
	Name() string

	// Capabilities reports what this backend supports.
	Capabilities() Capabilities

	// DoctorCheck probes runtime availability (binary on PATH, API key
	// present, etc.) without making real calls. Intended for fast UI
	// rendering; never burns tokens.
	DoctorCheck(ctx context.Context) DoctorCheck

	// Execute runs the prompt against the agent, materializes output files
	// under req.OutputDir, and returns token + timing metadata. Errors are
	// reported on Result.Err rather than the returned error so callers can
	// still capture a partial transcript.
	Execute(ctx context.Context, req Request) (*Result, error)
}

// Capabilities advertises what a backend can do. Used by the config TUI to
// gray-out incompatible options (e.g. "parallel workers" for a backend that
// serializes).
type Capabilities struct {
	SupportsTools    []string // "Read","Write","Bash","Glob","Grep","Skill"
	SupportsParallel bool     // can multiple Execute() run concurrently?
	DefaultModel     string   // suggested model when caller passes ""
	AvailableModels  []string // optional menu; empty means "anything goes"
	Pricing          *Pricing // best-effort cost estimate; may be nil
}

// Pricing is a tiny per-million-token schedule. Callers multiply against
// prompt/completion counts to derive Result.CostUSD.
type Pricing struct {
	PromptUSDPerMtok     float64
	CompletionUSDPerMtok float64
}

// DoctorCheck is the per-runner health probe result.
type DoctorCheck struct {
	Available   bool
	Version     string // "claude-code 2.1.0", "anthropic-sdk-go v0.3.0", ...
	Reason      string // why unavailable, human-readable
	Fix         string // one-line remediation for static + TUI doctor views
	RequiresKey string // "anthropic" | "openai" | "" if none
}

// Request is what the harness hands to a runner.
type Request struct {
	// SkillDir is the absolute path to the skill directory. Empty for the
	// no_skill arm.
	SkillDir string

	// Prompt is the user-facing prompt for this session.
	Prompt string

	// InputFiles are absolute paths that should be made available in the
	// scratch cwd (typically under ./inputs/). Runners copy rather than
	// symlink so the agent can mutate them safely.
	InputFiles []string

	// OutputDir is where the runner must write any emitted output files
	// and the transcript. Guaranteed to exist when Execute runs.
	OutputDir string

	// Model overrides Capabilities.DefaultModel. Empty means "use default".
	Model string

	// SystemPrompt is the optional top-of-context instruction. The harness
	// sets this to the skill's SKILL.md body for API-backed runners so the
	// skill content is actually loaded.
	SystemPrompt string

	// AllowedTools restricts the tool surface. Empty means all advertised
	// tools are allowed.
	AllowedTools []string

	// Timeout caps wall-clock duration for the whole session.
	Timeout time.Duration
}

// Result is what every runner returns.
type Result struct {
	PromptTokens     int            `json:"prompt_tokens"`
	CompletionTokens int            `json:"completion_tokens"`
	TotalTokens      int            `json:"total_tokens"`
	CostUSD          float64        `json:"cost_usd"`
	DurationMs       int64          `json:"duration_ms"`
	Transcript       []byte         `json:"-"` // written to transcript.txt alongside
	ToolCalls        map[string]int `json:"tool_calls,omitempty"`
	OutputFiles      []string       `json:"output_files,omitempty"`
	Err              error          `json:"-"`
}

// EstimateCost multiplies token counts against pricing. Safe against nil
// pricing (returns 0). Keeps the cost-calc logic in one place so every
// backend reports consistent numbers.
func EstimateCost(p *Pricing, promptTok, completionTok int) float64 {
	if p == nil {
		return 0
	}
	return (float64(promptTok)*p.PromptUSDPerMtok +
		float64(completionTok)*p.CompletionUSDPerMtok) / 1_000_000.0
}
