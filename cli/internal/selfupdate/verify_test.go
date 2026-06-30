package selfupdate

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseChecksums(t *testing.T) {
	body := "deadbeef  humblskills_2.17.0_linux_amd64.tar.gz\n" +
		"c0ffee *humblskills_2.17.0_macos_arm64.tar.gz\n" +
		"\n" + // blank line should be skipped
		"badline\n" // single field should be skipped

	sums := parseChecksums(body)
	if len(sums) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(sums), sums)
	}
	if got := sums["humblskills_2.17.0_linux_amd64.tar.gz"]; got != "deadbeef" {
		t.Errorf("linux sum = %q, want deadbeef", got)
	}
	if got := sums["humblskills_2.17.0_macos_arm64.tar.gz"]; got != "c0ffee" {
		t.Errorf("macos sum (binary-mode '*') = %q, want c0ffee", got)
	}
}

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "archive.tar.gz")
	content := []byte("fake archive contents")
	if err := os.WriteFile(archivePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	want, err := sha256File(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("match", func(t *testing.T) {
		sums := map[string]string{"archive.tar.gz": want}
		if err := VerifyChecksum(archivePath, "archive.tar.gz", sums); err != nil {
			t.Errorf("expected match, got error: %v", err)
		}
	})

	t.Run("mismatch", func(t *testing.T) {
		sums := map[string]string{"archive.tar.gz": "0000000000000000000000000000000000000000000000000000000000000000"}
		if err := VerifyChecksum(archivePath, "archive.tar.gz", sums); err == nil {
			t.Error("expected checksum mismatch error, got nil")
		}
	})

	t.Run("missing entry", func(t *testing.T) {
		sums := map[string]string{}
		if err := VerifyChecksum(archivePath, "archive.tar.gz", sums); err == nil {
			t.Error("expected missing-entry error, got nil")
		}
	})
}

func TestDownloadToFile(t *testing.T) {
	const payload = "hello from a fake release asset"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != userAgent {
			t.Errorf("missing/unexpected User-Agent: %q", r.Header.Get("User-Agent"))
		}
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.bin")
	if err := downloadToFile(srv.Client(), srv.URL, dest); err != nil {
		t.Fatalf("downloadToFile: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != payload {
		t.Errorf("downloaded content = %q, want %q", got, payload)
	}
}

func TestDownloadToFile_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.bin")
	if err := downloadToFile(srv.Client(), srv.URL, dest); err == nil {
		t.Error("expected error for HTTP 404, got nil")
	}
	if _, err := os.Stat(dest); err == nil {
		t.Error("expected no file to be written on HTTP error")
	}
}

func TestDownloadChecksums(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abc123  humblskills_2.17.0_linux_amd64.tar.gz\n"))
	}))
	defer srv.Close()

	sums, err := downloadChecksums(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("downloadChecksums: %v", err)
	}
	if sums["humblskills_2.17.0_linux_amd64.tar.gz"] != "abc123" {
		t.Errorf("got %v", sums)
	}
}
