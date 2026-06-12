package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

const (
	InstallModeLinked = "linked"
	InstallModeGlobal = "global"
)

// CanonicalSkillPath returns the humblskills-owned source directory for a
// skill. Platform install targets should point at this directory with symlinks.
func CanonicalSkillPath(skillName, scope string, global bool) (string, error) {
	if skillName == "" {
		return "", fmt.Errorf("canonical store: empty skill name")
	}
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home: %w", err)
		}
		return filepath.Join(home, ".humblskills", "skills", skillName), nil
	}
	if scope == "project" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve cwd: %w", err)
		}
		return filepath.Join(cwd, ".humblskills", "skills", skillName), nil
	}
	if p, err := xdg.DataFile(filepath.Join("humblskills", "skills", skillName)); err == nil {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(home, ".local", "share", "humblskills", "skills", skillName), nil
}

func installMode(global bool) string {
	if global {
		return InstallModeGlobal
	}
	return InstallModeLinked
}
