package testutil

import (
	"context"
	"sync"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
)

// FakeRunner is a scriptable runner.Runner for eval tests. Responses
// can be queued in advance (Queue) or generated lazily (Handler).
// Every Execute call is recorded on Calls for assertions.
//
// Unlike the production mock runner at internal/eval/runner/mock (which
// mimics agent behaviour for end-to-end harness tests), FakeRunner
// exists purely for unit tests that need deterministic control over
// per-call behaviour — e.g. "first call succeeds, second returns rate
// limit, third times out".
type FakeRunner struct {
	NameVal    string
	Caps       runner.Capabilities
	DoctorVal  runner.DoctorCheck
	Handler    func(ctx context.Context, req runner.Request) (*runner.Result, error)

	mu    sync.Mutex
	queue []runnerResponse
	calls []runner.Request
}

type runnerResponse struct {
	res *runner.Result
	err error
}

// NewFakeRunner returns a FakeRunner with name "fake" and empty caps.
func NewFakeRunner() *FakeRunner {
	return &FakeRunner{NameVal: "fake"}
}

// Queue appends a canned (result, err) to the response queue. Execute
// pops from this queue in order; if empty, Handler is consulted; if
// both are unset, Execute returns a minimal successful Result.
func (f *FakeRunner) Queue(res *runner.Result, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queue = append(f.queue, runnerResponse{res: res, err: err})
}

// Calls returns the requests observed so far, in call order. Safe to
// call concurrently with Execute.
func (f *FakeRunner) Calls() []runner.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]runner.Request, len(f.calls))
	copy(out, f.calls)
	return out
}

// Reset clears both the call log and the response queue.
func (f *FakeRunner) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = nil
	f.queue = nil
}

func (f *FakeRunner) Name() string                        { return f.NameVal }
func (f *FakeRunner) Capabilities() runner.Capabilities   { return f.Caps }
func (f *FakeRunner) DoctorCheck(context.Context) runner.DoctorCheck {
	return f.DoctorVal
}

func (f *FakeRunner) Execute(ctx context.Context, req runner.Request) (*runner.Result, error) {
	f.mu.Lock()
	f.calls = append(f.calls, req)
	if len(f.queue) > 0 {
		resp := f.queue[0]
		f.queue = f.queue[1:]
		f.mu.Unlock()
		return resp.res, resp.err
	}
	handler := f.Handler
	f.mu.Unlock()

	if handler != nil {
		return handler(ctx, req)
	}
	return &runner.Result{TotalTokens: 1}, nil
}
