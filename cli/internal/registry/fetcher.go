package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultURL is the canonical hosted location for the humblSKILLS registry.
const DefaultURL = "https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/registry.json"

// DefaultTTL is how long a cached registry.json is considered fresh.
const DefaultTTL = 10 * time.Minute

// Origin reports where a Registry came from on a given Load call.
type Origin string

const (
	OriginCache   Origin = "cache"
	OriginNetwork Origin = "network"
	OriginFile    Origin = "file"
)

// Fetcher loads Registry documents, caching HTTP responses on disk.
type Fetcher struct {
	URL      string
	CacheDir string
	TTL      time.Duration
	HTTP     *http.Client
	Now      func() time.Time
	// Token, when non-empty, is sent as a Bearer Authorization header on the
	// HTTP registry fetch so a private registry can be read. Ignored for
	// file:// URLs and bare paths.
	Token string
}

// NewFetcher returns a Fetcher with sensible defaults. cacheDir should be the
// humblskills-specific cache directory (e.g. $XDG_CACHE_HOME/humblskills).
func NewFetcher(regURL, cacheDir string) *Fetcher {
	return &Fetcher{
		URL:      regURL,
		CacheDir: cacheDir,
		TTL:      DefaultTTL,
		HTTP:     &http.Client{Timeout: 15 * time.Second},
		Now:      time.Now,
	}
}

// CacheInfo describes the on-disk cache state for this Fetcher's URL.
type CacheInfo struct {
	URL       string        `json:"url"`
	Path      string        `json:"path"`
	Exists    bool          `json:"exists"`
	FetchedAt time.Time     `json:"fetched_at,omitempty"`
	Age       time.Duration `json:"age_seconds,omitempty"`
}

type cacheMeta struct {
	URL       string    `json:"url"`
	FetchedAt time.Time `json:"fetched_at"`
}

// Load returns the registry, using the cache when fresh. For file:// URLs or
// bare filesystem paths, reads directly and skips the cache.
func (f *Fetcher) Load() (*Registry, Origin, error) {
	if isLocal(f.URL) {
		r, err := f.loadLocal()
		return r, OriginFile, err
	}

	if r, ok := f.tryCache(); ok {
		return r, OriginCache, nil
	}
	r, err := f.fetchAndCache()
	if err != nil {
		return nil, "", err
	}
	return r, OriginNetwork, nil
}

// Refresh forces a network (or file) reload, ignoring the cache TTL. HTTP
// URLs refresh the cache; file URLs do not write the cache.
func (f *Fetcher) Refresh() (*Registry, Origin, error) {
	if isLocal(f.URL) {
		r, err := f.loadLocal()
		return r, OriginFile, err
	}
	r, err := f.fetchAndCache()
	if err != nil {
		return nil, "", err
	}
	return r, OriginNetwork, nil
}

// LoadCached returns the registry from local sources only, never hitting the
// network. A file:// URL (or bare path) is read directly; an http(s) URL is
// served from the on-disk cache if present, ignoring the TTL. ok is false when
// no local copy is available.
//
// This is for read-only views (e.g. `list`) that must stay fast and
// offline-friendly and should not trigger a fetch. Freshness is a separate
// concern: `registry refresh`, `update`, and `search` refill the cache.
func (f *Fetcher) LoadCached() (*Registry, bool) {
	if isLocal(f.URL) {
		r, err := f.loadLocal()
		if err != nil {
			return nil, false
		}
		return r, true
	}
	m, err := f.readMeta()
	if err != nil || m.URL != f.URL {
		return nil, false
	}
	data, err := os.ReadFile(f.bodyPath())
	if err != nil {
		return nil, false
	}
	r, err := parseRegistry(data)
	if err != nil {
		return nil, false
	}
	return r, true
}

