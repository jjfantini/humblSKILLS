package selfupdate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fakeVersionBinary writes a tiny shell script that mimics
// `humblskills version --json`'s output, so VerifyInstalledVersion can be
// tested without a real build of the CLI.
func fakeVersionBinary(t *testing.T, version string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fake binary isn't supported on windows")
	}
	path := filepath.Join(t.TempDir(), "fake-humblskills")
	script := "#!/bin/sh\necho '{\"version\":\"" + version + "\",\"commit\":\"abc123\"}'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestVerifyInstalledVersion(t *testing.T) {
	bin := fakeVersionBinary(t, "2.17.0")
	got, err := VerifyInstalledVersion(bin)
	if err != nil {
		t.Fatalf("VerifyInstalledVersion: %v", err)
	}
	if got != "2.17.0" {
		t.Errorf("got %q, want 2.17.0", got)
	}
}

func TestVerifyInstalledVersion_EmptyVersion(t *testing.T) {
	bin := fakeVersionBinary(t, "")
	if _, err := VerifyInstalledVersion(bin); err == nil {
		t.Error("expected error for empty version field, got nil")
	}
}

func TestVerifyInstalledVersion_NotExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission semantics differ on windows")
	}
	path := filepath.Join(t.TempDir(), "not-executable")
	if err := os.WriteFile(path, []byte("not a script"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyInstalledVersion(path); err == nil {
		t.Error("expected error for a non-executable file, got nil")
	}
}
