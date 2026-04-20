package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

// execResult captures stdout/stderr and the error Cobra surfaced.
type execResult struct {
	Out string
	Err string
	// RunErr is what RunE returned; non-nil doesn't mean process exit
	// since main wraps with os.Exit — tests assert on RunErr directly.
	RunErr error
}

// runCLI builds the root command tree and executes it with args.
// Writers are captured so tests can assert on output.
//
// Critical: root installs a PersistentPreRunE that reads env / XDG
// paths. Callers should place runCLI inside a testutil.NewSandbox so
// those resolve to sandboxed locations.
func runCLI(t *testing.T, args ...string) execResult {
	t.Helper()
	root := newRootCmd()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	root.SetArgs(args)

	res := execResult{RunErr: root.Execute()}

	// PersistentPreRunE constructs app.UI pointing at os.Stdout/os.Stderr
	// regardless of SetOut. To capture command output we would need to
	// inject an app-level writer. For now we capture what Cobra's own
	// writers receive (error/usage paths); per-command assertions use
	// the app-level --json flag where possible since JSON output goes
	// through app.UI.JSON which respects the injected writer via
	// runCLIWithStdoutCapture below.
	res.Out = outBuf.String()
	res.Err = errBuf.String()
	return res
}

// runCLIWithStdoutCapture redirects os.Stdout/os.Stderr for the
// duration of Execute so commands that print via fmt.Fprintln(app.UI.Out())
// can be asserted against. Slightly heavier than runCLI — prefer
// this when you need full stdout.
func runCLIWithStdoutCapture(t *testing.T, args ...string) execResult {
	t.Helper()

	// Pipe for stdout / stderr.
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	origStdout := os.Stdout
	origStderr := os.Stderr
	os.Stdout = stdoutW
	os.Stderr = stderrW
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	outCh := make(chan string, 1)
	errCh := make(chan string, 1)
	go func() {
		var b bytes.Buffer
		_, _ = b.ReadFrom(stdoutR)
		outCh <- b.String()
	}()
	go func() {
		var b bytes.Buffer
		_, _ = b.ReadFrom(stderrR)
		errCh <- b.String()
	}()

	root := newRootCmd()
	root.SetArgs(args)
	res := execResult{RunErr: root.Execute()}

	// Close the writers so the goroutines see EOF.
	_ = stdoutW.Close()
	_ = stderrW.Close()
	res.Out = <-outCh
	res.Err = <-errCh
	return res
}

// seedTestRegistry creates a registry.json file under s.CacheDir and
// returns the file:// URL to feed via --registry. This lets tests
// avoid spinning up an httptest server when all they need is a fixed
// registry document.
func seedTestRegistry(t *testing.T, s *testutil.Sandbox, fixtures []testutil.SkillFixture) string {
	t.Helper()
	reg := testutil.BuildRegistry(t, s.CacheDir, "example", "repo", "0123456789abcdef", fixtures)

	path := filepath.Join(s.Root, "registry.json")
	data := mustJSON(t, reg)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write registry.json: %v", err)
	}
	return "file://" + path
}

func mustJSON(t *testing.T, reg *registry.Registry) []byte {
	t.Helper()
	// Use the registry package's own serialization shape via json.
	// Importing encoding/json here keeps the helper self-contained.
	data, err := registryToJSON(reg)
	if err != nil {
		t.Fatalf("marshal registry: %v", err)
	}
	return data
}

// assertContains fails when haystack doesn't contain needle.
func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q, got:\n%s", needle, haystack)
	}
}

// assertNotContains fails when haystack contains needle.
func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", needle, haystack)
	}
}
