package fetch_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFetch_SendsBearerToken verifies that a configured Token is sent as a
// Bearer Authorization header on the codeload tarball fetch, so skill content
// can be pulled from a private repo.
func TestFetch_SendsBearerToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write([]byte("tarball-bytes"))
	}))
	defer srv.Close()

	f := fetcherWithMockHTTP(t.TempDir(), srv)
	f.Token = "s3cret"

	if _, err := f.Fetch("owner/repo", "deadbeef0000"); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if want := "Bearer s3cret"; gotAuth != want {
		t.Errorf("Authorization header: got %q, want %q", gotAuth, want)
	}
}

// TestFetch_NoAuthHeaderWithoutToken verifies no Authorization header is sent
// when Token is empty.
func TestFetch_NoAuthHeaderWithoutToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write([]byte("tarball-bytes"))
	}))
	defer srv.Close()

	f := fetcherWithMockHTTP(t.TempDir(), srv)

	if _, err := f.Fetch("owner/repo", "deadbeef0000"); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("expected no Authorization header, got %q", gotAuth)
	}
}
