// Package cursor wraps the `cursor-agent` CLI. Command shape mirrors
// Claude Code but with Cursor-specific flag names.
package cursor

import (
	"os"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/clitool"
)

// New returns a cursor-agent runner.
func New() *clitool.Runner {
	d := clitool.Driver{
		Name:         "cursor-agent",
		Binary:       "cursor-agent",
		VersionArgs:  []string{"--version"},
		DefaultModel: "auto",
		Args: func(req runner.Request, scratchDir, promptPath string) []string {
			args := []string{
				"--prompt", req.Prompt,
				"--output-format", "json",
				"--output-dir", "outputs",
			}
			if req.SkillDir != "" {
				args = append(args, "--skill", "skill")
			}
			if req.Model != "" {
				args = append(args, "--model", req.Model)
			}
			if extra := os.Getenv("CURSOR_AGENT_EXTRA_ARGS"); extra != "" {
				args = append(args, strings.Fields(extra)...)
			}
			return args
		},
		ParseEvent: parseEvent,
	}
	return clitool.New(d)
}

func parseEvent(line []byte) clitool.Event {
	m, err := clitool.ParseJSONEvent(line)
	if err != nil || m == nil {
		return clitool.Event{}
	}
	var ev clitool.Event
	if t, _ := m["event"].(string); t == "tool_call" {
		if name, _ := m["tool"].(string); name != "" {
			ev.ToolName = name
		}
	}
	if usage, ok := m["usage"].(map[string]any); ok {
		ev.PromptTokensDelta = intField(usage, "prompt_tokens")
		ev.CompletionTokensDelta = intField(usage, "completion_tokens")
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
