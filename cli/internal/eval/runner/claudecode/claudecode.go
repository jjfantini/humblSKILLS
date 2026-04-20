// Package claudecode wraps the `claude` CLI (Claude Code headless mode).
//
// The CLI is invoked as:
//
//	claude -p "<prompt>" --output-format stream-json \
//	       --skill-dir <scratch>/skill \
//	       --output-dir <scratch>/outputs
//
// Exact flags vary by Claude Code version; the driver passes a minimal
// sensible set and relies on the agent to interpret. Users can customize
// via CLAUDE_CODE_EXTRA_ARGS (space-separated) in their profile.
package claudecode

import (
	"os"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/clitool"
)

// New returns a claudecode runner.
func New() *clitool.Runner {
	d := clitool.Driver{
		Name:         "claudecode",
		Binary:       "claude",
		VersionArgs:  []string{"--version"},
		DefaultModel: "claude-opus-4-7",
		Args: func(req runner.Request, scratchDir, promptPath string) []string {
			args := []string{
				"-p", req.Prompt,
				"--output-format", "stream-json",
				"--output-dir", "outputs",
			}
			if req.SkillDir != "" {
				args = append(args, "--skill-dir", "skill")
			}
			if req.Model != "" {
				args = append(args, "--model", req.Model)
			}
			if extra := os.Getenv("CLAUDE_CODE_EXTRA_ARGS"); extra != "" {
				args = append(args, strings.Fields(extra)...)
			}
			return args
		},
		ParseEvent: parseEvent,
	}
	return clitool.New(d)
}

// parseEvent recognizes the subset of Claude Code stream-json events we
// care about: tool_use and the final result event with usage.
func parseEvent(line []byte) clitool.Event {
	m, err := clitool.ParseJSONEvent(line)
	if err != nil || m == nil {
		return clitool.Event{}
	}
	var ev clitool.Event
	if t, _ := m["type"].(string); t == "tool_use" {
		if name, _ := m["name"].(string); name != "" {
			ev.ToolName = name
		}
	}
	if usage, ok := m["usage"].(map[string]any); ok {
		ev.PromptTokensDelta = intField(usage, "input_tokens")
		ev.CompletionTokensDelta = intField(usage, "output_tokens")
	}
	return ev
}

func intField(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	}
	return 0
}
