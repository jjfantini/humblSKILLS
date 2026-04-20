package main

import (
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestRoot_QuietAndVerboseMutuallyExclusive(t *testing.T) {
	testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"version",
		"--quiet", "--verbose",
	)
	if res.RunErr == nil {
		t.Fatal("expected error when --quiet and --verbose both set")
	}
	if !strings.Contains(res.RunErr.Error(), "mutually exclusive") {
		t.Errorf("err = %v", res.RunErr)
	}
}

func TestRoot_Help_IncludesCommandList(t *testing.T) {
	testutil.NewSandbox(t)

	res := runCLI(t, "--help")
	if res.RunErr != nil {
		t.Fatalf("help returned error: %v", res.RunErr)
	}
	// Cobra --help output goes through the Cobra-supplied writer
	// (captured by runCLI).
	for _, want := range []string{"install", "update", "list", "search", "doctor", "profile"} {
		if !strings.Contains(res.Out, want) {
			t.Errorf("help missing %q:\n%s", want, res.Out)
		}
	}
}

func TestRoot_UnknownCommandErrors(t *testing.T) {
	testutil.NewSandbox(t)
	res := runCLI(t, "not-a-command")
	if res.RunErr == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "", "c", "d"); got != "c" {
		t.Errorf("got %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Errorf("got %q", got)
	}
	if got := firstNonEmpty("a"); got != "a" {
		t.Errorf("got %q", got)
	}
}

func TestResolveCacheDir_Override(t *testing.T) {
	got, err := resolveCacheDir("/tmp/override")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/override" {
		t.Errorf("override not honoured: %q", got)
	}
}

func TestResolveCacheDir_UsesXDG(t *testing.T) {
	s := testutil.NewSandbox(t)
	got, err := resolveCacheDir("")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, s.XDGCacheHome) {
		t.Errorf("default cache dir %q not under XDG_CACHE_HOME %q", got, s.XDGCacheHome)
	}
}

func TestResolveManifestPath_Override(t *testing.T) {
	got, err := resolveManifestPath("/tmp/override.json")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/override.json" {
		t.Errorf("got %q", got)
	}
}

func TestResolveProfilePath_Override(t *testing.T) {
	got, err := resolveProfilePath("/tmp/profile.json")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/profile.json" {
		t.Errorf("got %q", got)
	}
}

func TestConfigureApp_SetsCacheManifestProfileFromSandbox(t *testing.T) {
	s := testutil.NewSandbox(t)

	// Run any real command under the sandbox to trigger configureApp,
	// then verify the install happened under the sandboxed paths.
	enableClaudeCode(t, s)
	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "foo", Version: "1.0.0", Platforms: []string{"claude-code"},
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})
	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--platform", "claude-code", "--scope", "user",
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("install: %v\n%s", res.RunErr, res.Err)
	}
	// Manifest should have been created under the XDG sandbox,
	// not the developer's real ~/.local/state.
	if !strings.HasPrefix(s.ManifestPath, s.XDGStateHome) {
		t.Fatalf("sandbox manifest path %q not under XDG_STATE_HOME %q",
			s.ManifestPath, s.XDGStateHome)
	}
}
