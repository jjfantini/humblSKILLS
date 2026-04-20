package codex

import (
	"os"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
)

func TestNew(t *testing.T) {
	r := New()
	if r.Name() != "codex" {
		t.Errorf("Name = %q", r.Name())
	}
	if r.Capabilities().DefaultModel != "gpt-5-codex" {
		t.Errorf("DefaultModel = %q", r.Capabilities().DefaultModel)
	}
}

func TestArgs_ExecAndJSONFlag(t *testing.T) {
	r := New()
	args := r.D.Args(runner.Request{Prompt: "do a thing"}, "/s", "/s/p")
	if len(args) < 3 {
		t.Fatalf("args too short: %v", args)
	}
	if args[0] != "exec" {
		t.Errorf("first arg = %q, want exec", args[0])
	}
	if args[1] != "do a thing" {
		t.Errorf("second arg = %q", args[1])
	}
	hasJSON := false
	for _, a := range args {
		if a == "--json" {
			hasJSON = true
		}
	}
	if !hasJSON {
		t.Errorf("missing --json: %v", args)
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
	t.Setenv("CODEX_EXTRA_ARGS", "--zap")
	defer os.Unsetenv("CODEX_EXTRA_ARGS")

	r := New()
	args := r.D.Args(runner.Request{Prompt: "p"}, "/s", "/s/p")
	has := false
	for _, a := range args {
		if a == "--zap" {
			has = true
		}
	}
	if !has {
		t.Errorf("missing --zap: %v", args)
	}
}

func TestParseEvent_ToolCall(t *testing.T) {
	// Codex emits "response.output_item.added" with `type` containing "tool_call".
	ev := parseEvent([]byte(`{"type":"response.tool_call.delta","name":"Read"}`))
	if ev.ToolName != "Read" {
		t.Errorf("got %q", ev.ToolName)
	}
}

func TestParseEvent_Usage(t *testing.T) {
	ev := parseEvent([]byte(`{"usage":{"input_tokens":11,"output_tokens":22,"total_tokens":33}}`))
	if ev.PromptTokensDelta != 11 {
		t.Errorf("PromptTokensDelta = %d", ev.PromptTokensDelta)
	}
	if ev.CompletionTokensDelta != 22 {
		t.Errorf("CompletionTokensDelta = %d", ev.CompletionTokensDelta)
	}
	if ev.TotalTokensDelta != 33 {
		t.Errorf("TotalTokensDelta = %d", ev.TotalTokensDelta)
	}
}

func TestParseEvent_NonJSONIgnored(t *testing.T) {
	ev := parseEvent([]byte("not json"))
	if ev.ToolName != "" || ev.PromptTokensDelta != 0 {
		t.Errorf("non-json produced: %+v", ev)
	}
}

func TestIntField(t *testing.T) {
	m := map[string]any{"a": float64(5), "b": 3, "c": "bogus"}
	if got := intField(m, "a"); got != 5 {
		t.Errorf("a = %d", got)
	}
	if got := intField(m, "b"); got != 3 {
		t.Errorf("b = %d", got)
	}
	if got := intField(m, "c"); got != 0 {
		t.Errorf("c = %d", got)
	}
	if got := intField(m, "missing"); got != 0 {
		t.Errorf("missing = %d", got)
	}
}
