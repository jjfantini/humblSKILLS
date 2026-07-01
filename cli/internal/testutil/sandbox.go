// Package testutil provides shared fixtures for humblskills tests:
// isolated filesystem/env sandboxes, an HTTP registry test server,
// tarball seeders, fake keyrings, golden-tree assertions, and a
// configurable fake eval runner.
//
// This package is test-only. Nothing here should be imported by
// production code.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

// Sandbox is an isolated filesystem + env for one test. Every XDG
// base directory, $HOME, and the resolved humblskills paths live
// under t.TempDir(), so tests never touch real user state.
type Sandbox struct {
	Root         string // t.TempDir()
	Home         string
	XDGDataHome  string
	XDGConfigHome string
	XDGStateHome string
	XDGCacheHome string

	// CacheDir is humblskills' cache root (tarballs + registry.json).
	CacheDir string
	// ManifestPath is the install manifest the CLI will read/write.
	ManifestPath string
	// ProfilePath is the profile config the CLI will read/write.
	ProfilePath string
	// SecretsPath is the file-fallback path for secrets.
	SecretsPath string
}

// NewSandbox returns a fully-isolated Sandbox for the current test.
// On return, HOME and all XDG_* env vars point inside t.TempDir(),
// and xdg.Reload() has been called so xdg.ConfigFile/StateFile
// resolve to sandboxed paths. Cleanup restores the previous env.
func NewSandbox(t testing.TB) *Sandbox {
	t.Helper()
	root := t.TempDir()

	s := &Sandbox{
		Root:          root,
		Home:          filepath.Join(root, "home"),
		XDGDataHome:   filepath.Join(root, "xdg", "data"),
		XDGConfigHome: filepath.Join(root, "xdg", "config"),
		XDGStateHome:  filepath.Join(root, "xdg", "state"),
		XDGCacheHome:  filepath.Join(root, "xdg", "cache"),
	}
	for _, d := range []string{s.Home, s.XDGDataHome, s.XDGConfigHome, s.XDGStateHome, s.XDGCacheHome} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	s.CacheDir = filepath.Join(s.XDGCacheHome, "humblskills")
	s.ManifestPath = filepath.Join(s.XDGStateHome, "humblskills", "manifest.json")
	s.ProfilePath = filepath.Join(s.Home, ".humblskills", "profile.json")
	s.SecretsPath = filepath.Join(s.XDGConfigHome, "humblskills", "secrets.json")

	setenv(t, "HOME", s.Home)
	// Windows resolves os.UserHomeDir() via %USERPROFILE%; without this
	// override, ~-expansion and adapter detection on Windows CI still
	// points at the real runner home rather than the sandbox.
	setenv(t, "USERPROFILE", s.Home)
	setenv(t, "XDG_DATA_HOME", s.XDGDataHome)
	setenv(t, "XDG_CONFIG_HOME", s.XDGConfigHome)
	setenv(t, "XDG_STATE_HOME", s.XDGStateHome)
	setenv(t, "XDG_CACHE_HOME", s.XDGCacheHome)
	// Neutralise HUMBLSKILLS_* vars so tests can't pick up the caller's
	// shell config. Tests that want to exercise these opt back in.
	setenv(t, "HUMBLSKILLS_REGISTRY", "")
	setenv(t, "HUMBLSKILLS_CACHE_DIR", "")
	setenv(t, "HUMBLSKILLS_MANIFEST", "")
	setenv(t, "HUMBLSKILLS_PROFILE", "")
	// Provider API keys must not leak from the developer's shell into
	// secrets-resolution tests. Tests that need a key set it explicitly.
	setenv(t, "ANTHROPIC_API_KEY", "")
	setenv(t, "OPENAI_API_KEY", "")

	// xdg package snapshots env at package init; reload so our overrides
	// actually take effect for xdg.ConfigFile/StateFile/DataFile callers.
	xdg.Reload()
	t.Cleanup(xdg.Reload)

	return s
}

// setenv sets an env var and registers a t.Cleanup that restores the
// previous value (or unsets it if previously unset).
func setenv(t testing.TB, key, value string) {
	t.Helper()
	prev, had := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("setenv %s: %v", key, err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, prev)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

// Setenv exposes setenv so tests can override individual vars while
// still getting cleanup tracking.
func (s *Sandbox) Setenv(t testing.TB, key, value string) {
	t.Helper()
	setenv(t, key, value)
}

// WriteFile writes data under the sandbox root at the given relative
// path, creating parents as needed. Returns the absolute path.
func (s *Sandbox) WriteFile(t testing.TB, rel string, data []byte) string {
	t.Helper()
	abs := filepath.Join(s.Root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
	}
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", abs, err)
	}
	return abs
}
