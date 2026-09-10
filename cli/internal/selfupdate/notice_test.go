package selfupdate

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestDisplayVersion(t *testing.T) {
	cases := map[string]string{
		"2.17.0":       "v2.17.0",
		"2.52.0-pre.1": "v2.52.0-pre.1",
		"v2.17.0":      "v2.17.0",
		"dev":          "dev",
		"":             "",
	}
	for in, want := range cases {
		if got := DisplayVersion(in); got != want {
			t.Errorf("DisplayVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNotice_CLILine_GitHubAndHomebrew(t *testing.T) {
	gh := Notice{
		CurrentVersion: "2.15.0",
		LatestVersion:  "2.17.0",
		Channel:        ChannelStable,
		Available:      true,
	}
	wantGH := "newer version available: v2.15.0 → v2.17.0 (stable) — run `humblskills upgrade`"
	if got := gh.CLILine(); got != wantGH {
		t.Errorf("github CLILine = %q, want %q", got, wantGH)
	}

	brew := Notice{
		CurrentVersion: "2.15.0",
		LatestVersion:  "2.17.0",
		Channel:        ChannelStable,
		Homebrew:       true,
		Formula:        FormulaStable,
		Available:      true,
	}
	wantBrew := "newer version available: v2.15.0 → v2.17.0 (stable) — run `brew upgrade humblskills`"
	if got := brew.CLILine(); got != wantBrew {
		t.Errorf("brew CLILine = %q, want %q", got, wantBrew)
	}

	brewBeta := Notice{
		CurrentVersion: "2.51.0",
		LatestVersion:  "2.52.0-pre.1",
		Channel:        ChannelBeta,
		Homebrew:       true,
		Formula:        FormulaPre,
		CurrentFormula: FormulaPre,
		Available:      true,
	}
	wantBeta := "newer version available: v2.51.0 → v2.52.0-pre.1 (beta) — run `brew upgrade humblskills-pre`"
	if got := brewBeta.CLILine(); got != wantBeta {
		t.Errorf("brew beta CLILine = %q, want %q", got, wantBeta)
	}

	switchNotice := Notice{
		CurrentVersion: "2.52.0-pre.1",
		LatestVersion:  "2.52.0",
		Channel:        ChannelBeta,
		Homebrew:       true,
		Formula:        FormulaStable,
		CurrentFormula: FormulaPre,
		Available:      true,
	}
	wantSwitch := "newer version available: v2.52.0-pre.1 → v2.52.0 (beta) — run `brew uninstall humblskills-pre && brew install humblskills`"
	if got := switchNotice.CLILine(); got != wantSwitch {
		t.Errorf("brew switch CLILine = %q, want %q", got, wantSwitch)
	}

	quiet := Notice{CurrentVersion: "2.17.0", LatestVersion: "2.17.0", Available: false}
	if got := quiet.CLILine(); got != "" {
		t.Errorf("current CLILine should be empty, got %q", got)
	}
}

func TestNotice_UpdateCommand(t *testing.T) {
	if got := (Notice{}).UpdateCommand(); got != "humblskills upgrade" {
		t.Errorf("default UpdateCommand = %q", got)
	}
	if got := (Notice{Homebrew: true, Formula: FormulaPre, CurrentFormula: FormulaPre}).UpdateCommand(); got != "brew upgrade humblskills-pre" {
		t.Errorf("beta brew UpdateCommand = %q", got)
	}
	if got := (Notice{Homebrew: true, Formula: FormulaStable, CurrentFormula: FormulaPre}).UpdateCommand(); got != "brew uninstall humblskills-pre && brew install humblskills" {
		t.Errorf("switch UpdateCommand = %q", got)
	}
}

func TestCheck_DefaultStableHitsLatest(t *testing.T) {
	hits := newReleaseCounter(t, "2.17.0", "", false)
	n := Check(CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.15.0",
		ExePath:        "/usr/local/bin/humblskills",
		CacheDir:       t.TempDir(),
	})
	if !n.Available {
		t.Fatalf("expected available, got %+v", n)
	}
	if n.Channel != ChannelStable {
		t.Errorf("Channel = %q, want stable", n.Channel)
	}
	if n.LatestVersion != "2.17.0" {
		t.Errorf("LatestVersion = %q", n.LatestVersion)
	}
	if n.Homebrew {
		t.Error("plain /usr/local path should not be Homebrew")
	}
	if n.CLILine() != "newer version available: v2.15.0 → v2.17.0 (stable) — run `humblskills upgrade`" {
		t.Errorf("CLILine = %q", n.CLILine())
	}
	if hits.latest.Load() != 1 {
		t.Errorf("/releases/latest hits = %d, want 1", hits.latest.Load())
	}
	if hits.list.Load() != 0 {
		t.Errorf("stable must not list releases, got %d", hits.list.Load())
	}
}

func TestCheck_BetaPicksStableWhenNewerThanPre(t *testing.T) {
	// Jennings dry-run: 2.52.0 > 2.52.0-pre.1, so beta must recommend the
	// stable and tell a brew-pre user to switch formulas.
	hits := newChannelReleases(t, "2.52.0", "2.52.0-pre.1")
	n := Check(CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.52.0-pre",
		Channel:        ChannelBeta,
		ExePath:        "/opt/homebrew/Cellar/humblskills-pre/2.52.0-pre/bin/humblskills",
		CacheDir:       t.TempDir(),
	})
	if !n.Available {
		t.Fatalf("expected available, got %+v", n)
	}
	if n.LatestVersion != "2.52.0" {
		t.Errorf("LatestVersion = %q, want 2.52.0", n.LatestVersion)
	}
	if n.Formula != FormulaStable {
		t.Errorf("Formula = %q, want %s", n.Formula, FormulaStable)
	}
	if n.CurrentFormula != FormulaPre {
		t.Errorf("CurrentFormula = %q, want %s", n.CurrentFormula, FormulaPre)
	}
	want := "newer version available: v2.52.0-pre → v2.52.0 (beta) — run `brew uninstall humblskills-pre && brew install humblskills`"
	if n.CLILine() != want {
		t.Errorf("CLILine = %q, want %q", n.CLILine(), want)
	}
	if hits.latest.Load() != 1 || hits.list.Load() != 1 {
		t.Errorf("beta must hit both endpoints, latest=%d list=%d", hits.latest.Load(), hits.list.Load())
	}
}

func TestCheck_BetaPicksPreWhenNewerThanStable(t *testing.T) {
	hits := newChannelReleases(t, "2.52.0", "2.53.0-pre.1")
	n := Check(CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.52.0",
		Channel:        ChannelBeta,
		ExePath:        "/opt/homebrew/Cellar/humblskills-pre/2.52.0/bin/humblskills",
		CacheDir:       t.TempDir(),
	})
	if n.LatestVersion != "2.53.0-pre.1" {
		t.Errorf("LatestVersion = %q, want 2.53.0-pre.1", n.LatestVersion)
	}
	if n.Formula != FormulaPre {
		t.Errorf("Formula = %q, want %s", n.Formula, FormulaPre)
	}
	if n.UpdateCommand() != "brew upgrade humblskills-pre" {
		t.Errorf("UpdateCommand = %q", n.UpdateCommand())
	}
}

