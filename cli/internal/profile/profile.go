// Package profile persists the user's humblskills preferences: default
// install platforms and default scope. The profile is plain JSON on disk
// and mirrors the manifest package's Load/Save shape.
package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

const SchemaVersion = 1

// Profile is the full on-disk document.
type Profile struct {
	SchemaVersion    int          `json:"schema_version"`
	DefaultPlatforms []string     `json:"default_platforms,omitempty"`
	DefaultScope     string       `json:"default_scope,omitempty"`
	Eval             *EvalProfile `json:"eval,omitempty"`
}

// EvalProfile captures eval-specific defaults. Secrets (API keys) do NOT
// live here - they go through cli/internal/secrets which supports env +
// OS keyring + 0600 file fallback.
type EvalProfile struct {
	Runner                 string `json:"runner,omitempty"`                   // claudecode|cursor-agent|codex|anthropic-api|openai-api|mock
	ExecutorModel          string `json:"executor_model,omitempty"`
	GraderModel            string `json:"grader_model,omitempty"`
	RunsPerConfiguration   int    `json:"runs_per_configuration,omitempty"`
	Parallel               int    `json:"parallel,omitempty"`
	DefaultWorkspace       string `json:"default_workspace,omitempty"`
	IncludeBlindComparator bool   `json:"include_blind_comparator,omitempty"`
}

// DefaultPath resolves the profile path using XDG_CONFIG_HOME (falling back
// to ~/.humblskills/config.json if XDG resolution fails).
func DefaultPath() (string, error) {
	if p, err := xdg.ConfigFile("humblskills/config.json"); err == nil {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve profile path: %w", err)
	}
	return filepath.Join(home, ".humblskills", "config.json"), nil
}

// Load reads the profile at path. If the file doesn't exist, returns an
// empty Profile with the current schema version (not an error).
func Load(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Profile{SchemaVersion: SchemaVersion}, nil
		}
		return nil, fmt.Errorf("read profile: %w", err)
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}
	if p.SchemaVersion == 0 {
		p.SchemaVersion = SchemaVersion
	}
	if p.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("unsupported profile schema_version %d (expected %d)", p.SchemaVersion, SchemaVersion)
	}
	return &p, nil
}

// Save writes the profile to path atomically (write tmp, rename).
func Save(path string, p *Profile) error {
	if p == nil {
		return errors.New("nil profile")
	}
	if p.SchemaVersion == 0 {
		p.SchemaVersion = SchemaVersion
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename tmp: %w", err)
	}
	return nil
}

// Delete removes the profile file. Missing files are not an error.
func Delete(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete profile: %w", err)
	}
	return nil
}

// FilterKnownPlatforms drops platform names that aren't in `known`. Returns
// the cleaned list and the names that were dropped (for caller warnings).
func FilterKnownPlatforms(platforms []string, known map[string]struct{}) (kept, dropped []string) {
	for _, p := range platforms {
		if _, ok := known[p]; ok {
			kept = append(kept, p)
		} else {
			dropped = append(dropped, p)
		}
	}
	return kept, dropped
}

// IsValidScope reports whether s is one of the accepted scope values.
func IsValidScope(s string) bool {
	return s == "" || s == "user" || s == "project"
}
