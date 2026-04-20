package anthropicjudge

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
)

func TestExtractJSONObject(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"plain", `{"a":1}`, `{"a":1}`},
		{"fenced_json", "```json\n{\"a\":1}\n```", `{"a":1}`},
		{"preamble", "some preamble {\"a\":1} trailing", `{"a":1}`},
		{"nested", `{"outer":{"inner":1}}`, `{"outer":{"inner":1}}`},
		{"only_fence_no_brace", "```json```", ""},
		{"empty_after_trim", "", ""},
		{"whitespace_only", "   ", ""},
		{"multiline_preamble_and_trailing_prose",
			"Here is the result:\n{\n  \"expectations\": [1,2]\n}\nHope this helps.",
			"{\n  \"expectations\": [1,2]\n}",
		},
		{"nothing_braced_passes_through", "not json at all", "not json at all"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractJSONObject(tc.in); got != tc.want {
				t.Errorf("extractJSONObject(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("short string truncated: %q", got)
	}
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("long truncate = %q", got)
	}
	if got := truncate("", 5); got != "" {
		t.Errorf("empty truncate = %q", got)
	}
}

func TestNew_DefaultsToDefaultModelWhenEmpty(t *testing.T) {
	j := New("sk-unused", "")
	if j.model != DefaultModel {
		t.Errorf("model = %q, want %q", j.model, DefaultModel)
	}

	j2 := New("sk-unused", "claude-haiku-4-5")
	if j2.model != "claude-haiku-4-5" {
		t.Errorf("model = %q, want override", j2.model)
	}
}

func TestGrade_EmptyAssertionsSkipsAPICall(t *testing.T) {
	// No server set up — an API call would hang. Grade must short-circuit.
	j := New("sk-unused", "")
	got, err := j.Grade(context.Background(), "prompt", []byte("transcript"), "outputs", nil)
	if err != nil {
		t.Errorf("err = %v", err)
	}
	if got != nil {
		t.Errorf("result = %v, want nil", got)
	}
}

// ---- mock-server-backed tests ---------------------------------------------

// messagesServer is a tiny Anthropic Messages API mock. It captures the
// last request for assertions and returns whatever body the test set.
type messagesServer struct {
	t *testing.T

	mu   sync.Mutex
	body string
	code int

	lastRequest []byte
}

func newMessagesServer(t *testing.T) (*messagesServer, *httptest.Server) {
	t.Helper()
	m := &messagesServer{t: t, code: 200}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, _ := io.ReadAll(r.Body)
		m.mu.Lock()
		m.lastRequest = buf
		body := m.body
		code := m.code
		m.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return m, srv
}

func (m *messagesServer) reply(body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.body = body
	m.code = 200
}

func (m *messagesServer) replyStatus(code int, body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.body = body
	m.code = code
}

func (m *messagesServer) lastSent() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]byte(nil), m.lastRequest...)
}