func TestCheck_StableNeverPicksPre(t *testing.T) {
	hits := newChannelReleases(t, "2.52.0", "2.53.0-pre.1")
	n := Check(CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.51.0",
		Channel:        ChannelStable,
		CacheDir:       t.TempDir(),
	})
	if n.LatestVersion != "2.52.0" {
		t.Errorf("LatestVersion = %q, want 2.52.0", n.LatestVersion)
	}
	if hits.list.Load() != 0 {
		t.Errorf("stable must not list releases, got %d", hits.list.Load())
	}
}

func TestCheck_CurrentIsQuiet(t *testing.T) {
	hits := newReleaseCounter(t, "2.17.0", "", false)
	n := Check(CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.17.0",
		CacheDir:       t.TempDir(),
	})
	if n.Available {
		t.Errorf("current should be quiet, got %+v", n)
	}
	if n.CLILine() != "" {
		t.Errorf("CLILine should be empty, got %q", n.CLILine())
	}
	if n.LatestVersion != "2.17.0" {
		t.Errorf("LatestVersion still recorded = %q", n.LatestVersion)
	}
}

func TestCheck_CacheSkipsRefetch(t *testing.T) {
	hits := newReleaseCounter(t, "2.17.0", "", false)
	dir := t.TempDir()
	opts := CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.15.0",
		CacheDir:       dir,
	}
	first := Check(opts)
	if !first.Available {
		t.Fatalf("first check: %+v", first)
	}
	second := Check(opts)
	if !second.Available || second.LatestVersion != "2.17.0" {
		t.Fatalf("second check: %+v", second)
	}
	if hits.latest.Load() != 1 {
		t.Errorf("expected one fetch, got %d", hits.latest.Load())
	}
}

func TestCheck_DiskCacheSurvivesProcessMemory(t *testing.T) {
	hits := newReleaseCounter(t, "2.17.0", "", false)
	dir := t.TempDir()
	opts := CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.15.0",
		CacheDir:       dir,
	}
	if n := Check(opts); !n.Available {
		t.Fatalf("warm: %+v", n)
	}
	InvalidateNoticeCache(dir)
	if n := Check(opts); !n.Available {
		t.Fatalf("rewarm: %+v", n)
	}
	if hits.latest.Load() != 2 {
		t.Fatalf("rewarm hits = %d, want 2", hits.latest.Load())
	}
	noticeMemMu.Lock()
	delete(noticeMem, noticeMemKey{cacheDir: dir, channel: ChannelStable})
	noticeMemMu.Unlock()
	if n := Check(opts); !n.Available {
		t.Fatalf("disk hit: %+v", n)
	}
	if hits.latest.Load() != 2 {
		t.Errorf("disk cache should not refetch, hits = %d", hits.latest.Load())
	}
}

