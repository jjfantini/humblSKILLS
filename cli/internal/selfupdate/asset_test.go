package selfupdate

import "testing"

func TestAssetName_MatchesGoreleaserTemplate(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"linux", "amd64", "humblskills_2.17.0_linux_amd64.tar.gz"},
		{"linux", "arm64", "humblskills_2.17.0_linux_arm64.tar.gz"},
		{"darwin", "amd64", "humblskills_2.17.0_macos_amd64.tar.gz"},
		{"darwin", "arm64", "humblskills_2.17.0_macos_arm64.tar.gz"},
		{"windows", "amd64", "humblskills_2.17.0_windows_amd64.zip"},
		{"windows", "arm64", "humblskills_2.17.0_windows_arm64.zip"},
	}
	for _, c := range cases {
		got, err := AssetName("2.17.0", c.goos, c.goarch)
		if err != nil {
			t.Fatalf("AssetName(%q, %q): %v", c.goos, c.goarch, err)
		}
		if got != c.want {
			t.Errorf("AssetName(%q, %q) = %q, want %q", c.goos, c.goarch, got, c.want)
		}
	}
}

func TestAssetName_RejectsUnsupportedPlatforms(t *testing.T) {
	if _, err := AssetName("2.17.0", "plan9", "amd64"); err == nil {
		t.Error("expected error for unsupported OS, got nil")
	}
	if _, err := AssetName("2.17.0", "linux", "riscv64"); err == nil {
		t.Error("expected error for unsupported arch, got nil")
	}
}

func TestBinaryName(t *testing.T) {
	if got := BinaryName("windows"); got != "humblskills.exe" {
		t.Errorf("BinaryName(windows) = %q, want humblskills.exe", got)
	}
	if got := BinaryName("linux"); got != "humblskills" {
		t.Errorf("BinaryName(linux) = %q, want humblskills", got)
	}
	if got := BinaryName("darwin"); got != "humblskills" {
		t.Errorf("BinaryName(darwin) = %q, want humblskills", got)
	}
}
