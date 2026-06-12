// Package manifest reads and writes the local install manifest that tracks
// which skills humblskills has installed on this machine.
package manifest

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

// SchemaVersion is the current manifest schema. Bumped on breaking changes.
const SchemaVersion = 1

// Manifest is the full on-disk document.
type Manifest struct {
	SchemaVersion int            `json:"schema_version"`
	Installations []Installation `json:"installations"`
}

// Installation is one installed skill.
type Installation struct {
	Skill       string    `json:"skill"`
	Version     string    `json:"version"`
	Platform    string    `json:"platform"`
	Scope       string    `json:"scope"`
	Path        string    `json:"path"`
	StorePath   string    `json:"store_path,omitempty"`
	InstallMode string    `json:"install_mode,omitempty"`
	InstalledAt time.Time `json:"installed_at"`
	SourceSHA   string    `json:"source_sha"`
	RegistryRef string    `json:"registry_ref"`
}

// DefaultPath resolves the manifest path using XDG_STATE_HOME (falling back
// to ~/.local/state per spec, then ~/.humblskills/manifest.json as last
// resort if XDG resolution fails).
func DefaultPath() (string, error) {
	if p, err := xdg.StateFile("humblskills/manifest.json"); err == nil {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve manifest path: %w", err)
	}
	return filepath.Join(home, ".humblskills", "manifest.json"), nil
}

// Load reads the manifest at path. If the file doesn't exist, returns an
// empty Manifest and no error — a first-run user has no installations.
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Manifest{SchemaVersion: SchemaVersion}, nil
		}
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.SchemaVersion == 0 {
		// Legacy / hand-written file: treat as current schema and continue.
		m.SchemaVersion = SchemaVersion
	}
	if m.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("unsupported manifest schema_version %d (expected %d)", m.SchemaVersion, SchemaVersion)
	}
	return &m, nil
}

// Save writes the manifest to path atomically (write tmp, rename). Creates
// parent directories as needed.
func Save(path string, m *Manifest) error {
	if m == nil {
		return errors.New("nil manifest")
	}
	if m.SchemaVersion == 0 {
		m.SchemaVersion = SchemaVersion
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
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

// Find returns a pointer to the first Installation matching the given skill
// name, or nil if not installed. Use FindAll when a skill may be installed on
// multiple (platform, scope) pairs.
func (m *Manifest) Find(skill string) *Installation {
	for i := range m.Installations {
		if m.Installations[i].Skill == skill {
			return &m.Installations[i]
		}
	}
	return nil
}

// FindAll returns pointers to every Installation of the given skill, across
// platforms and scopes.
func (m *Manifest) FindAll(skill string) []*Installation {
	var out []*Installation
	for i := range m.Installations {
		if m.Installations[i].Skill == skill {
			out = append(out, &m.Installations[i])
		}
	}
	return out
}

// FindOne returns the Installation pinned to exactly (skill, platform, scope),
// or nil if none matches.
func (m *Manifest) FindOne(skill, platform, scope string) *Installation {
	for i := range m.Installations {
		e := &m.Installations[i]
		if e.Skill == skill && e.Platform == platform && e.Scope == scope {
			return e
		}
	}
	return nil
}

// Upsert inserts or replaces the Installation keyed on (skill, platform,
// scope).
func (m *Manifest) Upsert(inst Installation) {
	for i := range m.Installations {
		e := &m.Installations[i]
		if e.Skill == inst.Skill && e.Platform == inst.Platform && e.Scope == inst.Scope {
			m.Installations[i] = inst
			return
		}
	}
	m.Installations = append(m.Installations, inst)
}

// Remove drops every Installation matching skill (all platforms and scopes)
// and returns how many were removed.
func (m *Manifest) Remove(skill string) int {
	kept := m.Installations[:0]
	removed := 0
	for _, e := range m.Installations {
		if e.Skill == skill {
			removed++
			continue
		}
		kept = append(kept, e)
	}
	m.Installations = kept
	return removed
}

// RemoveOne drops exactly the Installation pinned to (skill, platform, scope).
// Returns true if an entry was removed.
func (m *Manifest) RemoveOne(skill, platform, scope string) bool {
	for i, e := range m.Installations {
		if e.Skill == skill && e.Platform == platform && e.Scope == scope {
			m.Installations = append(m.Installations[:i], m.Installations[i+1:]...)
			return true
		}
	}
	return false
}
