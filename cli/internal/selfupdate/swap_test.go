package selfupdate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReplaceBinary(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "humblskills")
	newBin := filepath.Join(dir, "humblskills.new")

	if err := os.WriteFile(target, []byte("old version"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newBin, []byte("new version"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := ReplaceBinary(target, newBin); err != nil {
		t.Fatalf("ReplaceBinary: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new version" {
		t.Errorf("target content = %q, want %q", got, "new version")
	}
	if _, err := os.Stat(newBin); !os.IsNotExist(err) {
		t.Errorf("expected newBin to be moved (gone), stat err = %v", err)
	}
	if _, err := os.Stat(target + ".old"); !os.IsNotExist(err) {
		t.Errorf("expected .old leftover to be cleaned up, stat err = %v", err)
	}
}

// TestReplaceBinary_WhileOpen simulates swapping the binary while another
// file handle still has it open — the scenario that matters most, since
// targetPath is normally the currently *running* executable.
//
// POSIX-only: a plain os.Open handle doesn't request FILE_SHARE_DELETE on
// Windows, so renaming it there fails with ERROR_SHARING_VIOLATION — that's
// a real Windows restriction on ordinary open file handles, not a bug in
// ReplaceBinary. The actual self-upgrade scenario (renaming a *running*
// .exe) relies on the OS loader's own, more permissive image-sharing
// semantics, which this synthetic os.Open-based test can't represent
// without truly exec'ing a process — not something worth doing here when
// every other ReplaceBinary case (the ones that don't depend on Windows'
// running-executable special case) is already covered cross-platform.
func TestReplaceBinary_WhileOpen(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows doesn't allow renaming a plain (non-FILE_SHARE_DELETE) open file handle; see comment above")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "humblskills")
	newBin := filepath.Join(dir, "humblskills.new")

	if err := os.WriteFile(target, []byte("old version"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newBin, []byte("new version"), 0o755); err != nil {
		t.Fatal(err)
	}

	openHandle, err := os.Open(target)
	if err != nil {
		t.Fatal(err)
	}
	defer openHandle.Close()

	if err := ReplaceBinary(target, newBin); err != nil {
		t.Fatalf("ReplaceBinary while open: %v", err)
	}

	// The open handle should still be able to read the *old* content (POSIX
	// rename semantics: the inode stays alive until every fd closes).
	old := make([]byte, len("old version"))
	if _, err := openHandle.Read(old); err != nil {
		t.Fatalf("read from still-open old handle: %v", err)
	}
	if string(old) != "old version" {
		t.Errorf("old handle content = %q, want %q", old, "old version")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new version" {
		t.Errorf("target content after swap = %q, want %q", got, "new version")
	}
}

func TestReplaceBinary_NewBinaryMissing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "humblskills")
	if err := os.WriteFile(target, []byte("old version"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := ReplaceBinary(target, filepath.Join(dir, "does-not-exist"))
	if err == nil {
		t.Fatal("expected error when newBinaryPath doesn't exist")
	}

	// Revert should have restored the original binary at target.
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("expected target to be restored after failed swap, read err: %v", readErr)
	}
	if string(got) != "old version" {
		t.Errorf("target content after failed swap = %q, want %q (revert)", got, "old version")
	}
}

func TestIsPermissionError(t *testing.T) {
	dir := t.TempDir()
	roDir := filepath.Join(dir, "ro")
	if err := os.Mkdir(roDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(roDir, "humblskills")
	if err := os.WriteFile(target, []byte("old version"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Renaming target requires write+execute on its parent directory, not
	// on target itself — lock that down only after the file already exists.
	if err := os.Chmod(roDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(roDir, 0o755) })

	newBin := filepath.Join(dir, "humblskills.new")
	if err := os.WriteFile(newBin, []byte("new version"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := ReplaceBinary(target, newBin)
	if err == nil {
		t.Skip("running as a user that can write to a 0555 dir (e.g. root) — skipping")
	}
	if !IsPermissionError(err) {
		t.Errorf("expected IsPermissionError(true) for %v", err)
	}
}
