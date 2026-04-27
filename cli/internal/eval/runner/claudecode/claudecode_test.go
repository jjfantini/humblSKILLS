package claudecode

import (
	"os"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
)

func TestNew_ReturnsClaudecodeRunner(t *testing.T) {
	r := New()
	if r.Name() != "claudecode" {
		t.Errorf("Name = %q", r.Name())
	}
	if r.Capabilities().DefaultModel != "claude-opus-4-7" {
		t.Errorf("DefaultModel = %q", r.Capabilities().DefaultModel)
	}
}

func TestArgs_IncludesPromptAndOutputFormat(t *testing.T) {
	r := New()
	args := r.D.Args(runner.Request{Prompt: "do a thing"}, "/scratch", "/scratch/prompt.txt")
	haveP, haveFormat := false, false
	for i, a := range args {
		if a == "-p" && i+1 < len(args) && args[i+1] == "do a thing" {
			haveP = true
		}
		if a == "--output-format" && i+1 < len(args) && args[i+1] == "stream-json" {
			haveFormat = true
		}
	}
	if !haveP {
		t.Errorf("missing -p: %v", args)
	}
	if !haveFormat {
		t.Errorf("missing --output-format stream-json: %v", args)
	}
}

func TestArgs_IncludesPermissionModeAcceptEdits(t *testing.T) {
	// Headless -p runs need --permission-mode because the default mode asks
	// for per-tool approval on every Write / Edit / Bash, which blocks
	// indefinitely without a user. acceptEdits auto-approves file writes
	// (the only tool the scenario needs) and keeps bash behind a prompt.
	r := New()
	args := r.D.Args(runner.Request{Prompt: "p"}, "/s", "/s/p")
	has := false
	for i, a := range args {
		if a == "--permission-mode" && i+1 < len(args) && args[i+1] == "acceptEdits" {
			has = true
		}
	}
	if !has {
		t.Errorf("missing --permission-mode acceptEdits: %v", args)
	}
}

func TestArgs_ModelFlag(t *testing.T) {
	r := New()
	args := r.D.Args(runner.Request{Prompt: "p", Model: "claude-sonnet-4-6"}, "/s", "/s/p")
	has := false
	for i, a := range args {
		if a == "--model" && i+1 < len(args) && args[i+1] == "claude-sonnet-4-6" {
			has = true
		}
	}
	if !has {
		t.Errorf("missing --model: %v", args)
	}
}

func TestArgs_ExtraArgsFromEnv(t *testing.T) {
	t.Setenv("CLAUDE_CODE_EXTRA_ARGS", "--foo bar --baz")
	defer os.Unsetenv("CLAUDE_CODE_EXTRA_ARGS")

	r := New()
	args := r.D.Args(runner.Request{Prompt: "p"}, "/s", "/s/p")
	found := 0
	for _, a := range args {
		if a == "--foo" || a == "bar" || a == "--baz" {
			found++
		}
	}
	if found != 3 {
		t.Errorf("extra args missing: got %v", args)
	}
}

func TestParseEvent_ToolUse(t *testing.T) {
	ev := parseEvent([]byte(`{"type":"tool_use","name":"Read"}`))
	if ev.ToolName != "Read" {
		t.Errorf("ToolName = %q", ev.ToolName)
	}
}

func TestParseEvent_Usage(t *testing.T) {
	ev := parseEvent([]byte(`{"type":"result","usage":{"input_tokens":123,"output_tokens":456}}`))
	if ev.PromptTokensDelta != 123 {
		t.Errorf("PromptTokensDelta = %d", ev.PromptTokensDelta)
	}
	if ev.CompletionTokensDelta != 456 {
		t.Errorf("CompletionTokensDelta = %d", ev.CompletionTokensDelta)
	}
}

func TestParseEvent_IgnoresNonJSON(t *testing.T) {
	ev := parseEvent([]byte("garbage line"))
	if ev.ToolName != "" || ev.PromptTokensDelta != 0 {
		t.Errorf("non-JSON produced event: %+v", ev)
	}
}

func TestParseEvent_IgnoresUnknownType(t *testing.T) {
	// type="user_input" isn't recognized — event stays zero.
	ev := parseEvent([]byte(`{"type":"user_input","text":"hi"}`))
	if ev.ToolName != "" {
		t.Errorf("unexpected ToolName: %q", ev.ToolName)
	}
}

func TestIntField_HandlesFloatAndInt(t *testing.T) {
	if got := intField(map[string]any{"x": float64(42)}, "x"); got != 42 {
		t.Errorf("float64: got %d", got)
	}
	if got := intField(map[string]any{"x": 7}, "x"); got != 7 {
		t.Errorf("int: got %d", got)
	}
	if got := intField(map[string]any{}, "missing"); got != 0 {
		t.Errorf("missing: got %d", got)
	}
	if got := intField(map[string]any{"x": "bogus"}, "x"); got != 0 {
		t.Errorf("non-numeric: got %d", got)
	}
}
