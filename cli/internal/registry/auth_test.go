package registry

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHTTPFetchError_ActionableMessages verifies auth failures map to guidance.
func TestHTTPFetchError_ActionableMessages(t *testing.T) {
	if err := httpFetchError("u", 401, true); err == nil || !strings.Contains(err.Error(), "registry login") {
		t.Errorf("401 should mention `registry login`: %v", err)
	}
	if err := httpFetchError("u", 404, false); err == nil || !strings.Contains(err.Error(), "private registry") {
		t.Errorf("404 without token should hint at a private registry: %v", err)
	}
	if err := httpFetchError("u", 500, false); err == nil || !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("other codes keep the bare status: %v", err)
	}
}

// TestFetch_SendsBearerToken verifies that a configured Token is sent as a
// Bearer Authorization header on the registry HTTP fetch, so a private registry
// can be read.
func TestFetch_SendsBearerToken(t *testing.T) {
	body := testRegistryBody(t)

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write(body)
	}))
	defer srv.Close()

	f := NewFetcher(srv.URL, t.TempDir())
	f.Token = "s3cret"

	if _, _, err := f.Load(); err != nil {
		t.Fatal(err)
	}
	if want := "Bearer s3cret"; gotAuth != want {
		t.Errorf("Authorization header: got %q, want %q", gotAuth, want)
	}
}

// TestFetch_NoAuthHeaderWithoutToken verifies that no Authorization header is
// sent when Token is empty (the public default).
func TestFetch_NoAuthHeaderWithoutToken(t *testing.T) {
	body := testRegistryBody(t)

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write(body)
	}))
	defer srv.Close()

	f := NewFetcher(srv.URL, t.TempDir())

	if _, _, err := f.Load(); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "" {
		t.Errorf("expected no Authorization header, got %q", gotAuth)
	}
}
