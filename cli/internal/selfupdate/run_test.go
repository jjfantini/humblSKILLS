package selfupdate

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fakeReleaseServer serves a fake GitHub "releases/latest" response (with
// asset URLs pointing back at itself) plus the asset bytes themselves, so
// ResolvePlan/Apply can be exercised end-to-end without touching the real
// network. archiveContent is what the (fake) binary's bytes are inside the
// tar.gz/zip asset.
func fakeReleaseServer(t *testing.T, latestVersion, archiveContent string) (*httptest.Server, string) {
	t.Helper()

	assetName, err := CurrentAssetName(latestVersion)
	if err != nil {
		t.Fatalf("CurrentAssetName: %v", err)
	}
	binName := BinaryName(runtime.GOOS)

	var archivePath string
	if runtime.GOOS == "windows" {
		archivePath = buildZip(t, map[string]string{binName: archiveContent})
	} else {
		archivePath = buildTarGz(t, map[string]string{binName: archiveContent})
	}
	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	sum, err := sha256File(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	checksums := sum + "  " + assetName + "\n"

	// srv.URL needs to be embedded in the release JSON, but srv doesn't
	// exist until after httptest.NewServer returns. The handler closure
	// below reads srv at *request* time (after the assignment a few lines
	// down), not at registration time, so this works.
	var srv *httptest.Server

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/jjfantini/humblSKILLS/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v` + latestVersion + `",
			"assets": [
				{"name": "` + assetName + `", "browser_download_url": "` + srv.URL + `/assets/` + assetName + `"},
				{"name": "checksums.txt", "browser_download_url": "` + srv.URL + `/assets/checksums.txt"}
			]
		}`))
	})
	mux.HandleFunc("/assets/"+assetName, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveBytes)
	})
	mux.HandleFunc("/assets/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(checksums))
	})

	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv, assetName
}

func TestResolvePlan_UpgradeAvailable(t *testing.T) {
	srv, assetName := fakeReleaseServer(t, "2.17.0", "fake binary v2.17.0")
	prev := GitHubAPIBase
	GitHubAPIBase = srv.URL
	t.Cleanup(func() { GitHubAPIBase = prev })

	var events []Event
	plan, err := ResolvePlan(srv.Client(), DefaultRepo, "2.15.0", "/usr/local/bin/humblskills", func(ev Event) {
		events = append(events, ev)
	})
	if err != nil {
		t.Fatalf("ResolvePlan: %v", err)
	}
	if !plan.UpgradeAvailable {
		t.Error("expected UpgradeAvailable = true")
	}
	if plan.LatestVersion != "2.17.0" {
		t.Errorf("LatestVersion = %q, want 2.17.0", plan.LatestVersion)
	}
	if plan.Homebrew {
		t.Error("expected Homebrew = false for /usr/local/bin path")
	}
	if plan.AssetName != assetName {
		t.Errorf("AssetName = %q, want %q", plan.AssetName, assetName)
	}
	if len(events) == 0 || events[0].Phase != PhaseCheckingLatest {
		t.Errorf("expected first event to be PhaseCheckingLatest, got %v", events)
	}
}

func TestResolvePlan_AlreadyUpToDate(t *testing.T) {
	srv, _ := fakeReleaseServer(t, "2.17.0", "fake binary v2.17.0")
	prev := GitHubAPIBase
	GitHubAPIBase = srv.URL
	t.Cleanup(func() { GitHubAPIBase = prev })

	plan, err := ResolvePlan(srv.Client(), DefaultRepo, "2.17.0", "/usr/local/bin/humblskills", nil)
	if err != nil {
		t.Fatalf("ResolvePlan: %v", err)
	}
	if plan.UpgradeAvailable {
		t.Error("expected UpgradeAvailable = false when current == latest")
	}
	// No download metadata should be populated when nothing's upgrading.
	if plan.AssetName != "" {
		t.Errorf("expected empty AssetName when up to date, got %q", plan.AssetName)
	}
}

func TestResolvePlan_HomebrewDetected(t *testing.T) {
	srv, _ := fakeReleaseServer(t, "2.17.0", "fake binary v2.17.0")
	prev := GitHubAPIBase
	GitHubAPIBase = srv.URL
	t.Cleanup(func() { GitHubAPIBase = prev })

	plan, err := ResolvePlan(srv.Client(), DefaultRepo, "2.15.0", "/opt/homebrew/Cellar/humblskills/2.15.0/bin/humblskills", nil)
	if err != nil {
		t.Fatalf("ResolvePlan: %v", err)
	}
	if !plan.Homebrew {
		t.Error("expected Homebrew = true for a Cellar path")
	}
}

func TestApply_DownloadsVerifiesAndSwaps(t *testing.T) {
	const newContent = "fake binary v2.17.0"
	srv, _ := fakeReleaseServer(t, "2.17.0", newContent)
	prev := GitHubAPIBase
	GitHubAPIBase = srv.URL
	t.Cleanup(func() { GitHubAPIBase = prev })

	dir := t.TempDir()
	exePath := filepath.Join(dir, "humblskills")
	if err := os.WriteFile(exePath, []byte("old binary v2.15.0"), 0o755); err != nil {
		t.Fatal(err)
	}

	plan, err := ResolvePlan(srv.Client(), DefaultRepo, "2.15.0", exePath, nil)
	if err != nil {
		t.Fatalf("ResolvePlan: %v", err)
	}

	var events []Event
	cacheDir := filepath.Join(dir, "cache")
	if err := Apply(srv.Client(), plan, cacheDir, exePath, func(ev Event) { events = append(events, ev) }); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != newContent {
		t.Errorf("exePath content after Apply = %q, want %q", got, newContent)
	}

	wantPhases := []Phase{PhaseDownloading, PhaseVerifyingSum, PhaseInstalling}
	if len(events) != len(wantPhases) {
		t.Fatalf("got %d events, want %d: %v", len(events), len(wantPhases), events)
	}
	for i, ev := range events {
		if ev.Phase != wantPhases[i] {
			t.Errorf("event[%d].Phase = %q, want %q", i, ev.Phase, wantPhases[i])
		}
		if ev.Err != nil {
			t.Errorf("event[%d] unexpected error: %v", i, ev.Err)
		}
	}
}

func TestApply_ChecksumMismatch(t *testing.T) {
	srv, assetName := fakeReleaseServer(t, "2.17.0", "fake binary v2.17.0")
	prev := GitHubAPIBase
	GitHubAPIBase = srv.URL
	t.Cleanup(func() { GitHubAPIBase = prev })

	dir := t.TempDir()
	exePath := filepath.Join(dir, "humblskills")
	if err := os.WriteFile(exePath, []byte("old binary v2.15.0"), 0o755); err != nil {
		t.Fatal(err)
	}

	plan, err := ResolvePlan(srv.Client(), DefaultRepo, "2.15.0", exePath, nil)
	if err != nil {
		t.Fatalf("ResolvePlan: %v", err)
	}
	// Corrupt the expected asset name so VerifyChecksum can't find a
	// matching entry in checksums.txt — simulates a checksum mismatch
	// without needing a second server route.
	plan.AssetName = assetName + ".tampered"

	var sawError bool
	cacheDir := filepath.Join(dir, "cache")
	err = Apply(srv.Client(), plan, cacheDir, exePath, func(ev Event) {
		if ev.Phase == PhaseError {
			sawError = true
		}
	})
	if err == nil {
		t.Fatal("expected Apply to fail on checksum mismatch")
	}
	if !sawError {
		t.Error("expected a PhaseError event to be emitted")
	}
	// exePath must be untouched since checksum verification fails before
	// the swap ever runs.
	got, _ := os.ReadFile(exePath)
	if string(got) != "old binary v2.15.0" {
		t.Errorf("exePath was modified despite Apply failing: %q", got)
	}
}
