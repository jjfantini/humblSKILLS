package selfupdate

import (
	"fmt"
	"runtime"
)

// ChecksumsAssetName is the fixed name goreleaser publishes for the
// checksums manifest (.goreleaser.yaml: checksum.name_template).
const ChecksumsAssetName = "checksums.txt"

// archiveOS maps runtime.GOOS to the tag goreleaser's archive
// name_template uses (.goreleaser.yaml renames "darwin" to "macos"; every
// other supported GOOS is used verbatim).
func archiveOS(goos string) (string, error) {
	switch goos {
	case "darwin":
		return "macos", nil
	case "linux", "windows":
		return goos, nil
	default:
		return "", fmt.Errorf("unsupported OS %q for self-upgrade", goos)
	}
}

func archiveArch(goarch string) (string, error) {
	switch goarch {
	case "amd64", "arm64":
		return goarch, nil
	default:
		return "", fmt.Errorf("unsupported architecture %q for self-upgrade", goarch)
	}
}

// archiveExt returns the archive container goreleaser uses for goos:
// windows ships .zip (format_overrides), every other OS ships .tar.gz.
func archiveExt(goos string) string {
	if goos == "windows" {
		return "zip"
	}
	return "tar.gz"
}

// AssetName builds the exact archive filename goreleaser publishes for
// version/goos/goarch, e.g. "humblskills_2.17.0_macos_arm64.tar.gz".
// version must be bare (no leading "v").
func AssetName(version, goos, goarch string) (string, error) {
	osTag, err := archiveOS(goos)
	if err != nil {
		return "", err
	}
	archTag, err := archiveArch(goarch)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("humblskills_%s_%s_%s.%s", version, osTag, archTag, archiveExt(goos)), nil
}

// CurrentAssetName is AssetName for the running process's own GOOS/GOARCH.
func CurrentAssetName(version string) (string, error) {
	return AssetName(version, runtime.GOOS, runtime.GOARCH)
}

// BinaryName returns the binary's filename inside the extracted archive:
// "humblskills.exe" on Windows, "humblskills" everywhere else.
func BinaryName(goos string) string {
	if goos == "windows" {
		return "humblskills.exe"
	}
	return "humblskills"
}