// Inspect reports on-disk cache state without triggering a fetch.
func (f *Fetcher) Inspect() CacheInfo {
	info := CacheInfo{URL: f.URL, Path: f.bodyPath()}
	if isLocal(f.URL) {
		return info
	}
	m, err := f.readMeta()
	if err != nil {
		return info
	}
	if _, err := os.Stat(f.bodyPath()); err != nil {
		return info
	}
	info.Exists = true
	info.FetchedAt = m.FetchedAt
	info.Age = f.Now().Sub(m.FetchedAt)
	return info
}

func (f *Fetcher) tryCache() (*Registry, bool) {
	m, err := f.readMeta()
	if err != nil {
		return nil, false
	}
	if m.URL != f.URL {
		return nil, false
	}
	if f.Now().Sub(m.FetchedAt) > f.TTL {
		return nil, false
	}
	data, err := os.ReadFile(f.bodyPath())
	if err != nil {
		return nil, false
	}
	r, err := parseRegistry(data)
	if err != nil {
		return nil, false
	}
	return r, true
}

func (f *Fetcher) fetchAndCache() (*Registry, error) {
	req, err := http.NewRequest(http.MethodGet, f.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "humblskills-cli")
	if f.Token != "" {
		req.Header.Set("Authorization", "Bearer "+f.Token)
	}

	resp, err := f.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", f.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, httpFetchError(f.URL, resp.StatusCode, f.Token != "")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	r, err := parseRegistry(body)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	if err := f.writeCache(body); err != nil {
		// Cache write failure shouldn't block the caller — log via return value.
		return r, fmt.Errorf("fetch ok but cache write failed: %w", err)
	}
	return r, nil
}

func (f *Fetcher) loadLocal() (*Registry, error) {
	path := localPath(f.URL)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return parseRegistry(data)
}

func (f *Fetcher) bodyPath() string { return filepath.Join(f.CacheDir, "registry.json") }
func (f *Fetcher) metaPath() string { return filepath.Join(f.CacheDir, "registry.meta.json") }
func (f *Fetcher) writeCache(body []byte) error {
	if err := os.MkdirAll(f.CacheDir, 0o755); err != nil {
		return err
	}
	if err := writeAtomic(f.bodyPath(), body); err != nil {
		return err
	}
	m := cacheMeta{URL: f.URL, FetchedAt: f.Now().UTC()}
	metaBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(f.metaPath(), append(metaBytes, '\n'))
}

func (f *Fetcher) readMeta() (cacheMeta, error) {
	var m cacheMeta
	data, err := os.ReadFile(f.metaPath())
	if err != nil {
		return m, err
	}
	return m, json.Unmarshal(data, &m)
}

func parseRegistry(body []byte) (*Registry, error) {
	var r Registry
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("unsupported registry schema_version %d (expected %d)", r.SchemaVersion, SchemaVersion)
	}
	return &r, nil
}

// httpFetchError turns a non-2xx status into an actionable message, mapping the
// common private-registry auth failures to next steps instead of a bare code.
func httpFetchError(url string, status int, hasToken bool) error {
	switch status {
	case 401, 403:
		return fmt.Errorf("fetch %s: HTTP %d — registry token rejected (missing, expired, or lacks access); re-run `humblskills registry login`", url, status)
	case 404:
		if !hasToken {
			return fmt.Errorf("fetch %s: HTTP %d — not found; if this is a private registry, add a token with `humblskills registry login`", url, status)
		}
		return fmt.Errorf("fetch %s: HTTP %d — not found; check the registry URL, or re-run `humblskills registry login` if the token is missing, expired, or lacks access", url, status)
	default:
		return fmt.Errorf("fetch %s: HTTP %d", url, status)
	}
}

func isLocal(u string) bool {
	if strings.HasPrefix(u, "file://") {
		return true
	}
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return false
	}
	// No scheme at all: treat as a local path.
	return true
}

func localPath(u string) string {
	if !strings.HasPrefix(u, "file://") {
		return u
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return strings.TrimPrefix(u, "file://")
	}
	return parsed.Path
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
