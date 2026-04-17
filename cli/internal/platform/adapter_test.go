package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAll_Happy(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "b.yaml", `
name: cursor
detect:
  any_of:
    - path_exists: ~/.cursor
install_targets:
  user: ~/.cursor/skills
default_scope: user
transform: passthrough
`)
	writeYAML(t, dir, "a.yaml", `
name: claude-code
detect:
  any_of:
    - path_exists: ~/.claude
install_targets:
  user: ~/.claude/skills
  project: .claude/skills
default_scope: user
transform: passthrough
`)

	adapters, err := LoadAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(adapters) != 2 {
		t.Fatalf("got %d adapters", len(adapters))
	}
	if adapters[0].Name != "claude-code" || adapters[1].Name != "cursor" {
		t.Errorf("expected sorted output, got %q, %q", adapters[0].Name, adapters[1].Name)
	}
}

func TestLoadAll_SkipsNonYAML(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "a.yaml", "name: foo\n")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	adapters, err := LoadAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(adapters) != 1 {
		t.Errorf("expected 1 adapter, got %d", len(adapters))
	}
}

func TestLoadAll_MissingName(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "a.yaml", "detect: {}\n")
	if _, err := LoadAll(dir); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestNameSet(t *testing.T) {
	set := NameSet([]Adapter{{Name: "a"}, {Name: "b"}})
	if _, ok := set["a"]; !ok {
		t.Error("missing a")
	}
	if _, ok := set["b"]; !ok {
		t.Error("missing b")
	}
	if _, ok := set["c"]; ok {
		t.Error("unexpected c")
	}
}

func writeYAML(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
