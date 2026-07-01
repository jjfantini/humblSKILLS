// Package clitool is the common scaffolding used by all three agent-CLI
// runners (claudecode, cursor-agent, codex). Each one shells out to a
// third-party binary, feeds it a prompt, and extracts tokens + duration
// from the output. The only differences between them are:
//
//   - binary name + version flag
//   - command-line syntax (prompt-positional, flag-based, stdin-pipe)
//   - how they surface token usage in their stream-json output
//
// This package captures the common glue (scratch-cwd setup, input-file
// staging, stream capture, transcript writing). Each runner supplies a
// Driver implementation describing its specific shape.
package clitool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/fsutil"
)

// Driver abstracts over the specific CLI's command shape.
type Driver struct {
	// Name is the runner name reported via Runner.Name().
	Name string

	// Binary is the executable name searched on $PATH.
	Binary string

	// VersionArgs is how to probe for the version string
	// (e.g. []string{"--version"}).
	VersionArgs []string

	// Pricing is optional; supplies cost estimates when present.
	Pricing *runner.Pricing

	// DefaultModel is returned via Capabilities.
	DefaultModel string

	// Args builds the argv this CLI wants for a given request. Receives
	// the scratch cwd and the staged prompt file path. Returns argv
	// minus the leading binary name.
	Args func(req runner.Request, scratchDir, promptPath string) []string

	// Stdin, if non-nil, is fed to the CLI on stdin.
	Stdin func(req runner.Request) []byte

	// ParseEvent consumes one line of stream-json output and returns
	// token deltas / tool call counts when recognized. Unknown lines
	// return zeros and nil err - the package silently skips them.
	ParseEvent func(line []byte) Event

	// RequiresKey is the secrets provider this CLI needs, if any
	// (typically empty - the CLI manages its own login).
	RequiresKey string
}

// Event is what ParseEvent returns per stream line.
type Event struct {
	PromptTokensDelta     int
	CompletionTokensDelta int
	TotalTokensDelta      int
	ToolName              string // "" means not a tool-call event
}

// Runner wraps a Driver to satisfy runner.Runner.
type Runner struct{ D Driver }

// New returns a new CLI-backed runner.
func New(d Driver) *Runner { return &Runner{D: d} }

func (r *Runner) Name() string { return r.D.Name }

func (r *Runner) Capabilities() runner.Capabilities {
	return runner.Capabilities{
		SupportsTools:    []string{"Read", "Write", "Bash", "Glob", "Grep"},
		SupportsParallel: true,
		DefaultModel:     r.D.DefaultModel,
		Pricing:          r.D.Pricing,
	}
}

func (r *Runner) DoctorCheck(ctx context.Context) runner.DoctorCheck {
	path, err := exec.LookPath(r.D.Binary)
	if err != nil {
		return runner.DoctorCheck{
			Available: false,
			Reason:    r.D.Binary + " not in PATH",
			Fix:       "install " + r.D.Binary + " and ensure it's on your PATH",
			RequiresKey: r.D.RequiresKey,
		}
	}
	var ver string
	if len(r.D.VersionArgs) > 0 {
		ctx2, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		out, _ := exec.CommandContext(ctx2, path, r.D.VersionArgs...).Output()
		ver = strings.TrimSpace(firstLine(string(out)))
	}
	return runner.DoctorCheck{Available: true, Version: r.D.Binary + " " + ver, RequiresKey: r.D.RequiresKey}
}

