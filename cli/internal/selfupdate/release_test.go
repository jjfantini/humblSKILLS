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

func TestNormalizeChannel(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ChannelStable},
		{"stable", ChannelStable},
		{"nightly", ChannelStable},
		{"beta", ChannelBeta},
	}
	for _, c := range cases {
		if got := NormalizeChannel(c.in); got != c.want {
			t.Errorf("NormalizeChannel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLatestReleaseForChannel_StableHitsLatest(t *testing.T) {
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/jjfantini/humblSKILLS/releases/latest" {
			t.Errorf("stable channel should hit /releases/latest, got %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"tag_name": "v2.51.0", "prerelease": false}`))
	}))
	rel, err := LatestReleaseForChannel(srv.Client(), DefaultRepo, ChannelStable)
	if err != nil {
		t.Fatalf("LatestReleaseForChannel: %v", err)
	}
	if rel.TagName != "v2.51.0" {
		t.Errorf("TagName = %q, want v2.51.0", rel.TagName)
	}
}

func TestLatestReleaseForChannel_BetaPicksStableWhenNewerThanPre(t *testing.T) {
	// Jennings dry-run: 2.52.0 > 2.52.0-pre.1, so beta must pick the stable.
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/jjfantini/humblSKILLS/releases/latest":
			_, _ = w.Write([]byte(`{"tag_name": "v2.52.0", "prerelease": false}`))
		case "/repos/jjfantini/humblSKILLS/releases":
			_, _ = w.Write([]byte(`[
				{"tag_name": "v2.52.0", "prerelease": false, "draft": false},
				{"tag_name": "v2.52.0-pre.2", "prerelease": true, "draft": true},
				{"tag_name": "v2.52.0-pre.1", "prerelease": true, "draft": false}
			]`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	rel, err := LatestReleaseForChannel(srv.Client(), DefaultRepo, ChannelBeta)
	if err != nil {
		t.Fatalf("LatestReleaseForChannel: %v", err)
	}
	if rel.TagName != "v2.52.0" {
		t.Errorf("TagName = %q, want v2.52.0 (stable > 2.52.0-pre.1)", rel.TagName)
	}
}

func TestLatestReleaseForChannel_BetaPicksPreWhenNewerThanStable(t *testing.T) {
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/jjfantini/humblSKILLS/releases/latest":
			_, _ = w.Write([]byte(`{"tag_name": "v2.52.0", "prerelease": false}`))
		case "/repos/jjfantini/humblSKILLS/releases":
			_, _ = w.Write([]byte(`[
				{"tag_name": "v2.53.0-pre.1", "prerelease": true, "draft": false},
				{"tag_name": "v2.52.0", "prerelease": false, "draft": false}
			]`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	rel, err := LatestReleaseForChannel(srv.Client(), DefaultRepo, ChannelBeta)
	if err != nil {
		t.Fatalf("LatestReleaseForChannel: %v", err)
	}
	if rel.TagName != "v2.53.0-pre.1" {
		t.Errorf("TagName = %q, want v2.53.0-pre.1 (pre > 2.52.0)", rel.TagName)
	}
}

func TestLatestReleaseForChannel_StableNeverPicksPre(t *testing.T) {
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/jjfantini/humblSKILLS/releases/latest" {
			t.Errorf("stable channel should hit /releases/latest, got %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"tag_name": "v2.52.0", "prerelease": false}`))
	}))
	rel, err := LatestReleaseForChannel(srv.Client(), DefaultRepo, ChannelStable)
	if err != nil {
		t.Fatalf("LatestReleaseForChannel: %v", err)
	}
	if rel.TagName != "v2.52.0" {
		t.Errorf("TagName = %q, want v2.52.0", rel.TagName)
	}
}

func TestLatestReleaseForChannel_BetaPicksPreTagWithoutFlag(t *testing.T) {
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/jjfantini/humblSKILLS/releases/latest":
			_, _ = w.Write([]byte(`{"tag_name": "v2.51.0", "prerelease": false}`))
		case "/repos/jjfantini/humblSKILLS/releases":
			_, _ = w.Write([]byte(`[
				{"tag_name": "v2.51.0", "prerelease": false},
				{"tag_name": "v2.52.0-pre.1", "prerelease": false}
			]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	rel, err := LatestReleaseForChannel(srv.Client(), DefaultRepo, "beta")
	if err != nil {
		t.Fatalf("LatestReleaseForChannel: %v", err)
	}
	if rel.TagName != "v2.52.0-pre.1" {
		t.Errorf("TagName = %q, want v2.52.0-pre.1 (dash in tag beats 2.51.0)", rel.TagName)
	}
}

func TestLatestReleaseForChannel_BetaFallsBackToStableWhenNoPre(t *testing.T) {
	srv := withFakeGitHubAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/jjfantini/humblSKILLS/releases/latest":
			_, _ = w.Write([]byte(`{"tag_name": "v2.51.0", "prerelease": false}`))
		case "/repos/jjfantini/humblSKILLS/releases":
			_, _ = w.Write([]byte(`[
				{"tag_name": "v2.51.0", "prerelease": false, "draft": false}
			]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	rel, err := LatestReleaseForChannel(srv.Client(), DefaultRepo, ChannelBeta)
	if err != nil {
		t.Fatalf("beta with only a stable should pick it, got %v", err)
	}
	if rel.TagName != "v2.51.0" {
		t.Errorf("TagName = %q, want v2.51.0", rel.TagName)
	}
}