// okResponse builds a minimal Messages API response whose text content
// is `text`. The SDK parses any valid Messages shape; this is the
// smallest one that exercises the "text block" branch in Grade.
func okResponse(text string) string {
	resp := map[string]any{
		"id":            "msg_test_0001",
		"type":          "message",
		"role":          "assistant",
		"model":         "claude-opus-4-5",
		"stop_reason":   "end_turn",
		"stop_sequence": nil,
		"usage":         map[string]int{"input_tokens": 1, "output_tokens": 1},
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

func newTestJudge(t *testing.T, baseURL string) *Judge {
	t.Helper()
	return NewWithOptions("claude-test-model",
		option.WithAPIKey("sk-test"),
		option.WithBaseURL(baseURL),
	)
}

func TestGrade_HappyPath_ParsesJSONAndPreservesOrder(t *testing.T) {
	m, srv := newMessagesServer(t)
	j := newTestJudge(t, srv.URL)

	assertions := []scenarios.Assertion{
		{Text: "The agent explained the bug"},
		{Text: "The agent proposed a fix"},
	}
	// Model paraphrases the text; Grade must overwrite with the
	// original assertion text so downstream aggregation still keys
	// cleanly.
	judgeReply := `{"expectations":[
		{"text":"paraphrased version of bug claim","passed":true,"evidence":"quoted: 'root cause is a nil deref'"},
		{"text":"paraphrased fix claim","passed":false,"evidence":"no fix shown"}
	]}`
	m.reply(okResponse(judgeReply))

	got, err := j.Grade(context.Background(), "explain the bug", []byte("transcript body"), "outputs body", assertions)
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("results = %d", len(got))
	}
	if got[0].Text != assertions[0].Text || got[1].Text != assertions[1].Text {
		t.Errorf("text not restored to original: got [%q, %q]", got[0].Text, got[1].Text)
	}
	if !got[0].Passed || got[1].Passed {
		t.Errorf("passed flags wrong: %+v", got)
	}
	if got[0].Evidence == "" || got[1].Evidence == "" {
		t.Errorf("evidence empty: %+v", got)
	}

	// Verify request shape: contains assertion numbering and the eval prompt.
	req := string(m.lastSent())
	if !strings.Contains(req, `"model":"claude-test-model"`) {
		t.Errorf("request missing model field: %s", truncate(req, 200))
	}
	if !strings.Contains(req, "1. The agent explained the bug") {
		t.Errorf("request missing first assertion: %s", truncate(req, 300))
	}
	if !strings.Contains(req, "2. The agent proposed a fix") {
		t.Errorf("request missing second assertion")
	}
	if !strings.Contains(req, "explain the bug") {
		t.Errorf("request missing eval prompt")
	}
}

func TestGrade_CountMismatchErrors(t *testing.T) {
	m, srv := newMessagesServer(t)
	j := newTestJudge(t, srv.URL)

	m.reply(okResponse(`{"expectations":[{"text":"a","passed":true,"evidence":"x"}]}`))

	assertions := []scenarios.Assertion{
		{Text: "first"},
		{Text: "second"},
	}
	_, err := j.Grade(context.Background(), "p", []byte("t"), "o", assertions)
	if err == nil {
		t.Fatal("expected error when judge returns wrong expectation count")
	}
	if !strings.Contains(err.Error(), "returned 1 expectations, want 2") {
		t.Errorf("err = %v", err)
	}
}

func TestGrade_MalformedJSONSurfaceTruncatedReply(t *testing.T) {
	m, srv := newMessagesServer(t)
	j := newTestJudge(t, srv.URL)

	// Content block has text that is not valid JSON at all.
	m.reply(okResponse("this is not JSON and has no braces"))

	_, err := j.Grade(context.Background(), "p", []byte("t"), "o",
		[]scenarios.Assertion{{Text: "a"}})
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse judge reply") {
		t.Errorf("err doesn't mention parsing: %v", err)
	}
}

func TestGrade_FencedReplyIsUnwrapped(t *testing.T) {
	m, srv := newMessagesServer(t)
	j := newTestJudge(t, srv.URL)

	wrapped := "```json\n" +
		`{"expectations":[{"text":"a","passed":true,"evidence":"e"}]}` +
		"\n```"
	m.reply(okResponse(wrapped))

	got, err := j.Grade(context.Background(), "p", []byte("t"), "o",
		[]scenarios.Assertion{{Text: "a"}})
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	if len(got) != 1 || !got[0].Passed {
		t.Errorf("result = %+v", got)
	}
}

func TestGrade_TranscriptAndOutputsAreTruncated(t *testing.T) {
	m, srv := newMessagesServer(t)
	j := newTestJudge(t, srv.URL)

	m.reply(okResponse(`{"expectations":[{"text":"a","passed":true,"evidence":"e"}]}`))

	big := strings.Repeat("A", 60*1024)     // 60 KiB, well above the 40 KiB cap
	bigOut := strings.Repeat("B", 60*1024)

	_, err := j.Grade(context.Background(), "p", []byte(big), bigOut,
		[]scenarios.Assertion{{Text: "a"}})
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	req := m.lastSent()
	// Transcript payload capped by the 40 KiB limit + truncation marker.
	if !strings.Contains(string(req), "transcript truncated") {
		t.Errorf("transcript not truncated in request")
	}
	if !strings.Contains(string(req), "outputs truncated") {
		t.Errorf("outputs not truncated in request")
	}
}

func TestGrade_APIErrorPropagates(t *testing.T) {
	m, srv := newMessagesServer(t)
	j := newTestJudge(t, srv.URL)

	m.replyStatus(500, `{"type":"error","error":{"type":"api_error","message":"boom"}}`)

	_, err := j.Grade(context.Background(), "p", []byte("t"), "o",
		[]scenarios.Assertion{{Text: "a"}})
	if err == nil {
		t.Fatal("expected API error to propagate")
	}
	if !strings.Contains(err.Error(), "grader.messages.new") {
		t.Errorf("err missing context: %v", err)
	}
}

func TestGrade_ContextCancellationPropagates(t *testing.T) {
	_, srv := newMessagesServer(t)
	j := newTestJudge(t, srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already done before Grade is called

	_, err := j.Grade(ctx, "p", []byte("t"), "o",
		[]scenarios.Assertion{{Text: "a"}})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
