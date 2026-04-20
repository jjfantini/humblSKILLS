package fetch_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/fetch"
)

// tarballBytes returns a gzipped tar with one top-level directory and
// the given files underneath. Shared across the coverage-lift tests.
func tarballBytes(t *testing.T, prefix string, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: prefix + "/", Typeflag: tar.TypeDir, Mode: 0o755})
	for name, body := range files {
		_ = tw.WriteHeader(&tar.Header{
			Name:     prefix + "/" + name,
			Typeflag: tar.TypeReg,
			Mode:     0o644,
			Size:     int64(len(body)),
		})
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

// fetcherWithMockHTTP builds a Fetcher whose HTTP client is redirected
// to srv regardless of URL. We intercept all outbound requests via a
// RoundTripper shim because production code hard-codes codeload URLs.
func fetcherWithMockHTTP(cacheDir string, srv *httptest.Server) *fetch.Fetcher {
	f := fetch.NewFetcher(cacheDir)
	f.HTTP = &http.Client{
		Timeout: 5 * time.Second,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Rewrite to point at the httptest server.
			u := *req.URL
			u.Scheme = "http"
			u.Host = srv.Listener.Addr().String()
			req.URL = &u
			req.Host = u.Host
			return http.DefaultTransport.RoundTrip(req)
		}),
	}
	return f
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetch_WritesTarballIntoCache(t *testing.T) {
	body := tarballBytes(t, "o-r-abcd", map[string]string{
		"skills/foo/SKILL.md": "# foo",
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/tar.gz/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	cache := t.TempDir()
	f := fetcherWithMockHTTP(cache, srv)

	path, err := f.Fetch("owner/repo", "deadbeef0000")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !strings.HasPrefix(path, cache) {
		t.Errorf("path %q not under cache %q", path, cache)
	}
	// Second Fetch must be a cache hit (no network).
	srv.Close()
	path2, err := f.Fetch("owner/repo", "deadbeef0000")
	if err != nil {
		t.Fatalf("Fetch (cache hit): %v", err)
	}
	if path != path2 {
		t.Errorf("cache hit path differs: %q vs %q", path, path2)
	}
}

func TestFetch_EmptySHA(t *testing.T) {
	f := fetch.NewFetcher(t.TempDir())
	if _, err := f.Fetch("o/r", ""); err == nil {
		t.Fatal("expected error for empty sha")
	}
}

func TestFetch_InvalidRepo(t *testing.T) {
	f := fetch.NewFetcher(t.TempDir())
	if _, err := f.Fetch("not-a-repo", "abc"); err == nil {
		t.Fatal("expected error for malformed repo")
	}
}

func TestFetch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()

	f := fetcherWithMockHTTP(t.TempDir(), srv)
	_, err := f.Fetch("owner/repo", "deadbeef0000")
	if err == nil {
		t.Fatal("expected HTTP error to surface")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("err missing status: %v", err)
	}
}

func TestFetch_NetworkFailure(t *testing.T) {
	srv := httptest.NewServer(nil)
	addr := srv.Listener.Addr().String()
	srv.Close() // kill it immediately

	f := fetch.NewFetcher(t.TempDir())
	f.HTTP = &http.Client{
		Timeout: 500 * time.Millisecond,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			u := *req.URL
			u.Scheme = "http"
			u.Host = addr
			req.URL = &u
			return http.DefaultTransport.RoundTrip(req)
		}),
	}

	if _, err := f.Fetch("owner/repo", "sha"); err == nil {
		t.Fatal("expected network error")
	}
}

func TestExtract_EmptySkillPath(t *testing.T) {
	body := tarballBytes(t, "pfx", map[string]string{"file.md": "x"})
	tarPath := filepath.Join(t.TempDir(), "a.tar.gz")
	if err := os.WriteFile(tarPath, body, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := fetch.Extract(tarPath, "", t.TempDir()); err == nil {
		t.Fatal("expected error for empty skill path")
	}
	if err := fetch.Extract(tarPath, ".", t.TempDir()); err == nil {
		t.Fatal("expected error for '.' skill path")
	}
}

func TestExtract_MissingTarball(t *testing.T) {
	if err := fetch.Extract(filepath.Join(t.TempDir(), "nope.tar.gz"), "skills/foo", t.TempDir()); err == nil {
		t.Fatal("expected error for missing tarball")
	}
}

func TestExtract_NotGzip(t *testing.T) {
	tarPath := filepath.Join(t.TempDir(), "plain.tar.gz")
	if err := os.WriteFile(tarPath, []byte("not gzip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := fetch.Extract(tarPath, "skills/foo", t.TempDir()); err == nil {
		t.Fatal("expected gzip error")
	}
}

func TestExtract_RejectsDotDotEntry(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: "pfx/", Typeflag: tar.TypeDir, Mode: 0o755})
	_ = tw.WriteHeader(&tar.Header{
		Name:     "../escape",
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     3,
	})
	_, _ = tw.Write([]byte("pwn"))
	_ = tw.Close()
	_ = gz.Close()

	tarPath := filepath.Join(t.TempDir(), "evil.tar.gz")
	if err := os.WriteFile(tarPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := fetch.Extract(tarPath, "skills/foo", t.TempDir()); err == nil {
		t.Fatal("expected ../ rejection")
	}
}

func TestExtract_DirectoryEntriesAreCreated(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: "pfx/", Typeflag: tar.TypeDir, Mode: 0o755})
	_ = tw.WriteHeader(&tar.Header{Name: "pfx/skills/foo/", Typeflag: tar.TypeDir, Mode: 0o755})
	_ = tw.WriteHeader(&tar.Header{Name: "pfx/skills/foo/sub/", Typeflag: tar.TypeDir, Mode: 0o755})
	body := "hi"
	_ = tw.WriteHeader(&tar.Header{
		Name: "pfx/skills/foo/sub/README.md", Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(body)),
	})
	_, _ = tw.Write([]byte(body))
	_ = tw.Close()
	_ = gz.Close()

	tarPath := filepath.Join(t.TempDir(), "dirs.tar.gz")
	if err := os.WriteFile(tarPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out")
	if err := fetch.Extract(tarPath, "skills/foo", dest); err != nil {
		t.Fatalf("extract: %v", err)
	}
	info, err := os.Stat(filepath.Join(dest, "sub"))
	if err != nil {
		t.Fatalf("stat sub: %v", err)
	}
	if !info.IsDir() {
		t.Error("sub not a directory")
	}
}

func TestNewFetcher_Defaults(t *testing.T) {
	f := fetch.NewFetcher("/tmp/cache")
	if f.CacheDir != "/tmp/cache" {
		t.Errorf("CacheDir = %q", f.CacheDir)
	}
	if f.HTTP == nil {
		t.Error("HTTP should be non-nil")
	}
	if f.HTTP.Timeout != fetch.DefaultHTTPTimeout {
		t.Errorf("Timeout = %v, want %v", f.HTTP.Timeout, fetch.DefaultHTTPTimeout)
	}
}
