package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/selfupdate"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/testutil"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/ui"
)

func withNoticeNetwork(t *testing.T) {
	t.Helper()
	prev := selfupdate.SkipNetwork
	selfupdate.SkipNetwork = false
	t.Cleanup(func() { selfupdate.SkipNetwork = prev })
}

func TestNotice_VersionStderr_StableBehind(t *testing.T) {
	s := testutil.NewSandbox(t)
	withNoticeNetwork(t)
	startFakeReleaseAPI(t, "99.0.0", "fake")

	res := runCLIWithStdoutCapture(t, "version", "--profile", s.ProfilePath, "--cache-dir", s.CacheDir)
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	if !strings.Contains(res.Err, "newer version available") {
		t.Errorf("stderr missing notice:\n%s", res.Err)
	}
	if !strings.Contains(res.Err, "humblskills upgrade") {
		t.Errorf("stderr missing upgrade command:\n%s", res.Err)
	}
	if strings.Contains(res.Out, "newer version available") {
		t.Error("notice must go to stderr, not stdout")
	}
}

func TestNotice_JSONStaysQuiet(t *testing.T) {
	s := testutil.NewSandbox(t)
	withNoticeNetwork(t)
	startFakeReleaseAPI(t, "99.0.0", "fake")

	res := runCLIWithStdoutCapture(t, "version", "--json", "--profile", s.ProfilePath, "--cache-dir", s.CacheDir)
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	if strings.Contains(res.Err, "newer version available") {
		t.Errorf("--json must stay quiet, stderr:\n%s", res.Err)
	}
	var info versionInfo
	if err := json.Unmarshal([]byte(strings.TrimSpace(res.Out)), &info); err != nil {
		t.Fatalf("version --json: %v", err)
	}
}

func TestNotice_ChannelFlag_BetaPicksStableWinner(t *testing.T) {
	s := testutil.NewSandbox(t)
	withNoticeNetwork(t)
	startFakeChannelAPI(t, "2.52.0", "2.52.0-pre.1")
	exe := filepath.Join(s.Root, "Cellar", "humblskills-pre", "2.52.0-pre.1", "bin", "humblskills")
	if err := os.MkdirAll(filepath.Dir(exe), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	withFakeExecutable(t, exe)

	res := runCLIWithStdoutCapture(t,
		"version", "--channel", "beta",
		"--profile", s.ProfilePath, "--cache-dir", s.CacheDir,
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	if !strings.Contains(res.Err, "v2.52.0") {
		t.Errorf("beta notice should pick stable 2.52.0:\n%s", res.Err)
	}
	// Homebrew detection looks for "/Cellar/" — filepath.Join uses backslash
	// on Windows, so that install is a GitHub binary there. Assert the
	// command the same helper production uses, not a brew string on Windows.
	want := selfupdate.RecommendedUpgradeCommand(
		selfupdate.IsHomebrewManaged(exe),
		selfupdate.InstalledFormula(exe),
		selfupdate.FormulaForVersion("2.52.0"),
	)
	if !strings.Contains(res.Err, want) {
		t.Errorf("notice missing %q:\n%s", want, res.Err)
	}
	if strings.Contains(res.Err, "brew upgrade humblskills-pre") {
		t.Errorf("must not recommend brew upgrade pre when stable won:\n%s", res.Err)
	}
}

func TestNotice_ProfileChannel_BetaPicksNewerPre(t *testing.T) {
	s := testutil.NewSandbox(t)
	if err := profile.Save(s.ProfilePath, &profile.Profile{Channel: profile.ChannelBeta}); err != nil {
		t.Fatal(err)
	}
	withNoticeNetwork(t)
	startFakeChannelAPI(t, "2.52.0", "2.53.0-pre.1")

	res := runCLIWithStdoutCapture(t,
		"version", "--profile", s.ProfilePath, "--cache-dir", s.CacheDir,
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	if !strings.Contains(res.Err, "2.53.0-pre.1") {
		t.Errorf("profile beta should pick newer pre:\n%s", res.Err)
	}
}

func TestNotice_DoctorUsesSameResolver(t *testing.T) {
	s := testutil.NewSandbox(t)
	withNoticeNetwork(t)
	startFakeReleaseAPI(t, "99.0.0", "fake")

	res := runCLIWithStdoutCapture(t, "doctor", "--yes", "--profile", s.ProfilePath, "--cache-dir", s.CacheDir)
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	if !strings.Contains(res.Err, "newer version available") {
		t.Errorf("doctor stderr missing notice:\n%s", res.Err)
	}
}

func TestTuiVersionNotice_UsesResolver(t *testing.T) {
	s := testutil.NewSandbox(t)
	withNoticeNetwork(t)
	startFakeChannelAPI(t, "2.52.0", "2.52.0-pre.1")
	exe := filepath.Join(s.Root, "Cellar", "humblskills-pre", "2.52.0-pre.1", "bin", "humblskills")
	if err := os.MkdirAll(filepath.Dir(exe), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	withFakeExecutable(t, exe)

	app := &App{
		UI: ui.New(ui.Options{}),
		Config: Config{
			ProfilePath: s.ProfilePath,
			CacheDir:    s.CacheDir,
			Channel:     profile.ChannelBeta,
		},
	}
	n := app.tuiVersionNotice()
	if n == nil {
		t.Fatal("expected dashboard notice")
	}
	if n.Latest != "v2.52.0" {
		t.Errorf("Latest = %q, want v2.52.0", n.Latest)
	}
	if n.Channel != profile.ChannelBeta {
		t.Errorf("Channel = %q, want beta", n.Channel)
	}
	// filepath.Join(s.Root, "Cellar", ...) is Homebrew on Unix ("/Cellar/")
	// and a plain GitHub install on Windows ("\Cellar\"). Same helper
	// upgrade/notices use — don't assert brew commands where brew isn't.
	want := selfupdate.RecommendedUpgradeCommand(
		selfupdate.IsHomebrewManaged(exe),
		selfupdate.InstalledFormula(exe),
		selfupdate.FormulaForVersion("2.52.0"),
	)
	if n.Command != want {
		t.Errorf("Command = %q, want %q", n.Command, want)
	}
}

func TestNotice_SkipNetworkKeepsQuiet(t *testing.T) {
	s := testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t, "version", "--profile", s.ProfilePath)
	if res.RunErr != nil {
		t.Fatalf("run: %v", res.RunErr)
	}
	if strings.Contains(res.Err, "newer version available") {
		t.Errorf("SkipNetwork should keep version quiet:\n%s", res.Err)
	}
}
