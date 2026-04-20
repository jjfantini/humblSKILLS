package anthropicapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

// fakeStore is a minimal in-memory secrets.Store.
type fakeStore struct {
	keys   map[string]string
	absent bool
}

func (f *fakeStore) Get(p string) (string, secrets.Source, error) {
	if f.absent {
		return "", secrets.SourceAbsent, nil
	}
	v, ok := f.keys[p]
	if !ok || v == "" {
		return "", secrets.SourceAbsent, nil
	}
	return v, secrets.SourceEnv, nil
}
func (f *fakeStore) Set(p, v string) (secrets.Source, error) {
	if f.keys == nil {
		f.keys = map[string]string{}
	}
	f.keys[p] = v
	return secrets.SourceEnv, nil
}
func (f *fakeStore) Delete(p string) error { delete(f.keys, p); return nil }

func TestName_AndCapabilities(t *testing.T) {
	r := New(&fakeStore{keys: map[string]string{"anthropic": "sk"}})
	if r.Name() != "anthropic-api" {
		t.Errorf("Name = %q", r.Name())
	}
	caps := r.Capabilities()
	if caps.DefaultModel != DefaultModel {
		t.Errorf("DefaultModel = %q", caps.DefaultModel)
	}
	if caps.Pricing == nil {
		t.Error("Pricing nil")
	}
}

func TestDoctorCheck_NilStore(t *testing.T) {
	r := New(nil)
	got := r.DoctorCheck(context.Background())
	if got.Available {
		t.Error("expected Available=false with nil store")
	}
}

func TestDoctorCheck_AbsentKey(t *testing.T) {
	r := New(&fakeStore{absent: true})
	got := r.DoctorCheck(context.Background())
	if got.Available {
		t.Error("expected Available=false when key absent")
	}
	if !strings.Contains(got.Reason, "ANTHROPIC_API_KEY") {
		t.Errorf("Reason = %q", got.Reason)
	}
}

func TestDoctorCheck_KeyPresent(t *testing.T) {
	r := New(&fakeStore{keys: map[string]string{"anthropic": "sk-test"}})
	got := r.DoctorCheck(context.Background())
	if !got.Available {
		t.Error("expected Available=true with key set")
	}
}

func TestExecute_WithoutAPIKey(t *testing.T) {
	r := New(&fakeStore{absent: true})
	_, err := r.Execute(context.Background(), runner.Request{
		Prompt: "p", OutputDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error when key absent")
	}
}

// messagesServer is a scriptable /messages endpoint.
type messagesServer struct {
	mu       sync.Mutex
	responses []string
	nextIdx  int
	calls    int
}

func (m *messagesServer) handle(w http.ResponseWriter, r *http.Request) {
	_, _ = io.ReadAll(r.Body)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	idx := m.nextIdx
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	m.nextIdx++
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(m.responses[idx]))
}

