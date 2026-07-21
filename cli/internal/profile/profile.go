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
	"time"

	"github.com/adrg/xdg"
)

const SchemaVersion = 1

// Scope constants for Profile.DefaultScope. ScopeGlobal is the recommended
// default: one canonical copy under ~/.humblskills/skills, symlinked to
// every selected platform. ScopeAdapterDefault opts back into each
// platform's own documented default scope (today: "user" for every
// adapter) instead of a concrete scope — it only exists as an explicit,
// profile-only escape hatch because it can't show a concrete location at
// install time.
const (
	ScopeGlobal         = "global"
	ScopeUser           = "user"
	ScopeProject        = "project"
	ScopeAdapterDefault = "adapter-default"
)

// DefaultStatusAutoReturnSeconds is how long a completed status/progress
// screen (registry refresh, install, update) stays on screen before
// auto-returning to the dashboard when StatusAutoReturnSeconds is unset.
const DefaultStatusAutoReturnSeconds = 5

// Profile is the full on-disk document.
type Profile struct {
	SchemaVersion    int      `json:"schema_version"`
	DefaultPlatforms []string `json:"default_platforms,omitempty"`
	DefaultScope     string   `json:"default_scope,omitempty"`
	// Registry, when set, is the default registry URL (or file:// path) used
	// when neither --registry nor HUMBLSKILLS_REGISTRY is provided. Empty means
	// the built-in hosted default.
	Registry string       `json:"registry,omitempty"`
	Eval     *EvalProfile `json:"eval,omitempty"`

	// StatusAutoReturnSeconds controls how long a completed status screen
	// (registry refresh, install, update) waits before automatically
	// dismissing itself and returning to the dashboard. nil (unset) means
	// the built-in default (DefaultStatusAutoReturnSeconds); 0 disables
	// the timer entirely (manual dismiss only, via enter/q/esc); any other
	// positive value is the number of seconds to wait. Only success
	// screens auto-return - a failed run always waits for the user.
	StatusAutoReturnSeconds *int `json:"status_auto_return_seconds,omitempty"`
}

// EvalProfile captures eval-specific defaults. Secrets (API keys) do NOT
// live here - they go through cli/internal/secrets which supports env +
// OS keyring + 0600 file fallback.
type EvalProfile struct {
	Runner                 string `json:"runner,omitempty"` // claudecode|cursor-agent|codex|anthropic-api|openai-api|mock
	ExecutorModel          string `json:"executor_model,omitempty"`
	GraderModel            string `json:"grader_model,omitempty"`
	RunsPerConfiguration   int    `json:"runs_per_configuration,omitempty"`
	Parallel               int    `json:"parallel,omitempty"`
	DefaultWorkspace       string `json:"default_workspace,omitempty"`
	IncludeBlindComparator bool   `json:"include_blind_comparator,omitempty"`
}

// legacyProfilePath is the pre-relocation profile location: XDG_CONFIG_HOME
// (~/Library/Application Support/humblskills/config.json on macOS,
// ~/.config/humblskills/config.json on Linux). DefaultPath migrates a file
// found there into the new ~/.humblskills/profile.json location so every
// humblskills file - skills, manifest fallback, and now the profile - lives
// together instead of fanning out across OS-specific config directories.
func legacyProfilePath() (string, error) {
	return xdg.SearchConfigFile("humblskills/config.json")
}

// DefaultPath resolves the profile path to ~/.humblskills/profile.json,
// alongside the canonical skill store (~/.humblskills/skills). If a legacy
// profile from the old XDG-config location exists and nothing has been
// written to the new location yet, it's moved (not copied) into place so
// there's exactly one profile file on disk afterward.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve profile path: %w", err)
	}
	newPath := filepath.Join(home, ".humblskills", "profile.json")

	if _, err := os.Stat(newPath); err == nil {
		return newPath, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("resolve profile path: %w", err)
	}

	if legacy, err := legacyProfilePath(); err == nil {
		_ = migrateLegacyProfile(legacy, newPath)
	}
	return newPath, nil
}

// migrateLegacyProfile moves the file at legacy to newPath: write first,
// then remove the source, and only once the write has succeeded. Copy-then-
// delete (not os.Rename) because the two paths can be on different
// filesystems - e.g. macOS's Application Support volume vs $HOME.
func migrateLegacyProfile(legacy, newPath string) error {
	data, err := os.ReadFile(legacy)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(newPath, data, 0o644); err != nil {
		return err
	}
	return os.Remove(legacy)
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

// IsValidScope reports whether s is one of the accepted scope values. "" is
// accepted as shorthand for the unset/default state — see ResolvedScope.
func IsValidScope(s string) bool {
	switch s {
	case "", ScopeGlobal, ScopeUser, ScopeProject, ScopeAdapterDefault:
		return true
	}
	return false
}

// ResolvedScope returns the profile's effective default scope, treating an
// unset DefaultScope ("") as ScopeGlobal — the recommended default of one
// canonical store symlinked to every selected platform. Callers that need to
// distinguish "explicitly global" from "unset" should read DefaultScope
// directly instead.
func (p *Profile) ResolvedScope() string {
	if p == nil || p.DefaultScope == "" {
		return ScopeGlobal
	}
	return p.DefaultScope
}

// StatusAutoReturnDuration resolves StatusAutoReturnSeconds to a
// time.Duration: nil -> the built-in default, 0 -> disabled (zero
// duration means "wait for the user", never an instant auto-dismiss),
// N>0 -> N seconds. Negative values are treated as disabled too, since
// there's no sane "negative seconds" reading.
func (p *Profile) StatusAutoReturnDuration() time.Duration {
	if p == nil || p.StatusAutoReturnSeconds == nil {
		return DefaultStatusAutoReturnSeconds * time.Second
	}
	if *p.StatusAutoReturnSeconds <= 0 {
		return 0
	}
	return time.Duration(*p.StatusAutoReturnSeconds) * time.Second
}