func TestCheck_TTLExpiryRefetches(t *testing.T) {
	hits := newReleaseCounter(t, "2.17.0", "", false)
	dir := t.TempDir()
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	opts := CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.15.0",
		CacheDir:       dir,
		TTL:            time.Hour,
		Now:            func() time.Time { return now },
	}
	if n := Check(opts); !n.Available {
		t.Fatalf("first: %+v", n)
	}
	now = now.Add(2 * time.Hour)
	if n := Check(opts); !n.Available {
		t.Fatalf("expired: %+v", n)
	}
	if hits.latest.Load() != 2 {
		t.Errorf("expired TTL should refetch, hits = %d", hits.latest.Load())
	}
}

func TestCheck_FetchFailureStaysQuiet(t *testing.T) {
	prev := GitHubAPIBase
	GitHubAPIBase = "http://127.0.0.1:0"
	t.Cleanup(func() { GitHubAPIBase = prev })

	n := Check(CheckOptions{
		Client:         &http.Client{Timeout: 50 * time.Millisecond},
		CurrentVersion: "2.15.0",
		CacheDir:       t.TempDir(),
	})
	if n.Available || n.CLILine() != "" {
		t.Errorf("failed check must stay quiet, got %+v %q", n, n.CLILine())
	}
}

func TestCheck_HomebrewFormula(t *testing.T) {
	hits := newReleaseCounter(t, "2.17.0", "", false)
	n := Check(CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.15.0",
		ExePath:        "/opt/homebrew/Cellar/humblskills/2.15.0/bin/humblskills",
		CacheDir:       t.TempDir(),
	})
	if !n.Homebrew {
		t.Fatal("expected Homebrew detection")
	}
	if n.Formula != FormulaStable {
		t.Errorf("Formula = %q", n.Formula)
	}
	if n.UpdateCommand() != "brew upgrade humblskills" {
		t.Errorf("UpdateCommand = %q", n.UpdateCommand())
	}
}

func TestCheck_UnknownChannelDefaultsStable(t *testing.T) {
	hits := newReleaseCounter(t, "2.17.0", "", false)
	n := Check(CheckOptions{
		Client:         hits.client,
		CurrentVersion: "2.15.0",
		Channel:        "nightly",
		CacheDir:       t.TempDir(),
	})
	if n.Channel != ChannelStable {
		t.Errorf("Channel = %q, want stable", n.Channel)
	}
	if hits.latest.Load() != 1 {
		t.Error("unknown channel should hit /releases/latest")
	}
}

type releaseHits struct {
	client *http.Client
	latest atomic.Int32
	list   atomic.Int32
}

func newReleaseCounter(t *testing.T, latest, pre string, prerelease bool) *releaseHits {
	t.Helper()
	h := &releaseHits{}
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/jjfantini/humblSKILLS/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		h.latest.Add(1)
		_, _ = w.Write([]byte(`{"tag_name": "v` + latest + `", "prerelease": false}`))
	})
	mux.HandleFunc("/repos/jjfantini/humblSKILLS/releases", func(w http.ResponseWriter, r *http.Request) {
		h.list.Add(1)
		tag := latest
		flag := "false"
		if prerelease {
			flag = "true"
		}
		if pre != "" {
			tag = pre
			flag = "true"
		}
		_, _ = w.Write([]byte(`[{"tag_name": "v` + tag + `", "prerelease": ` + flag + `, "draft": false}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	prev := GitHubAPIBase
	GitHubAPIBase = srv.URL
	t.Cleanup(func() { GitHubAPIBase = prev })
	h.client = srv.Client()
	return h
}

func newChannelReleases(t *testing.T, stable, pre string) *releaseHits {
	t.Helper()
	h := &releaseHits{}
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/jjfantini/humblSKILLS/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		h.latest.Add(1)
		if stable == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"tag_name": "v` + stable + `", "prerelease": false}`))
	})
	mux.HandleFunc("/repos/jjfantini/humblSKILLS/releases", func(w http.ResponseWriter, r *http.Request) {
		h.list.Add(1)
		body := `[`
		if stable != "" {
			body += `{"tag_name": "v` + stable + `", "prerelease": false, "draft": false}`
		}
		if pre != "" {
			if stable != "" {
				body += `,`
			}
			body += `{"tag_name": "v` + pre + `", "prerelease": true, "draft": false}`
		}
		body += `]`
		_, _ = w.Write([]byte(body))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	prev := GitHubAPIBase
	GitHubAPIBase = srv.URL
	t.Cleanup(func() { GitHubAPIBase = prev })
	h.client = srv.Client()
	return h
}

func TestNoticeCachePath(t *testing.T) {
	got := noticeCachePath("/tmp/cache", ChannelBeta)
	want := filepath.Join("/tmp/cache", "selfupdate", "latest-beta.json")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
}
