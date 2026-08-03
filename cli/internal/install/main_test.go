package install

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

// TestMain isolates HOME and every XDG base dir for the whole package.
//
// Tests that install at "user" scope resolve their canonical store through
// xdg.DataFile, which on a developer machine is the *real* data dir
// (~/Library/Application Support on macOS). Without this they write skills
// into real user state and — worse — read each other's leftovers back on the
// next run, because the engine now treats an existing store as a preserve
// source. That made results depend on whether a previous run had crashed.
func TestMain(m *testing.M) {
	code, err := runIsolated(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, "install tests: isolate env:", err)
		os.Exit(1)
	}
	os.Exit(code)
}

// newTestEngine builds an Engine whose canonical store lives under this test's
// own temp root.
//
// Nearly every test in this package installs a skill called "foo" at user
// scope, and that store path resolves through xdg.DataFile — one shared
// directory for the whole package. Without per-test isolation each test reads
// the previous test's store back as a preserve source, so a failure anywhere
// cascades into unrelated tests.
func newTestEngine(t *testing.T, root, cacheDir, manifestPath string) *Engine {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "xdg-data"))
	// --global stores resolve through os.UserHomeDir(), not xdg.
	t.Setenv("HOME", filepath.Join(root, "home"))
	t.Setenv("USERPROFILE", filepath.Join(root, "home"))
	xdg.Reload()
	t.Cleanup(xdg.Reload)
	return NewEngine(cacheDir, manifestPath)
}

func runIsolated(m *testing.M) (int, error) {
	root, err := os.MkdirTemp("", "humblskills-install-tests-")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(root)

	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		return 0, err
	}
	env := map[string]string{
		"HOME":            home,
		"USERPROFILE":     home, // os.UserHomeDir() on Windows
		"XDG_DATA_HOME":   filepath.Join(root, "xdg", "data"),
		"XDG_CONFIG_HOME": filepath.Join(root, "xdg", "config"),
		"XDG_STATE_HOME":  filepath.Join(root, "xdg", "state"),
		"XDG_CACHE_HOME":  filepath.Join(root, "xdg", "cache"),
	}
	for k, v := range env {
		if err := os.MkdirAll(v, 0o755); err != nil {
			return 0, err
		}
		if err := os.Setenv(k, v); err != nil {
			return 0, err
		}
	}
	// xdg snapshots the environment at package init, so the overrides above
	// only reach xdg.DataFile after a reload.
	xdg.Reload()

	return m.Run(), nil
}