// Execute stages the skill, stages input files, runs the CLI, tees the
// stream into the transcript, and returns the final Result.
func (r *Runner) Execute(ctx context.Context, req runner.Request) (*runner.Result, error) {
	start := time.Now()

	// Scratch dir lives alongside OutputDir so relative paths the CLI
	// creates ("./out.txt") remain inspectable by the grader.
	scratch, err := os.MkdirTemp(filepath.Dir(req.OutputDir), "scratch-")
	if err != nil {
		return nil, fmt.Errorf("create scratch: %w", err)
	}
	defer os.RemoveAll(scratch)

	// Stage skill (claudecode et al read skill from a known subdir).
	if req.SkillDir != "" {
		dst := filepath.Join(scratch, "skill")
		if err := fsutil.CopyTree(req.SkillDir, dst, fsutil.Options{}); err != nil {
			return nil, fmt.Errorf("stage skill: %w", err)
		}
	}
	// Stage input files under inputs/.
	inputs := filepath.Join(scratch, "inputs")
	if err := os.MkdirAll(inputs, 0o755); err != nil {
		return nil, err
	}
	for _, f := range req.InputFiles {
		if err := copyFile(f, filepath.Join(inputs, filepath.Base(f))); err != nil {
			return nil, fmt.Errorf("stage input %s: %w", f, err)
		}
	}

	promptPath := filepath.Join(scratch, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte(req.Prompt), 0o644); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return nil, err
	}

	args := r.D.Args(req, scratch, promptPath)
	ctx2 := ctx
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx2, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx2, r.D.Binary, args...)
	cmd.Dir = scratch
	if r.D.Stdin != nil {
		cmd.Stdin = bytes.NewReader(r.D.Stdin(req))
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	result := &runner.Result{ToolCalls: map[string]int{}}
	var transcript bytes.Buffer
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		transcript.Write(line)
		transcript.WriteByte('\n')
		if r.D.ParseEvent != nil {
			ev := r.D.ParseEvent(line)
			result.PromptTokens += ev.PromptTokensDelta
			result.CompletionTokens += ev.CompletionTokensDelta
			result.TotalTokens += ev.TotalTokensDelta
			if ev.ToolName != "" {
				result.ToolCalls[ev.ToolName]++
			}
		}
	}
	if err := cmd.Wait(); err != nil {
		// Stderr is the primary failure signal for most agent CLIs.
		result.Err = fmt.Errorf("%s failed: %w — %s", r.D.Binary, err, strings.TrimSpace(stderr.String()))
	}
	// Some drivers only surface totals, others only deltas. If totals
	// are still zero but we saw deltas, fill it in.
	if result.TotalTokens == 0 {
		result.TotalTokens = result.PromptTokens + result.CompletionTokens
	}
	if result.CostUSD == 0 {
		result.CostUSD = runner.EstimateCost(r.D.Pricing, result.PromptTokens, result.CompletionTokens)
	}
	result.DurationMs = time.Since(start).Milliseconds()
	result.Transcript = transcript.Bytes()

	// Collect any files the agent wrote. Some CLIs honour an --output-dir
	// flag (we create scratch/outputs/ for those); others write straight
	// into the cwd. Walk the whole scratch tree but skip the staged
	// inputs + skill + prompt.
	if files, err := collectScratchOutputs(scratch, req.OutputDir); err == nil {
		result.OutputFiles = files
	}
	// Sync the brain (scratch/skill/references/) back to req.SkillDir so
	// any lessons the agent appended to patterns.md / decisions.md /
	// log.md survive past this session. The harness's brain.Snapshot()
	// then captures the real post-session state.
	//
	// Without this, the agent writes into the scratch-private copy, we
	// rm-rf scratch, and every session starts fresh - compounding never
	// compounds. Scope the sync to `references/` to avoid surprising
	// mutations to SKILL.md or scripts/.
	if req.SkillDir != "" {
		src := filepath.Join(scratch, "skill", "references")
		dst := filepath.Join(req.SkillDir, "references")
		if _, err := os.Stat(src); err == nil {
			_ = syncTree(src, dst)
		}
	}
	// Also persist the stderr tail so the grader has something useful
	// when the CLI exited non-zero.
	if stderr.Len() > 0 {
		_ = os.WriteFile(filepath.Join(req.OutputDir, "stderr.txt"), stderr.Bytes(), 0o644)
	}
	return result, nil
}

// --- helpers ----------------------------------------------------------------

func firstLine(s string) string {
	if i := strings.Index(s, "\n"); i >= 0 {
		return s[:i]
	}
	return s
}

// collectScratchOutputs walks the whole scratch directory and copies any
// file the agent created into OutputDir, preserving relative paths.
// Staged inputs (inputs/), the staged skill (skill/), the prompt file,
// and the conventional outputs/ subdir are all handled: `inputs/` and
// `skill/` are skipped; files under `outputs/` are flattened into
// OutputDir (so `outputs/foo.md` → OutputDir/foo.md); everything else
// keeps its path relative to scratch.
func collectScratchOutputs(scratch, dst string) ([]string, error) {
	var out []string
	err := filepath.Walk(scratch, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(scratch, p)
		// Skip staged inputs/, the staged skill, and the prompt file.
		if rel == "prompt.txt" ||
			strings.HasPrefix(rel, "inputs"+string(os.PathSeparator)) ||
			strings.HasPrefix(rel, "skill"+string(os.PathSeparator)) {
			return nil
		}
		// Flatten outputs/foo → foo so scripts can reference paths
		// relative to OutputDir regardless of runner convention.
		target := rel
		if strings.HasPrefix(target, "outputs"+string(os.PathSeparator)) {
			target = strings.TrimPrefix(target, "outputs"+string(os.PathSeparator))
		}
		dstPath := filepath.Join(dst, target)
		if err := copyFile(p, dstPath); err != nil {
			return err
		}
		out = append(out, target)
		return nil
	})
	return out, err
}

// syncTree mirrors src into dst by replacing dst entirely. Used to sync
// the agent-mutated `scratch/skill/references/` back to the harness's
// persistent working copy so brain updates (patterns.md, log.md,
// decisions.md, wiki additions) survive past the session.
func syncTree(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return fsutil.CopyTree(src, dst, fsutil.Options{})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if info, err := os.Stat(src); err == nil {
		_ = os.Chmod(dst, info.Mode())
	}
	return nil
}

// ParseJSONEvent is a helper drivers can use to peel a JSON line. Returns
// (nil, nil) when the line isn't JSON so the caller skips it.
func ParseJSONEvent(line []byte) (map[string]any, error) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 || line[0] != '{' {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return nil, errors.New("not json")
	}
	return m, nil
}
