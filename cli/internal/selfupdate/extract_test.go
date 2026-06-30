package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func buildTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "archive.tar.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Size: int64(len(content)), Mode: 0o755}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func buildZip(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "archive.zip")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractBinary_TarGz(t *testing.T) {
	const want = "#!/bin/sh\necho fake binary\n"
	archive := buildTarGz(t, map[string]string{
		"humblskills_2.17.0_linux_amd64/humblskills": want,
		"humblskills_2.17.0_linux_amd64/LICENSE":     "MIT",
		"humblskills_2.17.0_linux_amd64/README.md":   "readme",
	})

	dest := filepath.Join(t.TempDir(), "humblskills.new")
	if err := ExtractBinary(archive, "humblskills", dest); err != nil {
		t.Fatalf("ExtractBinary: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("extracted content = %q, want %q", got, want)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(dest)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o100 == 0 {
			t.Errorf("expected extracted binary to be executable, mode = %v", info.Mode())
		}
	}
}

func TestExtractBinary_Zip(t *testing.T) {
	const want = "fake windows exe bytes"
	archive := buildZip(t, map[string]string{
		"humblskills_2.17.0_windows_amd64/humblskills.exe": want,
		"humblskills_2.17.0_windows_amd64/LICENSE":         "MIT",
	})

	dest := filepath.Join(t.TempDir(), "humblskills.exe.new")
	if err := ExtractBinary(archive, "humblskills.exe", dest); err != nil {
		t.Fatalf("ExtractBinary: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("extracted content = %q, want %q", got, want)
	}
}

func TestExtractBinary_NotFound(t *testing.T) {
	archive := buildTarGz(t, map[string]string{"some/other-file": "x"})
	dest := filepath.Join(t.TempDir(), "humblskills.new")
	if err := ExtractBinary(archive, "humblskills", dest); err == nil {
		t.Error("expected error when binary isn't present in archive, got nil")
	}
}
