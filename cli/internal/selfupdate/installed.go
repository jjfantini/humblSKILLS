package selfupdate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// versionJSON mirrors the shape cmd/humblskills/version.go's versionInfo
// marshals to. It can't be imported directly (that's package main), so the
// fields selfupdate cares about are duplicated here.
type versionJSON struct {
	Version string `json:"version"`
}

// VerifyInstalledVersion shells out to `exePath version --json` and returns
// the version it reports, so callers can confirm a swap or `brew upgrade`
// actually landed the expected release instead of just trusting that the
// checksum-verified download was installed correctly.
func VerifyInstalledVersion(exePath string) (string, error) {
	cmd := exec.Command(exePath, "version", "--json")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run %s version --json: %w", exePath, err)
	}
	var v versionJSON
	if err := json.Unmarshal(out.Bytes(), &v); err != nil {
		return "", fmt.Errorf("parse %s version --json output: %w", exePath, err)
	}
	if v.Version == "" {
		return "", fmt.Errorf("%s version --json returned an empty version", exePath)
	}
	return v.Version, nil
}
