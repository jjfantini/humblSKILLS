package env_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/env"
)

func TestLoadDotEnv_ReturnsZeroWhenNoFileFound(t *testing.T) {
	// Point at a guaranteed-empty tempdir with no .env and a .git
	// boundary at the top so the walk terminates quickly.
	top := t.TempDir()
	if err := os.Mkdir(filepath.Join(top, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := env.LoadDotEnv(top)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if res.Path != "" {
		t.Errorf("expected empty path, got %q", res.Path)
	}
	if len(res.Loaded) != 0 || len(res.Kept) != 0 {
		t.Errorf("expected nothing loaded: %+v", res)
	}
}

func TestLoadDotEnv_EmptyStartDirUsesCwd(t *testing.T) {
	// Running with startDir="" must fall back to os.Getwd. We don't
	// necessarily have a .env under the dev's cwd, but the call must
	// not error.
	res, err := env.LoadDotEnv("")
	if err != nil {
		t.Fatalf("LoadDotEnv(\"\"): %v", err)
	}
	_ = res // only asserting no error
}

func TestLoadDotEnv_KeepsAlreadySetEnvVars(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(root, ".env")
	if err := os.WriteFile(envPath, []byte("FOO=from_file\nBAR=from_file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// FOO already set in env → env wins, FOO goes into Kept.
	t.Setenv("FOO", "from_env")
	// BAR unset → ends up in Loaded.
	_ = os.Unsetenv("BAR")
	t.Cleanup(func() { _ = os.Unsetenv("BAR") })

	res, err := env.LoadDotEnv(root)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if !containsStr(res.Kept, "FOO") {
		t.Errorf("FOO not in Kept: %+v", res)
	}
	if !containsStr(res.Loaded, "BAR") {
		t.Errorf("BAR not in Loaded: %+v", res)
	}
	if os.Getenv("FOO") != "from_env" {
		t.Error("FOO was overwritten; env must win")
	}
	if os.Getenv("BAR") != "from_file" {
		t.Errorf("BAR not set to file value; got %q", os.Getenv("BAR"))
	}
}

func TestLoadDotEnv_ParsesExportAndQuotesAndComments(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"# a comment line",
		"",
		"export QUOTED_DOUBLE=\"hello world\"",
		"export QUOTED_SINGLE='single quoted'",
		"NO_EQUALS_SIGN_HERE",
		"=emptykey_should_skip",
		"PLAIN=plainval",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, k := range []string{"QUOTED_DOUBLE", "QUOTED_SINGLE", "PLAIN"} {
		_ = os.Unsetenv(k)
		kk := k
		t.Cleanup(func() { _ = os.Unsetenv(kk) })
	}
	res, err := env.LoadDotEnv(root)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if os.Getenv("QUOTED_DOUBLE") != "hello world" {
		t.Errorf("QUOTED_DOUBLE = %q", os.Getenv("QUOTED_DOUBLE"))
	}
	if os.Getenv("QUOTED_SINGLE") != "single quoted" {
		t.Errorf("QUOTED_SINGLE = %q", os.Getenv("QUOTED_SINGLE"))
	}
	if os.Getenv("PLAIN") != "plainval" {
		t.Errorf("PLAIN = %q", os.Getenv("PLAIN"))
	}
	// Res.Path set to the .env we wrote.
	if res.Path != filepath.Join(root, ".env") {
		t.Errorf("Path = %q", res.Path)
	}
}

func TestLoadDotEnv_FindsNearestUpwards(t *testing.T) {
	root := t.TempDir()
	// Layout:
	//   root/ (has .git boundary AND .env)
	//   root/a/b/c (walk starts here)
	// Nearest .env is at root.
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("UPWARDS=y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.Unsetenv("UPWARDS")
	t.Cleanup(func() { _ = os.Unsetenv("UPWARDS") })

	res, err := env.LoadDotEnv(deep)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if res.Path != filepath.Join(root, ".env") {
		t.Errorf("Path = %q, want root .env", res.Path)
	}
}

func TestLoadDotEnv_StopsAtGitBoundary(t *testing.T) {
	// We place a .env OUTSIDE the git boundary; the walk must stop at
	// .git and not pick it up.
	outer := t.TempDir()
	if err := os.WriteFile(filepath.Join(outer, ".env"), []byte("SHOULD_NOT_LOAD=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	inner := filepath.Join(outer, "repo")
	if err := os.MkdirAll(filepath.Join(inner, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	_ = os.Unsetenv("SHOULD_NOT_LOAD")
	res, err := env.LoadDotEnv(inner)
	if err != nil {
		t.Fatalf("LoadDotEnv: %v", err)
	}
	if res.Path != "" {
		t.Errorf("git boundary not honored: %q", res.Path)
	}
	if got := os.Getenv("SHOULD_NOT_LOAD"); got != "" {
		t.Errorf("SHOULD_NOT_LOAD leaked: %q", got)
	}
}

func TestLoadDotEnv_ScannerErrorPropagatesOnHugeLine(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	// bufio.Scanner default MaxScanTokenSize is ~64K. Writing a single
	// 128K line without a newline triggers bufio.ErrTooLong, which
	// parse() must surface.
	big := strings.Repeat("A=", 1)
	big += strings.Repeat("x", 128*1024)
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := env.LoadDotEnv(root)
	if err == nil {
		t.Fatal("expected scanner error on oversized line")
	}
}

func containsStr(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