func endTurnResponse(text string) string {
	resp := map[string]any{
		"id": "msg_1", "type": "message", "role": "assistant",
		"model": "claude-test", "stop_reason": "end_turn",
		"usage":   map[string]int{"input_tokens": 10, "output_tokens": 5},
		"content": []map[string]any{{"type": "text", "text": text}},
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

func toolUseResponse(toolName, toolID string, args map[string]any) string {
	argBytes, _ := json.Marshal(args)
	resp := map[string]any{
		"id": "msg_tu", "type": "message", "role": "assistant",
		"model": "claude-test", "stop_reason": "tool_use",
		"usage": map[string]int{"input_tokens": 12, "output_tokens": 3},
		"content": []map[string]any{
			{"type": "tool_use", "id": toolID, "name": toolName, "input": json.RawMessage(argBytes)},
		},
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

func TestExecute_HappyPath_EndsOnEndTurn(t *testing.T) {
	testutil.NewSandbox(t)

	ms := &messagesServer{
		responses: []string{endTurnResponse("hello from the model")},
	}
	srv := httptest.NewServer(http.HandlerFunc(ms.handle))
	defer srv.Close()

	r := NewWithOptions(&fakeStore{keys: map[string]string{"anthropic": "sk-test"}},
		option.WithBaseURL(srv.URL))

	outDir := filepath.Join(t.TempDir(), "out")
	res, err := r.Execute(context.Background(), runner.Request{
		Prompt: "hi", OutputDir: outDir,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(string(res.Transcript), "hello from the model") {
		t.Errorf("transcript missing text:\n%s", res.Transcript)
	}
	if res.PromptTokens != 10 || res.CompletionTokens != 5 {
		t.Errorf("tokens: %+v", res)
	}
	if ms.calls != 1 {
		t.Errorf("calls = %d, want 1", ms.calls)
	}
}

func TestExecute_ToolLoop(t *testing.T) {
	testutil.NewSandbox(t)

	// Response 1: tool_use Read foo.txt. Response 2: end_turn with summary.
	ms := &messagesServer{
		responses: []string{
			toolUseResponse("Read", "t_1", map[string]any{"path": "hello.txt"}),
			endTurnResponse("read complete"),
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(ms.handle))
	defer srv.Close()

	// Pre-create the file the model will "read" inside the sandbox that
	// Execute creates. Since we can't inject the scratch path, stage the
	// file via InputFiles so it lands at scratch/inputs/hello.txt.
	// Instead, write our tool_use with the correct path.
	ms.responses[0] = toolUseResponse("Read", "t_1", map[string]any{"path": "inputs/hello.txt"})

	workDir := t.TempDir()
	input := filepath.Join(workDir, "hello.txt")
	if err := os.WriteFile(input, []byte("hello content"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewWithOptions(&fakeStore{keys: map[string]string{"anthropic": "sk-test"}},
		option.WithBaseURL(srv.URL))

	outDir := filepath.Join(t.TempDir(), "out")
	res, err := r.Execute(context.Background(), runner.Request{
		Prompt:     "read the file",
		OutputDir:  outDir,
		InputFiles: []string{input},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if ms.calls != 2 {
		t.Errorf("calls = %d, want 2 (tool + end_turn)", ms.calls)
	}
	if res.ToolCalls["Read"] != 1 {
		t.Errorf("Read tool count = %d", res.ToolCalls["Read"])
	}
	// Sum of tokens across both rounds.
	if res.PromptTokens != 22 || res.CompletionTokens != 8 {
		t.Errorf("tokens = (%d, %d)", res.PromptTokens, res.CompletionTokens)
	}
}

func TestExecute_ApiErrorSurfacedOnResultErr(t *testing.T) {
	testutil.NewSandbox(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = io.WriteString(w, `{"type":"error","error":{"type":"api_error","message":"boom"}}`)
	}))
	defer srv.Close()

	r := NewWithOptions(&fakeStore{keys: map[string]string{"anthropic": "sk-test"}},
		option.WithBaseURL(srv.URL))

	res, err := r.Execute(context.Background(), runner.Request{
		Prompt: "p", OutputDir: filepath.Join(t.TempDir(), "out"),
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Err == nil {
		t.Error("expected res.Err on API failure")
	}
}

func TestBuildSystem_PrependsSkillBody(t *testing.T) {
	skill := t.TempDir()
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("# spec"), 0o644); err != nil {
		t.Fatal(err)
	}
	blocks := buildSystem(runner.Request{SkillDir: skill})
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks", len(blocks))
	}
	if !strings.Contains(blocks[0].Text, "# spec") {
		t.Errorf("skill body not included:\n%s", blocks[0].Text)
	}
}

func TestBuildSystem_ReturnsNilWhenNothing(t *testing.T) {
	blocks := buildSystem(runner.Request{})
	if blocks != nil {
		t.Errorf("got %+v, want nil", blocks)
	}
}

func TestBuildSystem_SystemPromptTakesPrecedenceOverMissingSkill(t *testing.T) {
	blocks := buildSystem(runner.Request{SystemPrompt: "be terse"})
	if len(blocks) != 1 || blocks[0].Text != "be terse" {
		t.Errorf("got %+v", blocks)
	}
}

func TestToolDefs_ReturnsFive(t *testing.T) {
	defs := toolDefs()
	if len(defs) != 5 {
		t.Errorf("got %d, want 5", len(defs))
	}
}

func TestOneLine_Truncates(t *testing.T) {
	long := strings.Repeat("a", 200)
	if got := oneLine(long, 20); len(got) > 30 || !strings.HasSuffix(got, "...") {
		t.Errorf("got %q", got)
	}
}

func TestStageInputs_CopiesFiles(t *testing.T) {
	work := t.TempDir()
	src := filepath.Join(work, "in.txt")
	if err := os.WriteFile(src, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	err := stageInputs(runner.Request{InputFiles: []string{src}}, scratch)
	if err != nil {
		t.Fatalf("stageInputs: %v", err)
	}
	if _, err := os.Stat(filepath.Join(scratch, "inputs", "in.txt")); err != nil {
		t.Errorf("file not staged: %v", err)
	}
}

func TestCollectIntoOutput_SkipsInputs(t *testing.T) {
	scratch := t.TempDir()
	_ = os.MkdirAll(filepath.Join(scratch, "inputs"), 0o755)
	_ = os.WriteFile(filepath.Join(scratch, "inputs", "x.txt"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(scratch, "out.md"), []byte("body"), 0o644)

	out := t.TempDir()
	files, err := collectIntoOutput(scratch, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "out.md" {
		t.Errorf("got %v", files)
	}
	if _, err := os.Stat(filepath.Join(out, "inputs", "x.txt")); err == nil {
		t.Error("inputs/ should not be copied to output")
	}
}
