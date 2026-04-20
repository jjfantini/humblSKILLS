package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSetsMissingOnly(t *testing.T) {
	dir := t.TempDir()
	// Create .git marker so findDotEnv stops here.
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `# comment
FOO=one
export BAR="two with spaces"
BAZ='three'
EXISTING=overwrite-me
MALFORMED
`
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	// EXISTING is already set in env - file should not overwrite.
	t.Setenv("EXISTING", "env-wins")
	t.Setenv("FOO", "") // will be unset for the test
	os.Unsetenv("FOO")
	os.Unsetenv("BAR")
	os.Unsetenv("BAZ")

	res, err := LoadDotEnv(dir)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if res.Path != filepath.Join(dir, ".env") {
		t.Fatalf("Path: %q", res.Path)
	}
	if got := os.Getenv("FOO"); got != "one" {
		t.Fatalf("FOO = %q, want one", got)
	}
	if got := os.Getenv("BAR"); got != "two with spaces" {
		t.Fatalf("BAR = %q, want 'two with spaces'", got)
	}
	if got := os.Getenv("BAZ"); got != "three" {
		t.Fatalf("BAZ = %q, want three", got)
	}
	if got := os.Getenv("EXISTING"); got != "env-wins" {
		t.Fatalf("EXISTING = %q, want env-wins", got)
	}
	// EXISTING should be in Kept; FOO/BAR/BAZ should be in Loaded.
	if len(res.Loaded) != 3 {
		t.Fatalf("Loaded: %v (want 3)", res.Loaded)
	}
	if len(res.Kept) != 1 || res.Kept[0] != "EXISTING" {
		t.Fatalf("Kept: %v (want [EXISTING])", res.Kept)
	}
	// Cleanup so other tests don't see these.
	os.Unsetenv("FOO")
	os.Unsetenv("BAR")
	os.Unsetenv("BAZ")
}

func TestMissingDotEnvIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := LoadDotEnv(dir)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if res.Path != "" {
		t.Fatalf("expected empty path, got %q", res.Path)
	}
}

func TestWalkUpFromSubdir(t *testing.T) {
	root := t.TempDir()
	_ = os.Mkdir(filepath.Join(root, ".git"), 0o755)
	sub := filepath.Join(root, "cli", "internal")
	_ = os.MkdirAll(sub, 0o755)
	_ = os.WriteFile(filepath.Join(root, ".env"), []byte("FROM_ROOT=yes\n"), 0o600)

	os.Unsetenv("FROM_ROOT")
	res, err := LoadDotEnv(sub)
	if err != nil {
		t.Fatal(err)
	}
	if res.Path != filepath.Join(root, ".env") {
		t.Fatalf("Path: %q", res.Path)
	}
	if os.Getenv("FROM_ROOT") != "yes" {
		t.Fatalf("FROM_ROOT not set from parent .env")
	}
	os.Unsetenv("FROM_ROOT")
}
