package toolbox_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner/toolbox"
)

func newBox(t *testing.T) *toolbox.Sandbox {
	t.Helper()
	sb, err := toolbox.NewSandbox(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sb.ExecTimeout = 5 * time.Second
	return sb
}

func TestDefaultTools_ReturnsFive(t *testing.T) {
	tools := toolbox.DefaultTools()
	if len(tools) != 5 {
		t.Fatalf("tools = %d, want 5", len(tools))
	}
	want := map[string]bool{"Read": true, "Write": true, "Bash": true, "Glob": true, "Grep": true}
	for _, tl := range tools {
		if !want[tl.Name] {
			t.Errorf("unexpected tool %q", tl.Name)
		}
		if tl.Description == "" {
			t.Errorf("tool %q missing description", tl.Name)
		}
		if tl.Schema == nil {
			t.Errorf("tool %q missing schema", tl.Name)
		}
	}
}

func TestCall_Read(t *testing.T) {
	sb := newBox(t)
	if err := os.WriteFile(filepath.Join(sb.Root, "x.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := sb.Call(context.Background(), "Read", map[string]any{"path": "x.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestCall_Write_ThenRead(t *testing.T) {
	sb := newBox(t)
	_, err := sb.Call(context.Background(), "Write", map[string]any{
		"path": "nested/a.txt", "content": "hi",
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, _ := sb.Call(context.Background(), "Read", map[string]any{"path": "nested/a.txt"})
	if got != "hi" {
		t.Errorf("got %q", got)
	}
}

func TestCall_Bash(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash tool uses sh -c; skipping on windows")
	}
	sb := newBox(t)
	got, err := sb.Call(context.Background(), "Bash", map[string]any{"command": "echo ok"})
	if err != nil {
		t.Fatalf("Bash: %v", err)
	}
	if !strings.Contains(got, "ok") {
		t.Errorf("got %q", got)
	}
}

func TestBash_NonZeroExitSurfacesError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	sb := newBox(t)
	out, err := sb.Bash(context.Background(), "exit 3")
	if err == nil {
		t.Fatal("expected non-zero exit error")
	}
	// Output can still contain anything the command emitted before dying.
	_ = out
}

func TestBash_EmptyCommand(t *testing.T) {
	sb := newBox(t)
	if _, err := sb.Bash(context.Background(), "   "); err == nil {
		t.Error("expected error for empty command")
	}
}

func TestBash_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	sb := newBox(t)
	sb.ExecTimeout = 100 * time.Millisecond
	start := time.Now()
	_, _ = sb.Bash(context.Background(), "sleep 5")
	if time.Since(start) > 2*time.Second {
		t.Error("Bash did not honor timeout")
	}
}

func TestCall_Glob_MatchesBasenameAndRelative(t *testing.T) {
	sb := newBox(t)
	_ = os.MkdirAll(filepath.Join(sb.Root, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(sb.Root, "a.md"), []byte("a"), 0o644)
	_ = os.WriteFile(filepath.Join(sb.Root, "sub", "b.md"), []byte("b"), 0o644)

	got, err := sb.Call(context.Background(), "Glob", map[string]any{"pattern": "*.md"})
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if !strings.Contains(got, "a.md") || !strings.Contains(got, "b.md") {
		t.Errorf("glob missing matches: %q", got)
	}
}

func TestGlob_EmptyPatternErrors(t *testing.T) {
	sb := newBox(t)
	if _, err := sb.Glob(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestGlob_NoMatches(t *testing.T) {
	sb := newBox(t)
	got, err := sb.Glob("nothing-matches-this")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "no matches") {
		t.Errorf("got %q", got)
	}
}

func TestCall_Grep_ReturnsMatchLocation(t *testing.T) {
	sb := newBox(t)
	_ = os.WriteFile(filepath.Join(sb.Root, "haystack.md"), []byte("line one\nneedle here\nthird line\n"), 0o644)

	got, err := sb.Call(context.Background(), "Grep", map[string]any{"pattern": "needle"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "haystack.md:2") {
		t.Errorf("got %q", got)
	}
}

func TestGrep_InvalidRegex(t *testing.T) {
	sb := newBox(t)
	if _, err := sb.Grep("[invalid", ""); err == nil {
		t.Fatal("expected regex error")
	}
}

func TestGrep_PathScoped(t *testing.T) {
	sb := newBox(t)
	_ = os.MkdirAll(filepath.Join(sb.Root, "a"), 0o755)
	_ = os.MkdirAll(filepath.Join(sb.Root, "b"), 0o755)
	_ = os.WriteFile(filepath.Join(sb.Root, "a", "f.txt"), []byte("match_a"), 0o644)
	_ = os.WriteFile(filepath.Join(sb.Root, "b", "f.txt"), []byte("match_b"), 0o644)

	got, err := sb.Grep("match", "a")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "f.txt") || strings.Contains(got, "match_b") {
		t.Errorf("path scope broken: %q", got)
	}
}

func TestGrep_NoMatches(t *testing.T) {
	sb := newBox(t)
	got, _ := sb.Grep("never-here", "")
	if !strings.Contains(got, "no matches") {
		t.Errorf("got %q", got)
	}
}

func TestCall_UnknownTool(t *testing.T) {
	sb := newBox(t)
	if _, err := sb.Call(context.Background(), "Bogus", nil); err == nil {
		t.Fatal("expected unknown tool error")
	}
}

func TestResolve_RejectsEscape(t *testing.T) {
	sb := newBox(t)
	// Reading ".." should error rather than succeeding.
	if _, err := sb.Read("../../etc/passwd"); err == nil {
		t.Error("expected escape rejection")
	}
}

func TestResolve_AbsolutePathOutsideSandboxRejected(t *testing.T) {
	sb := newBox(t)
	if _, err := sb.Read("/etc/passwd"); err == nil {
		t.Error("expected abs-outside rejection")
	}
}

func TestResolve_AbsolutePathInsideSandboxAllowed(t *testing.T) {
	sb := newBox(t)
	abs := filepath.Join(sb.Root, "z.txt")
	if err := os.WriteFile(abs, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := sb.Read(abs)
	if err != nil {
		t.Fatal(err)
	}
	if got != "ok" {
		t.Errorf("got %q", got)
	}
}

func TestRead_LargeFileTruncated(t *testing.T) {
	sb := newBox(t)
	big := strings.Repeat("x", 512*1024) // 512 KiB
	if err := os.WriteFile(filepath.Join(sb.Root, "big"), []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := sb.Read("big")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "truncated") {
		t.Errorf("missing truncation marker: len=%d", len(got))
	}
}

func TestRead_MissingFile(t *testing.T) {
	sb := newBox(t)
	if _, err := sb.Read("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestStringArg_Missing(t *testing.T) {
	// Indirectly via Call with missing args.
	sb := newBox(t)
	_, err := sb.Call(context.Background(), "Read", nil)
	// Read("") hits os.ReadFile which errors on an empty path.
	if err == nil {
		t.Error("expected error on missing path arg")
	}
}
