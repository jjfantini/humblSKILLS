package jsonutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFile_CreatesDirsAndIndents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "out.json")
	in := map[string]any{"b": 2, "a": 1}
	if err := WriteFile(path, in); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(raw), "\n") {
		t.Error("expected trailing newline")
	}
	if !strings.Contains(string(raw), "  \"a\"") {
		t.Errorf("expected two-space indentation, got:\n%s", raw)
	}
	var out map[string]int
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if out["a"] != 1 || out["b"] != 2 {
		t.Errorf("unexpected content: %+v", out)
	}
}

func TestWriteFile_LeavesNoTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.json")
	if err := WriteFile(path, []int{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file should be renamed away, stat err = %v", err)
	}
}
