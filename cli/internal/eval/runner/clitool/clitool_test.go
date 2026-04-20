package clitool_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/clitool"
)

// installStubBinary writes a POSIX shell script to a fresh temp dir,
// prepends that dir to $PATH, and returns the binary name. Tests use
// this to exercise Execute without needing a real agent CLI.
func installStubBinary(t *testing.T, name, body string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("stub binary uses POSIX shell")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	old := os.Getenv("PATH")
	_ = os.Setenv("PATH", dir+string(os.PathListSeparator)+old)
	t.Cleanup(func() { _ = os.Setenv("PATH", old) })
}

func TestParseJSONEvent_SkipsNonJSON(t *testing.T) {
	got, err := clitool.ParseJSONEvent([]byte("not json"))
	if err != nil {
		t.Errorf("should return nil err on non-json, got %v", err)
	}
	if got != nil {
		t.Errorf("should return nil map, got %v", got)
	}
}

func TestParseJSONEvent_Empty(t *testing.T) {
	got, err := clitool.ParseJSONEvent([]byte("  "))
	if err != nil || got != nil {
		t.Errorf("empty line returned got=%v err=%v", got, err)
	}
}

func TestParseJSONEvent_ParsesObject(t *testing.T) {
	got, err := clitool.ParseJSONEvent([]byte(`{"k":"v"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got["k"] != "v" {
		t.Errorf("got %v", got)
	}
}

func TestParseJSONEvent_MalformedObject(t *testing.T) {
	_, err := clitool.ParseJSONEvent([]byte("{bad json"))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestRunner_NameAndCapabilities(t *testing.T) {
	r := clitool.New(clitool.Driver{
		Name:         "test-cli",
		Binary:       "nonexistent-binary",
		DefaultModel: "model-x",
	})
	if r.Name() != "test-cli" {
		t.Errorf("Name = %q", r.Name())
	}
	caps := r.Capabilities()
	if caps.DefaultModel != "model-x" {
		t.Errorf("DefaultModel = %q", caps.DefaultModel)
	}
	if len(caps.SupportsTools) == 0 {
		t.Error("SupportsTools empty")
	}
}

func TestDoctorCheck_BinaryMissing(t *testing.T) {
	r := clitool.New(clitool.Driver{
		Name:        "test-cli",
		Binary:      "definitely-not-a-real-binary-xyz123",
		RequiresKey: "anthropic",
	})
	got := r.DoctorCheck(context.Background())
	if got.Available {
		t.Error("expected Available=false")
	}
	if !strings.Contains(got.Reason, "PATH") {
		t.Errorf("Reason = %q", got.Reason)
	}
	if got.RequiresKey != "anthropic" {
		t.Errorf("RequiresKey = %q", got.RequiresKey)
	}
}

func TestDoctorCheck_BinaryPresent(t *testing.T) {
	installStubBinary(t, "fake-cli", "echo 'fake-cli 1.2.3'\n")
	r := clitool.New(clitool.Driver{
		Name:        "test-cli",
		Binary:      "fake-cli",
		VersionArgs: []string{"--version"},
	})
	got := r.DoctorCheck(context.Background())
	if !got.Available {
		t.Error("expected Available=true")
	}
	if !strings.Contains(got.Version, "fake-cli") {
		t.Errorf("Version = %q", got.Version)
	}
}

func TestExecute_StreamsJSONAndCapturesTokens(t *testing.T) {
	// Stub emits two JSON lines — a tool_use then a usage event. The
	// Driver's ParseEvent extracts token counts from the latter.
	installStubBinary(t, "stub-agent", `
cat <<'JSON'
{"type":"tool_use","name":"Read"}
{"type":"result","usage":{"input_tokens":10,"output_tokens":20}}
JSON
`)

	r := clitool.New(clitool.Driver{
		Name:   "stub-agent",
		Binary: "stub-agent",
		Args:   func(req runner.Request, scratchDir, promptPath string) []string { return []string{} },
		ParseEvent: func(line []byte) clitool.Event {
			m, _ := clitool.ParseJSONEvent(line)
			if m == nil {
				return clitool.Event{}
			}
			var ev clitool.Event
			if t, _ := m["type"].(string); t == "tool_use" {
				ev.ToolName, _ = m["name"].(string)
			}
			if usage, ok := m["usage"].(map[string]any); ok {
				if pt, ok := usage["input_tokens"].(float64); ok {
					ev.PromptTokensDelta = int(pt)
				}
				if ct, ok := usage["output_tokens"].(float64); ok {
					ev.CompletionTokensDelta = int(ct)
				}
			}
			return ev
		},
	})

	outDir := filepath.Join(t.TempDir(), "out")
	res, err := r.Execute(context.Background(), runner.Request{
		Prompt:    "do a thing",
		OutputDir: outDir,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.PromptTokens != 10 || res.CompletionTokens != 20 {
		t.Errorf("tokens = (%d, %d)", res.PromptTokens, res.CompletionTokens)
	}
	if res.TotalTokens != 30 {
		t.Errorf("total tokens = %d, want 30", res.TotalTokens)
	}
	if res.ToolCalls["Read"] != 1 {
		t.Errorf("Read count = %d", res.ToolCalls["Read"])
	}
	if res.DurationMs < 0 {
		t.Errorf("DurationMs = %d", res.DurationMs)
	}
	if !strings.Contains(string(res.Transcript), "tool_use") {
		t.Errorf("transcript missing events:\n%s", res.Transcript)
	}
}

func TestExecute_BinaryFailureSurfacesOnResultErr(t *testing.T) {
	installStubBinary(t, "stub-fail", "echo 'boom' >&2; exit 1\n")
	r := clitool.New(clitool.Driver{
		Name:   "stub-fail",
		Binary: "stub-fail",
		Args:   func(req runner.Request, _, _ string) []string { return []string{} },
	})
	res, err := r.Execute(context.Background(), runner.Request{
		Prompt:    "p",
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Execute should not return infra error on binary failure: %v", err)
	}
	if res.Err == nil {
		t.Error("expected res.Err to be set when binary exits non-zero")
	}
	if res.Err != nil && !strings.Contains(res.Err.Error(), "boom") {
		t.Errorf("err missing stderr: %v", res.Err)
	}
}

func TestExecute_StdinFedToBinary(t *testing.T) {
	// Stub echoes stdin back. Driver supplies a Stdin function; the
	// Result.Transcript should contain that text.
	installStubBinary(t, "stub-stdin", "cat\n")
	r := clitool.New(clitool.Driver{
		Name:   "stub-stdin",
		Binary: "stub-stdin",
		Args:   func(req runner.Request, _, _ string) []string { return []string{} },
		Stdin:  func(req runner.Request) []byte { return []byte("piped-content\n") },
	})
	res, err := r.Execute(context.Background(), runner.Request{
		Prompt:    "p",
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res.Transcript), "piped-content") {
		t.Errorf("transcript missing stdin echo:\n%s", res.Transcript)
	}
}

func TestExecute_InputFilesStagedIntoScratch(t *testing.T) {
	// Stub writes every file it sees in inputs/ into outputs/.
	installStubBinary(t, "stub-inputs", `
mkdir -p outputs
for f in inputs/*; do
  cp "$f" "outputs/seen-$(basename "$f")"
done
`)

	workDir := t.TempDir()
	in := filepath.Join(workDir, "data.txt")
	if err := os.WriteFile(in, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := clitool.New(clitool.Driver{
		Name:   "stub-inputs",
		Binary: "stub-inputs",
		Args:   func(req runner.Request, _, _ string) []string { return []string{} },
	})
	outDir := filepath.Join(t.TempDir(), "outputs")
	res, err := r.Execute(context.Background(), runner.Request{
		Prompt:     "p",
		InputFiles: []string{in},
		OutputDir:  outDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	// outputs/seen-data.txt was produced — clitool flattens outputs/
	// into OutputDir.
	found := false
	for _, f := range res.OutputFiles {
		if strings.HasSuffix(f, "seen-data.txt") {
			found = true
		}
	}
	if !found {
		t.Errorf("input was not staged; OutputFiles=%v", res.OutputFiles)
	}
}

func TestExecute_TimeoutHonored(t *testing.T) {
	// exec -c's `sleep` runs as a child of the shell; on some platforms
	// SIGKILL to the shell doesn't cascade to sleep, so we use an
	// `exec sleep` pattern so the process-group receives the signal.
	installStubBinary(t, "stub-sleep", "exec sleep 10\n")
	r := clitool.New(clitool.Driver{
		Name:   "stub-sleep",
		Binary: "stub-sleep",
		Args:   func(req runner.Request, _, _ string) []string { return []string{} },
	})
	start := time.Now()
	_, err := r.Execute(context.Background(), runner.Request{
		Prompt:    "p",
		OutputDir: t.TempDir(),
		Timeout:   300 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Execute err: %v", err)
	}
	// Must be well under the 10s sleep to prove the timeout fired.
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("Timeout not honored, took %v", elapsed)
	}
}

func TestExecute_SkillDirStagedAndBrainSyncedBack(t *testing.T) {
	// Seed a skill with references/log.md that the "agent" will append
	// to. Verify the persistent skillDir has the appended content.
	skillDir := t.TempDir()
	refs := filepath.Join(skillDir, "references")
	_ = os.MkdirAll(refs, 0o755)
	logPath := filepath.Join(refs, "log.md")
	_ = os.WriteFile(logPath, []byte("initial\n"), 0o644)

	installStubBinary(t, "stub-brain", `
echo "agent-added" >> skill/references/log.md
`)

	r := clitool.New(clitool.Driver{
		Name:   "stub-brain",
		Binary: "stub-brain",
		Args:   func(req runner.Request, _, _ string) []string { return []string{} },
	})
	if _, err := r.Execute(context.Background(), runner.Request{
		Prompt:    "p",
		SkillDir:  skillDir,
		OutputDir: t.TempDir(),
	}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(logPath)
	if !strings.Contains(string(got), "agent-added") {
		t.Errorf("brain sync failed; log.md = %q", got)
	}
}
