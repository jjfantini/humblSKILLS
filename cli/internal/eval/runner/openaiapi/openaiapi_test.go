package openaiapi

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

	"github.com/openai/openai-go/option"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

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
	r := New(&fakeStore{keys: map[string]string{"openai": "sk"}})
	if r.Name() != "openai-api" {
		t.Errorf("Name = %q", r.Name())
	}
	if r.Capabilities().DefaultModel != DefaultModel {
		t.Errorf("DefaultModel = %q", r.Capabilities().DefaultModel)
	}
}

func TestDoctorCheck_NilStore(t *testing.T) {
	r := New(nil)
	if r.DoctorCheck(context.Background()).Available {
		t.Error("expected Available=false for nil store")
	}
}

func TestDoctorCheck_AbsentKey(t *testing.T) {
	r := New(&fakeStore{absent: true})
	got := r.DoctorCheck(context.Background())
	if got.Available {
		t.Error("expected Available=false")
	}
	if !strings.Contains(got.Reason, "OPENAI_API_KEY") {
		t.Errorf("Reason = %q", got.Reason)
	}
}

func TestDoctorCheck_KeyPresent(t *testing.T) {
	r := New(&fakeStore{keys: map[string]string{"openai": "sk"}})
	if !r.DoctorCheck(context.Background()).Available {
		t.Error("expected Available=true")
	}
}

func TestExecute_NoKeyErrors(t *testing.T) {
	r := New(&fakeStore{absent: true})
	_, err := r.Execute(context.Background(), runner.Request{
		Prompt: "p", OutputDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---- httptest-based tests ------------------------------------------------

type chatServer struct {
	mu        sync.Mutex
	responses []string
	nextIdx   int
	calls     int
}

func (c *chatServer) handle(w http.ResponseWriter, r *http.Request) {
	_, _ = io.ReadAll(r.Body)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	idx := c.nextIdx
	if idx >= len(c.responses) {
		idx = len(c.responses) - 1
	}
	c.nextIdx++
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(c.responses[idx]))
}

func chatResponse(content string, toolCalls []map[string]any) string {
	msg := map[string]any{"role": "assistant"}
	if content != "" {
		msg["content"] = content
	}
	if len(toolCalls) > 0 {
		msg["tool_calls"] = toolCalls
	}
	resp := map[string]any{
		"id":      "cmpl_1",
		"object":  "chat.completion",
		"created": 1,
		"model":   "gpt-test",
		"choices": []map[string]any{
			{"index": 0, "message": msg, "finish_reason": "stop"},
		},
		"usage": map[string]int{"prompt_tokens": 7, "completion_tokens": 3, "total_tokens": 10},
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

func TestExecute_EndsOnStop(t *testing.T) {
	testutil.NewSandbox(t)

	cs := &chatServer{responses: []string{chatResponse("done!", nil)}}
	srv := httptest.NewServer(http.HandlerFunc(cs.handle))
	defer srv.Close()

	r := NewWithOptions(&fakeStore{keys: map[string]string{"openai": "sk"}},
		option.WithBaseURL(srv.URL))

	res, err := r.Execute(context.Background(), runner.Request{
		Prompt: "p", OutputDir: filepath.Join(t.TempDir(), "out"),
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if cs.calls != 1 {
		t.Errorf("calls = %d", cs.calls)
	}
	if !strings.Contains(string(res.Transcript), "done!") {
		t.Errorf("transcript missing content:\n%s", res.Transcript)
	}
	if res.PromptTokens != 7 || res.CompletionTokens != 3 {
		t.Errorf("tokens: %+v", res)
	}
}

func TestExecute_ToolLoop(t *testing.T) {
	testutil.NewSandbox(t)

	workDir := t.TempDir()
	input := filepath.Join(workDir, "hello.txt")
	_ = os.WriteFile(input, []byte("content"), 0o644)

	toolCall := []map[string]any{
		{
			"id":   "call_1",
			"type": "function",
			"function": map[string]any{
				"name":      "Read",
				"arguments": `{"path":"inputs/hello.txt"}`,
			},
		},
	}
	cs := &chatServer{
		responses: []string{
			chatResponse("", toolCall),
			chatResponse("summary", nil),
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(cs.handle))
	defer srv.Close()

	r := NewWithOptions(&fakeStore{keys: map[string]string{"openai": "sk"}},
		option.WithBaseURL(srv.URL))

	res, err := r.Execute(context.Background(), runner.Request{
		Prompt: "read the file", OutputDir: filepath.Join(t.TempDir(), "out"),
		InputFiles: []string{input},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if cs.calls != 2 {
		t.Errorf("calls = %d, want 2", cs.calls)
	}
	if res.ToolCalls["Read"] != 1 {
		t.Errorf("Read tool count = %d", res.ToolCalls["Read"])
	}
}

func TestExecute_APIErrorOnResultErr(t *testing.T) {
	testutil.NewSandbox(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = io.WriteString(w, `{"error":{"message":"boom"}}`)
	}))
	defer srv.Close()

	r := NewWithOptions(&fakeStore{keys: map[string]string{"openai": "sk"}},
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

func TestBuildSystem_WithAndWithoutSkill(t *testing.T) {
	skill := t.TempDir()
	_ = os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("# s"), 0o644)
	if got := buildSystem(runner.Request{SkillDir: skill}); !strings.Contains(got, "# s") {
		t.Errorf("skill body missing: %q", got)
	}
	if got := buildSystem(runner.Request{}); got != "" {
		t.Errorf("empty: got %q", got)
	}
	if got := buildSystem(runner.Request{SystemPrompt: "be nice"}); !strings.Contains(got, "be nice") {
		t.Errorf("system prompt missing: %q", got)
	}
}

func TestToolDefs(t *testing.T) {
	if len(toolDefs()) != 5 {
		t.Errorf("expected 5 tool defs")
	}
}

func TestStageInputs_CopiesFiles(t *testing.T) {
	work := t.TempDir()
	src := filepath.Join(work, "in.txt")
	_ = os.WriteFile(src, []byte("x"), 0o644)
	scratch := t.TempDir()
	if err := stageInputs(runner.Request{InputFiles: []string{src}}, scratch); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(scratch, "inputs", "in.txt")); err != nil {
		t.Errorf("not staged: %v", err)
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
}

func TestOneLine(t *testing.T) {
	if got := oneLine(strings.Repeat("a", 100), 10); !strings.HasSuffix(got, "...") {
		t.Errorf("got %q", got)
	}
}
