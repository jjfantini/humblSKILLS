// Package toolbox is the sandboxed Read/Write/Bash/Glob/Grep toolkit shared
// by the anthropic-api and openai-api runners. Every tool resolves paths
// against a single scratch root - escape attempts (".." traversal) return
// an error without ever touching the filesystem outside the scratch.
package toolbox

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Tool is a single callable action the toolbox exposes. Name matches what
// the API-runners advertise as a "tool" to the model. Call executes against
// the given sandbox root.
type Tool struct {
	Name        string
	Description string
	Schema      map[string]any // JSONSchema-ish; provided verbatim to the API runner
}

// Sandbox is a scratch root. Path-typed inputs are resolved relative to Root.
type Sandbox struct {
	Root      string
	ExecTimeout time.Duration
}

// NewSandbox creates a scratch-rooted toolbox. The root directory is
// created if it doesn't exist. ExecTimeout defaults to 60 seconds.
func NewSandbox(root string) (*Sandbox, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Sandbox{Root: abs, ExecTimeout: 60 * time.Second}, nil
}

// DefaultTools returns the five-tool surface the plan specifies.
func DefaultTools() []Tool {
	return []Tool{
		{
			Name:        "Read",
			Description: "Read a file from the sandbox. Returns text contents.",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{"type": "string", "description": "Path relative to sandbox root."},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "Write",
			Description: "Write text to a file inside the sandbox. Creates parent dirs.",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":    map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name:        "Bash",
			Description: "Run a shell command inside the sandbox. Returns combined stdout+stderr.",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{"type": "string"},
				},
				"required": []string{"command"},
			},
		},
		{
			Name:        "Glob",
			Description: "Return paths under the sandbox matching a glob pattern (relative to sandbox root).",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pattern": map[string]any{"type": "string"},
				},
				"required": []string{"pattern"},
			},
		},
		{
			Name:        "Grep",
			Description: "Search file contents for a regex. Returns matching lines with file:line prefix.",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pattern": map[string]any{"type": "string"},
					"path":    map[string]any{"type": "string", "description": "Optional path prefix; defaults to the whole sandbox."},
				},
				"required": []string{"pattern"},
			},
		},
	}
}

// Call routes a tool invocation to the right implementation.
func (s *Sandbox) Call(ctx context.Context, tool string, args map[string]any) (string, error) {
	switch tool {
	case "Read":
		return s.Read(stringArg(args, "path"))
	case "Write":
		return s.Write(stringArg(args, "path"), stringArg(args, "content"))
	case "Bash":
		return s.Bash(ctx, stringArg(args, "command"))
	case "Glob":
		return s.Glob(stringArg(args, "pattern"))
	case "Grep":
		return s.Grep(stringArg(args, "pattern"), stringArg(args, "path"))
	default:
		return "", fmt.Errorf("unknown tool %q", tool)
	}
}

// --- path discipline --------------------------------------------------------

func (s *Sandbox) resolve(rel string) (string, error) {
	if filepath.IsAbs(rel) {
		// Absolute path: must still be under Root.
		if !strings.HasPrefix(rel, s.Root+string(os.PathSeparator)) && rel != s.Root {
			return "", fmt.Errorf("path %q outside sandbox", rel)
		}
		return rel, nil
	}
	p := filepath.Clean(filepath.Join(s.Root, rel))
	if !strings.HasPrefix(p, s.Root+string(os.PathSeparator)) && p != s.Root {
		return "", fmt.Errorf("path %q resolves outside sandbox", rel)
	}
	return p, nil
}

// --- tool implementations ---------------------------------------------------

// Read returns the file contents, capped at 256 KiB so runaway files don't
// blow up the prompt context.
func (s *Sandbox) Read(path string) (string, error) {
	p, err := s.resolve(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	const cap = 256 * 1024
	if len(data) > cap {
		return string(data[:cap]) + "\n... (truncated at 256 KiB)\n", nil
	}
	return string(data), nil
}

// Write creates parent dirs and replaces file contents.
func (s *Sandbox) Write(path, content string) (string, error) {
	p, err := s.resolve(path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), path), nil
}

// Bash runs a shell command inside the sandbox with timeout.
func (s *Sandbox) Bash(ctx context.Context, command string) (string, error) {
	if strings.TrimSpace(command) == "" {
		return "", errors.New("empty command")
	}
	ctx2, cancel := context.WithTimeout(ctx, s.ExecTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx2, "sh", "-c", command)
	cmd.Dir = s.Root
	out, err := cmd.CombinedOutput()
	const cap = 32 * 1024
	if len(out) > cap {
		out = append(out[:cap], []byte("\n... (truncated at 32 KiB)\n")...)
	}
	if err != nil {
		return string(out), fmt.Errorf("bash exit: %w", err)
	}
	return string(out), nil
}

// Glob walks the sandbox and returns matching relative paths.
func (s *Sandbox) Glob(pattern string) (string, error) {
	if pattern == "" {
		return "", errors.New("empty pattern")
	}
	var hits []string
	err := filepath.WalkDir(s.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(s.Root, p)
		if ok, _ := filepath.Match(pattern, filepath.Base(rel)); ok {
			hits = append(hits, rel)
			return nil
		}
		if ok, _ := filepath.Match(pattern, rel); ok {
			hits = append(hits, rel)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(hits) == 0 {
		return "(no matches)", nil
	}
	return strings.Join(hits, "\n"), nil
}

// Grep scans files for a regex. path may be empty to scan the whole sandbox.
func (s *Sandbox) Grep(pattern, path string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	root := s.Root
	if path != "" {
		r, err := s.resolve(path)
		if err != nil {
			return "", err
		}
		root = r
	}
	var hits []string
	limit := 200
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if len(hits) >= limit {
			return filepath.SkipAll
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(s.Root, p)
		for i, line := range strings.Split(string(data), "\n") {
			if re.MatchString(line) {
				hits = append(hits, fmt.Sprintf("%s:%d: %s", rel, i+1, line))
				if len(hits) >= limit {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if len(hits) == 0 {
		return "(no matches)", nil
	}
	return strings.Join(hits, "\n"), nil
}

// --- helpers ----------------------------------------------------------------

func stringArg(args map[string]any, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
