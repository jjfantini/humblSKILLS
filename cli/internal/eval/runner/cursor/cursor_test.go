package cursor

import (
	"os"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
)

func TestNew(t *testing.T) {
	r := New()
	if r.Name() != "cursor-agent" {
		t.Errorf("Name = %q", r.Name())
	}
	if r.Capabilities().DefaultModel != "sonnet-4" {
		t.Errorf("DefaultModel = %q", r.Capabilities().DefaultModel)
	}
}

func TestArgs_Contains_P_F_OutputFormat(t *testing.T) {
	r := New()
	args := r.D.Args(runner.Request{Prompt: "hi"}, "/s", "/s/p")
	hasP, hasF, hasFormat := false, false, false
	for _, a := range args {
		switch a {
		case "-p":
			hasP = true
		case "-f":
			hasF = true
		case "stream-json":
			hasFormat = true
		}
	}
	if !hasP || !hasF || !hasFormat {
		t.Errorf("missing flag(s): %v", args)
	}
}

func TestBuildPrompt_PrependsSkillPreamble(t *testing.T) {
	got := buildPrompt(runner.Request{Prompt: "base", SkillDir: "/x"})
	if !strings.Contains(got, "./skill/SKILL.md") {
		t.Errorf("missing skill preamble:\n%s", got)
	}
	if !strings.HasSuffix(got, "base") {
		t.Errorf("base prompt missing/relocated:\n%s", got)
	}
}

func TestBuildPrompt_NoPreambleWithoutSkill(t *testing.T) {
	got := buildPrompt(runner.Request{Prompt: "base"})
	if strings.Contains(got, "SKILL.md") {
		t.Errorf("unexpected preamble:\n%s", got)
	}
}

func TestBuildPrompt_MentionsInputsWhenPresent(t *testing.T) {
	got := buildPrompt(runner.Request{Prompt: "p", InputFiles: []string{"/a", "/b"}})
	if !strings.Contains(got, "./inputs/") {
		t.Errorf("missing inputs mention:\n%s", got)
	}
}

func TestArgs_ModelFlag(t *testing.T) {
	r := New()
	args := r.D.Args(runner.Request{Prompt: "p", Model: "gpt-5"}, "/s", "/s/p")
	has := false
	for i, a := range args {
		if a == "--model" && i+1 < len(args) && args[i+1] == "gpt-5" {
			has = true
		}
	}
	if !has {
		t.Errorf("missing --model: %v", args)
	}
}

func TestArgs_ExtraArgsFromEnv(t *testing.T) {
	t.Setenv("CURSOR_AGENT_EXTRA_ARGS", "--vendor-flag")
	defer os.Unsetenv("CURSOR_AGENT_EXTRA_ARGS")

	r := New()
	args := r.D.Args(runner.Request{Prompt: "p"}, "/s", "/s/p")
	has := false
	for _, a := range args {
		if a == "--vendor-flag" {
			has = true
		}
	}
	if !has {
		t.Errorf("missing extra arg: %v", args)
	}
}

func TestParseEvent_ToolCallWithToolKey(t *testing.T) {
	ev := parseEvent([]byte(`{"type":"tool_call","tool":"Read"}`))
	if ev.ToolName != "Read" {
		t.Errorf("got %q", ev.ToolName)
	}
}

func TestParseEvent_ToolCallWithNameKey(t *testing.T) {
	ev := parseEvent([]byte(`{"type":"tool_use","name":"Write"}`))
	if ev.ToolName != "Write" {
		t.Errorf("got %q", ev.ToolName)
	}
}

func TestParseEvent_CamelCaseUsage(t *testing.T) {
	ev := parseEvent([]byte(`{"usage":{"inputTokens":10,"outputTokens":20,"totalTokens":30}}`))
	if ev.PromptTokensDelta != 10 || ev.CompletionTokensDelta != 20 || ev.TotalTokensDelta != 30 {
		t.Errorf("camelCase not parsed: %+v", ev)
	}
}

func TestParseEvent_SnakeCaseUsage(t *testing.T) {
	ev := parseEvent([]byte(`{"usage":{"input_tokens":5,"output_tokens":6,"total_tokens":11}}`))
	if ev.PromptTokensDelta != 5 || ev.CompletionTokensDelta != 6 || ev.TotalTokensDelta != 11 {
		t.Errorf("snake_case not parsed: %+v", ev)
	}
}

func TestParseEvent_IgnoresGarbage(t *testing.T) {
	ev := parseEvent([]byte("not json"))
	if ev.ToolName != "" || ev.PromptTokensDelta != 0 {
		t.Errorf("garbage produced event: %+v", ev)
	}
}

func TestFirstNonZeroInt(t *testing.T) {
	m := map[string]any{"a": float64(0), "b": float64(5), "c": float64(0)}
	if got := firstNonZeroInt(m, "a", "b", "c"); got != 5 {
		t.Errorf("got %d", got)
	}
	if got := firstNonZeroInt(m, "a", "c"); got != 0 {
		t.Errorf("all-zero: got %d", got)
	}
}
