package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/selfupdate"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/testutil"
)

func enableNoticeNetwork(t *testing.T) {
	t.Helper()
	prev := selfupdate.SkipNetwork
	selfupdate.SkipNetwork = false
	t.Cleanup(func() { selfupdate.SkipNetwork = prev })
}

func withVersion(t *testing.T, v string) {
	t.Helper()
	prev := version
	version = v
	t.Cleanup(func() { version = prev })
}

type noticeAPI struct {
	latest atomic.Int32
	list   atomic.Int32
}

func startNoticeAPI(t *testing.T, stableTag, preTag string) *noticeAPI {
	t.Helper()
	api := &noticeAPI{}
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/"+selfupdate.DefaultRepo+"/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		api.latest.Add(1)
		_, _ = w.Write([]byte(`{"tag_name": "` + stableTag + `", "prerelease": false}`))
	})
	mux.HandleFunc("/repos/"+selfupdate.DefaultRepo+"/releases", func(w http.ResponseWriter, r *http.Request) {
		api.list.Add(1)
		tag := preTag
		if tag == "" {
			tag = "v9.9.9-pre.1"
		}
		_, _ = w.Write([]byte(`[{"tag_name": "` + tag + `", "prerelease": true, "draft": false}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	prev := selfupdate.GitHubAPIBase
	selfupdate.GitHubAPIBase = srv.URL
	t.Cleanup(func() { selfupdate.GitHubAPIBase = prev })
	return api
}

func TestDoctor_UpgradeNotice_DefaultStableOutdated(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.15.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	api := startNoticeAPI(t, "v2.17.0", "")

	res := runCLIWithStdoutCapture(t, "doctor", "--yes",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)

	if res.RunErr != nil && !strings.Contains(res.RunErr.Error(), "doctor") {
		t.Fatalf("doctor: %v\n%s", res.RunErr, res.Err)
	}
	want := "newer version available: v2.15.0 → v2.17.0 (stable) — run `humblskills upgrade`"
	assertContains(t, res.Err, want)
	assertNotContains(t, res.Out, "newer version available")
	if api.latest.Load() != 1 {
		t.Errorf("/releases/latest hits = %d, want 1", api.latest.Load())
	}
	if api.list.Load() != 0 {
		t.Error("default stable must not list prereleases")
	}
}

func TestDoctor_UpgradeNotice_BetaChannel(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.51.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	if err := profile.Save(s.ProfilePath, &profile.Profile{Channel: profile.ChannelBeta}); err != nil {
		t.Fatal(err)
	}
	api := startNoticeAPI(t, "v2.51.0", "v2.52.0-pre.1")

	res := runCLIWithStdoutCapture(t, "doctor", "--yes",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)

	if res.RunErr != nil && !strings.Contains(res.RunErr.Error(), "doctor") {
		t.Fatalf("doctor: %v\n%s", res.RunErr, res.Err)
	}
	want := "newer version available: v2.51.0 → v2.52.0-pre.1 (beta) — run `humblskills upgrade`"
	assertContains(t, res.Err, want)
	if api.latest.Load() != 0 {
		t.Error("beta must not hit /releases/latest")
	}
	if api.list.Load() != 1 {
		t.Errorf("releases list hits = %d, want 1", api.list.Load())
	}
}

func TestDoctor_UpgradeNotice_ChannelFlagOverridesProfile(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.15.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	if err := profile.Save(s.ProfilePath, &profile.Profile{Channel: profile.ChannelBeta}); err != nil {
		t.Fatal(err)
	}
	api := startNoticeAPI(t, "v2.17.0", "v2.52.0-pre.1")

	res := runCLIWithStdoutCapture(t, "doctor", "--yes", "--channel", "stable",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)

	if res.RunErr != nil && !strings.Contains(res.RunErr.Error(), "doctor") {
		t.Fatalf("doctor: %v\n%s", res.RunErr, res.Err)
	}
	assertContains(t, res.Err, "v2.15.0 → v2.17.0 (stable)")
	assertNotContains(t, res.Err, "2.52.0-pre.1")
	if api.latest.Load() != 1 {
		t.Errorf("--channel stable should hit /releases/latest, got %d", api.latest.Load())
	}
	if api.list.Load() != 0 {
		t.Error("--channel stable must not list prereleases")
	}
}

func TestDoctor_UpgradeNotice_CurrentIsQuiet(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.17.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	startNoticeAPI(t, "v2.17.0", "")

	res := runCLIWithStdoutCapture(t, "doctor", "--yes",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)

	if res.RunErr != nil && !strings.Contains(res.RunErr.Error(), "doctor") {
		t.Fatalf("doctor: %v\n%s", res.RunErr, res.Err)
	}
	assertNotContains(t, res.Err, "newer version available")
	assertNotContains(t, res.Out, "newer version available")
}

func TestDoctor_UpgradeNotice_JSONStaysMachineReadable(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.15.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	startNoticeAPI(t, "v2.17.0", "")

	res := runCLIWithStdoutCapture(t, "doctor", "--json",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)

	out := strings.TrimSpace(res.Out)
	if !strings.HasPrefix(out, "{") {
		t.Fatalf("expected JSON object on stdout, got:\n%s", res.Out)
	}
	assertNotContains(t, res.Out, "newer version available")
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, res.Out)
	}
	// --json skips the notice entirely (no banner on stderr either).
	assertNotContains(t, res.Err, "newer version available")
}

func TestDoctor_UpgradeNotice_CacheDoesNotRefetch(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.15.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	api := startNoticeAPI(t, "v2.17.0", "")

	args := []string{"doctor", "--yes", "--cache-dir", s.CacheDir, "--profile", s.ProfilePath}
	first := runCLIWithStdoutCapture(t, args...)
	assertContains(t, first.Err, "newer version available: v2.15.0 → v2.17.0 (stable)")
	second := runCLIWithStdoutCapture(t, args...)
	assertContains(t, second.Err, "newer version available: v2.15.0 → v2.17.0 (stable)")
	if api.latest.Load() != 1 {
		t.Errorf("expected one GitHub fetch across two doctor runs, got %d", api.latest.Load())
	}
}

func TestDoctor_UpgradeNotice_HomebrewFormula(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.15.0")
	withFakeExecutable(t, "/opt/homebrew/Cellar/humblskills/2.15.0/bin/humblskills")
	startNoticeAPI(t, "v2.17.0", "")

	res := runCLIWithStdoutCapture(t, "doctor", "--yes",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)

	assertContains(t, res.Err, "newer version available: v2.15.0 → v2.17.0 (stable) — run `brew upgrade humblskills`")
}

func TestStart_UpgradeNotice_OnFallback(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.15.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	startNoticeAPI(t, "v2.17.0", "")

	res := runCLIWithStdoutCapture(t, "start",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)

	if res.RunErr != nil {
		t.Fatalf("start: %v", res.RunErr)
	}
	assertContains(t, res.Out, "COMMANDS")
	assertContains(t, res.Err, "newer version available: v2.15.0 → v2.17.0 (stable) — run `humblskills upgrade`")
	assertNotContains(t, res.Out, "newer version available")
}

func TestVersion_UpgradeNotice_JSONStaysClean(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.15.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	startNoticeAPI(t, "v2.17.0", "")

	res := runCLIWithStdoutCapture(t, "version", "--json",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)
	if res.RunErr != nil {
		t.Fatalf("version --json: %v", res.RunErr)
	}
	assertNotContains(t, res.Out, "newer version available")
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(res.Out)), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, res.Out)
	}
}

func TestVersion_UpgradeNotice_TextShowsOnStderr(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableNoticeNetwork(t)
	withVersion(t, "2.15.0")
	withFakeExecutable(t, "/usr/local/bin/humblskills")
	startNoticeAPI(t, "v2.17.0", "")

	res := runCLIWithStdoutCapture(t, "version", "--yes",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)
	if res.RunErr != nil {
		t.Fatalf("version: %v", res.RunErr)
	}
	assertContains(t, res.Out, "humblskills")
	assertContains(t, res.Err, "newer version available: v2.15.0 → v2.17.0 (stable)")
}

func TestChannelFlag_InvalidRejected(t *testing.T) {
	s := testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t, "doctor", "--channel", "nightly",
		"--cache-dir", s.CacheDir, "--profile", s.ProfilePath)
	if res.RunErr == nil {
		t.Fatal("expected invalid --channel to error")
	}
	assertContains(t, res.RunErr.Error(), "invalid --channel")
}
