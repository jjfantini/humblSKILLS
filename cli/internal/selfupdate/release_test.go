package selfupdate

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func withFakeGitHubAPI(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	prev := GitHubAPIBase
	GitHubAPIBase = srv.URL
	t.Cleanup(func() {
		srv.Close()
		GitHubAPIBase = prev
	})
	return srv
}

func TestLatestRelease(t *testing.T) {
	const body = `{
		"tag_name": "v2.17.0",
		"assets": [
			{"name": "humblskills_2.17.0_linux_amd64.tar.gz", "browser_download_url": "http://example.invalid/a.tar.gz"},
			{"name": "checksums.txt", "browser_download_url": "http://example.invalid/checksums.txt"}
		]
	}`
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/jjfantini/humblSKILLS/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("User-Agent") != userAgent {
			t.Errorf("missing User-Agent header")
		}
		_, _ = w.Write([]byte(body))
	}))

	rel, err := LatestRelease(srv.Client(), DefaultRepo)
	if err != nil {
		t.Fatalf("LatestRelease: %v", err)
	}
	if rel.TagName != "v2.17.0" {
		t.Errorf("TagName = %q, want v2.17.0", rel.TagName)
	}
	if rel.Version() != "2.17.0" {
		t.Errorf("Version() = %q, want 2.17.0", rel.Version())
	}
	asset, ok := rel.Asset("checksums.txt")
	if !ok {
		t.Fatal("expected checksums.txt asset")
	}
	if asset.BrowserDownloadURL != "http://example.invalid/checksums.txt" {
		t.Errorf("unexpected checksums URL: %s", asset.BrowserDownloadURL)
	}
}

func TestLatestRelease_HTTPError(t *testing.T) {
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	if _, err := LatestRelease(srv.Client(), DefaultRepo); err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
}

func TestLatestRelease_MissingTagName(t *testing.T) {
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"assets": []}`))
	}))
	if _, err := LatestRelease(srv.Client(), DefaultRepo); err == nil {
		t.Error("expected error for missing tag_name, got nil")
	}
}
