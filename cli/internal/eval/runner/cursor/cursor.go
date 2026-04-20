// Package cursor wraps the `cursor-agent` CLI.
//
// Command shape (from `cursor-agent --help`):
//
//	cursor-agent -p --output-format stream-json -f --model <model> "<prompt>"
//
// The prompt is positional. -p enables non-interactive mode, -f auto-
// approves tool calls (writes + shell), --output-format stream-json gives
// us machine-parseable events with token usage.
//
// The runner runs with cwd set to a scratch dir that already contains
// `skill/` (the Smart Skill) and `inputs/` (staged input files). For the
// smart_skill / flat_skill arms we prepend a short preamble to the prompt
// pointing the agent at `./skill/SKILL.md`, because cursor-agent has no
// native "load this skill" flag.
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
		DefaultModel: "sonnet-4",
		Args: func(req runner.Request, scratchDir, promptPath string) []string {
			args := []string{
				"-p",
				"--output-format", "stream-json",
				"-f", // auto-approve writes + shell
			}
			if req.Model != "" {
				args = append(args, "--model", req.Model)
			}
			if extra := os.Getenv("CURSOR_AGENT_EXTRA_ARGS"); extra != "" {
				args = append(args, strings.Fields(extra)...)
			}
			args = append(args, buildPrompt(req))
			return args
		},
		ParseEvent: parseEvent,
	}
	return clitool.New(d)
}

// buildPrompt prepends a short system-ish preamble so cursor-agent knows
// where the skill lives (when applicable) and where to write its output.
func buildPrompt(req runner.Request) string {
	var sb strings.Builder
	if req.SkillDir != "" {
		sb.WriteString("You have access to a skill at ./skill/. Read ./skill/SKILL.md FIRST and use it as your primary guidance. For Smart Skills, also read ./skill/references/_index.md, patterns.md, decisions.md, log.md, and relevant wiki/ concepts before producing output.\n\n")
	}
	if len(req.InputFiles) > 0 {
		sb.WriteString("Input files are staged under ./inputs/. Read them as needed.\n\n")
	}
	sb.WriteString(req.Prompt)
	return sb.String()
}

func parseEvent(line []byte) clitool.Event {
	m, err := clitool.ParseJSONEvent(line)
	if err != nil || m == nil {
		return clitool.Event{}
	}
	var ev clitool.Event
	// cursor-agent stream-json emits envelopes with a `type` field and
	// tool calls via `type: "tool_call"` or `type: "tool_use"`.
	t, _ := m["type"].(string)
	if strings.Contains(t, "tool") {
		if name, ok := m["tool"].(string); ok && name != "" {
			ev.ToolName = name
		} else if name, ok := m["name"].(string); ok && name != "" {
			ev.ToolName = name
		}
	}
	if usage, ok := m["usage"].(map[string]any); ok {
		// cursor-agent emits camelCase: inputTokens, outputTokens,
		// cacheReadTokens, cacheWriteTokens. Fall back to snake_case
		// variants for future compatibility.
		ev.PromptTokensDelta = firstNonZeroInt(usage, "inputTokens", "input_tokens", "prompt_tokens")
		ev.CompletionTokensDelta = firstNonZeroInt(usage, "outputTokens", "output_tokens", "completion_tokens")
		ev.TotalTokensDelta = firstNonZeroInt(usage, "totalTokens", "total_tokens")
	}
	return ev
}

func firstNonZeroInt(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if n := intField(m, k); n != 0 {
			return n
		}
	}
	return 0
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
