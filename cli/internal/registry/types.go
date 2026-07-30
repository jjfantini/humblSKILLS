// Package registry defines the registry.json schema and (in later phases) the
// fetcher / cache for consuming it from GitHub raw.
package registry

import "github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"

// SchemaVersion is the current registry schema. Bumped on breaking change.
const SchemaVersion = 1

// Registry is the top-level registry.json document.
type Registry struct {
	SchemaVersion int     `json:"schema_version"`
	GeneratedAt   string  `json:"generated_at"`
	Source        Source  `json:"source"`
	Skills        []Skill `json:"skills"`
}

// Source identifies the upstream repo and commit that produced this registry.
type Source struct {
	Repo string `json:"repo"`
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
}

// Skill is one entry in the registry.
type Skill struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Category    string `json:"category,omitempty"`
	// Role is the optional target role this skill is scoped to (closed set,
	// see frontmatter.Roles). Empty means unscoped; single-valued like
	// Category. Used to sub-group skills within a registry when browsing.
	Role      string   `json:"role,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Platforms []string `json:"platforms,omitempty"`
	Requires  []string `json:"requires,omitempty"`
	Preserve  []string `json:"preserve,omitempty"`
	// PreviousNames are names this skill was published under before a rename.
	// The installer uses them to find and carry over an existing installation's
	// preserved data instead of stranding it under the old name.
	PreviousNames []string `json:"previous_names,omitempty"`
	// Upstream is set for skills that mirror a skill published elsewhere.
	// Carried through from SKILL.md so consumers can render provenance and
	// detect drift without reading every skill directory.
	Upstream *frontmatter.Upstream `json:"upstream,omitempty"`
	Path     string                `json:"path"`
	DirSHA   string                `json:"dir_sha"`
	// Registry is the source-registry name, stamped at load time when skills
	// from multiple registries are aggregated. It is never written to a
	// registry.json on disk (build-registry leaves it empty), so it does not
	// affect generation or the --check semantic diff.
	Registry string `json:"registry,omitempty"`
}
