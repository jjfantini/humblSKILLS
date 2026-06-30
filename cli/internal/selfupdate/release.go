// Package selfupdate upgrades the humblskills CLI binary itself to the
// latest published GitHub release: it resolves the latest release, builds
// the platform-specific archive name goreleaser publishes, downloads and
// verifies it against checksums.txt, extracts the binary, and swaps it onto
// the running executable's path. It also recognizes Homebrew-managed
// installs and defers to `brew upgrade` instead of fighting brew's own
// bookkeeping.
//
// This is distinct from internal/install, which upgrades installed
// *skills*; selfupdate upgrades the CLI binary that runs them.
package selfupdate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DefaultRepo is the canonical GitHub repo that publishes humblskills
// releases.
const DefaultRepo = "jjfantini/humblSKILLS"

// DefaultHTTPTimeout matches the timeout used for registry fetches
// (internal/registry.Fetcher) — no retries, fail fast.
const DefaultHTTPTimeout = 15 * time.Second

// userAgent is sent on every selfupdate HTTP request, matching the
// convention in internal/registry and internal/fetch.
const userAgent = "humblskills-cli"

// GitHubAPIBase is the GitHub API origin LatestRelease hits. It's a var
// (not a const), and exported, purely so tests — both inside this package
// and in cmd/humblskills's command-level tests — can point it at an
// httptest.Server instead of the real api.github.com.
var GitHubAPIBase = "https://api.github.com"

// Asset is one file attached to a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Release is the subset of the GitHub releases API response selfupdate
// needs.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Version returns the release's bare semver version, stripping the leading
// "v" goreleaser/release-please tags use (e.g. "v2.17.0" -> "2.17.0").
func (r *Release) Version() string {
	v := r.TagName
	if len(v) > 0 && (v[0] == 'v' || v[0] == 'V') {
		v = v[1:]
	}
	return v
}

// Asset returns the release asset named name, or ok=false if it isn't
// attached to this release.
func (r *Release) Asset(name string) (Asset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}

// NewHTTPClient returns the shared HTTP client used for every selfupdate
// network call: fixed timeout, no retries.
func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: DefaultHTTPTimeout}
}

// LatestRelease fetches https://api.github.com/repos/{repo}/releases/latest.
// client defaults to NewHTTPClient() when nil.
func LatestRelease(client *http.Client, repo string) (*Release, error) {
	if client == nil {
		client = NewHTTPClient()
	}
	url := fmt.Sprintf("%s/repos/%s/releases/latest", GitHubAPIBase, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("fetch %s: response had no tag_name", url)
	}
	return &rel, nil
}
