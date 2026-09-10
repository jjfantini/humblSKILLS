package selfupdate

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		current, latest string
		wantSign        int // -1, 0, +1 (sign only, magnitude unspecified)
	}{
		{"2.15.0", "2.17.0", -1},
		{"v2.15.0", "2.17.0", -1}, // already-prefixed current
		{"2.17.0", "2.17.0", 0},
		{"2.17.0", "v2.17.0", 0},
		{"2.18.0", "2.17.0", +1}, // local build ahead of latest release
		{"dev", "2.17.0", -1},    // unparseable current always upgradable
		{"", "2.17.0", -1},
		{"2.52.0-pre.1", "2.52.0", -1}, // graduated stable beats its own pre
		{"2.52.0-pre.3", "2.52.0", -1}, // stuck pre line: 2.52.0-pre.3 lost to 2.52.0
		{"2.52.0", "2.52.1-pre.1", -1}, // patch pre beats last stable (beta ship)
		{"2.52.0", "2.53.0-pre.1", -1}, // newer minor pre also beats last stable
		{"2.52.0", "2.52.0-pre.1", +1},
	}
	for _, c := range cases {
		got := Compare(c.current, c.latest)
		gotSign := sign(got)
		if gotSign != c.wantSign {
			t.Errorf("Compare(%q, %q) = %d (sign %d), want sign %d", c.current, c.latest, got, gotSign, c.wantSign)
		}
	}
}

func TestIsUpgradeAvailable(t *testing.T) {
	if !IsUpgradeAvailable("2.15.0", "2.17.0") {
		t.Error("expected upgrade available 2.15.0 -> 2.17.0")
	}
	if IsUpgradeAvailable("2.17.0", "2.17.0") {
		t.Error("expected no upgrade when versions are equal")
	}
	if IsUpgradeAvailable("2.18.0", "2.17.0") {
		t.Error("expected no upgrade when current is newer than latest")
	}
	if !IsUpgradeAvailable("dev", "2.17.0") {
		t.Error("expected dev build to always be upgradable")
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"2.17.0":  "v2.17.0",
		"v2.17.0": "v2.17.0",
		"":        "",
		"dev":     "vdev",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
