package toolbox

import (
	"context"
	"strings"
	"testing"
)

func TestReadWriteRoundtrip(t *testing.T) {
	s, err := NewSandbox(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write("a/b.txt", "hello"); err != nil {
		t.Fatal(err)
	}
	got, err := s.Read("a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("roundtrip: %q", got)
	}
}

func TestEscapeAttemptRejected(t *testing.T) {
	s, _ := NewSandbox(t.TempDir())
	if _, err := s.Read("../../etc/passwd"); err == nil {
		t.Fatalf("expected escape rejection")
	}
	if _, err := s.Write("../out.txt", "hi"); err == nil {
		t.Fatalf("expected escape rejection on write")
	}
}

func TestGlobAndGrep(t *testing.T) {
	s, _ := NewSandbox(t.TempDir())
	_, _ = s.Write("a.md", "hello world\nsecond line\n")
	_, _ = s.Write("deep/nested/b.md", "hi there\n")
	_, _ = s.Write("c.txt", "ignore me\n")

	md, err := s.Glob("*.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "a.md") {
		t.Fatalf("expected a.md in glob output, got %q", md)
	}

	hits, err := s.Grep("second", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(hits, "a.md:2") {
		t.Fatalf("unexpected grep output: %q", hits)
	}
}

func TestBashRuns(t *testing.T) {
	s, _ := NewSandbox(t.TempDir())
	out, err := s.Bash(context.Background(), "echo hi")
	if err != nil {
		t.Fatalf("Bash: %v", err)
	}
	if !strings.Contains(out, "hi") {
		t.Fatalf("Bash output: %q", out)
	}
}

func TestCallRoutesByName(t *testing.T) {
	s, _ := NewSandbox(t.TempDir())
	if _, err := s.Call(context.Background(), "Write", map[string]any{"path": "x", "content": "y"}); err != nil {
		t.Fatalf("Write via Call: %v", err)
	}
	out, err := s.Call(context.Background(), "Read", map[string]any{"path": "x"})
	if err != nil {
		t.Fatalf("Read via Call: %v", err)
	}
	if out != "y" {
		t.Fatalf("roundtrip: %q", out)
	}
}
