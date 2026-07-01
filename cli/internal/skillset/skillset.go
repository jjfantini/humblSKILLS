// Package skillset reads and writes the shareable skillset file: a small,
// commit-to-your-repo manifest of the skills a project (or a person) wants
// installed. `humblskills export` writes one from the local install manifest;
// `humblskills sync` installs everything it lists. It's the collaboration
// primitive - check `humblskills.json` into a repo and every teammate runs
// `humblskills sync` to land the same skill set.
package skillset

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// SchemaVersion is the current skillset schema. Bumped on breaking changes.
const SchemaVersion = 1

// DefaultFilename is the conventional skillset filename, resolved relative to
// the current directory so it can live at a repo root.
const DefaultFilename = "humblskills.json"

// Skill is one entry in a skillset. Version is informational (the version
// captured at export time); `sync` installs whatever the registry currently
// ships, matching `install` semantics.
type Skill struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Set is the full skillset document.
type Set struct {
	SchemaVersion int     `json:"schema_version"`
	Skills        []Skill `json:"skills"`
}

// New returns an empty, current-schema Set. Skills is a non-nil empty slice so
// an unmodified Set serializes as "skills": [] (an editable scaffold) rather
// than "skills": null.
func New() *Set {
	return &Set{SchemaVersion: SchemaVersion, Skills: []Skill{}}
}

// Add records a skill, de-duplicating by name (last version wins). Blank names
// are ignored.
func (s *Set) Add(name, version string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	for i := range s.Skills {
		if s.Skills[i].Name == name {
			s.Skills[i].Version = version
			return
		}
	}
	s.Skills = append(s.Skills, Skill{Name: name, Version: version})
}

// Sort orders skills by name for stable, diff-friendly output.
func (s *Set) Sort() {
	sort.Slice(s.Skills, func(i, j int) bool { return s.Skills[i].Name < s.Skills[j].Name })
}

// Names returns the skill names in file order.
func (s *Set) Names() []string {
	out := make([]string, 0, len(s.Skills))
	for _, sk := range s.Skills {
		out = append(out, sk.Name)
	}
	return out
}

// Validate checks the document is well-formed and safe to act on.
func (s *Set) Validate() error {
	if s.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported skillset schema_version %d (expected %d)", s.SchemaVersion, SchemaVersion)
	}
	seen := map[string]bool{}
	for _, sk := range s.Skills {
		if strings.TrimSpace(sk.Name) == "" {
			return fmt.Errorf("skillset contains a skill with an empty name")
		}
		if seen[sk.Name] {
			return fmt.Errorf("skillset lists %q more than once", sk.Name)
		}
		seen[sk.Name] = true
	}
	return nil
}

// Load reads and validates a skillset file.
func Load(path string) (*Set, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read skillset: %w", err)
	}
	var s Set
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse skillset %s: %w", path, err)
	}
	// A hand-written file may omit schema_version; treat 0 as current so a
	// minimal {"skills":[...]} works out of the box.
	if s.SchemaVersion == 0 {
		s.SchemaVersion = SchemaVersion
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// Save writes a skillset file (pretty-printed, trailing newline) after sorting
// for stable diffs.
func Save(path string, s *Set) error {
	if s.SchemaVersion == 0 {
		s.SchemaVersion = SchemaVersion
	}
	s.Sort()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal skillset: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write skillset %s: %w", path, err)
	}
	return nil
}
