// Package codex wraps the `codex` CLI (OpenAI's agent). Stream-json shape
// follows the OpenAI Responses API convention.
package codex

import (
	"os"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/clitool"
)

// New returns a codex runner.
func New() *clitool.Runner {
	d := clitool.Driver{
		Name:         "codex",
		Binary:       "codex",
		VersionArgs:  []string{"--version"},
		DefaultModel: "gpt-5-codex",
		Args: func(req runner.Request, scratchDir, promptPath string) []string {
			args := []string{
				"exec", req.Prompt,
				"--json",
			}
			if req.Model != "" {
				args = append(args, "--model", req.Model)
			}
			if extra := os.Getenv("CODEX_EXTRA_ARGS"); extra != "" {
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
	// Codex emits "response.output_item.added" with type "tool_call".
	if t, _ := m["type"].(string); strings.Contains(t, "tool_call") {
		if name, _ := m["name"].(string); name != "" {
			ev.ToolName = name
		}
	}
	if usage, ok := m["usage"].(map[string]any); ok {
		ev.PromptTokensDelta = intField(usage, "input_tokens")
		ev.CompletionTokensDelta = intField(usage, "output_tokens")
		ev.TotalTokensDelta = intField(usage, "total_tokens")
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
