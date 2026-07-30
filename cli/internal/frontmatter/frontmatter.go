// Package frontmatter parses SKILL.md YAML frontmatter and the humblSKILLS
// extension keys layered on top of the agentskills.io base format.
//
// Shape (Anthropic-compliant top level, humblSKILLS extensions under metadata):
//
//	---
//	name: my-skill
//	description: ...
//	license: MIT                       # optional
//	compatibility: ...                 # optional, <=500 chars
//	allowed-tools: "Bash(bash:*) Read" # optional
//	metadata:
//	  author: jjfantini
//	  version: 1.0.0                   # humblSKILLS requires this
//	  category: development            # humblSKILLS requires this, one of a closed set
//	  role: fde                        # optional, one of a closed set (see validate.Roles)
//	  tags: [...]
//	  platforms: [claude-code]
//	  requires: [...]
//	  preserve: [...]
//	---
//
// Legacy top-level `version`, `tags`, `platforms`, `requires`, `preserve`
// are still parsed for a deprecation period. They populate the accessor
// return values only when the corresponding metadata field is empty, and
// [Frontmatter.DeprecationWarnings] surfaces them so tooling can warn
// without failing the build.
package frontmatter

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

const delimiter = "---"

// Metadata is the humblSKILLS extension block nested under `metadata:` in
// SKILL.md frontmatter. Top level stays clean / agentskills.io-compatible.
type Metadata struct {
	Author    string   `yaml:"author,omitempty"`
	Version   string   `yaml:"version,omitempty"`
	Category  string   `yaml:"category,omitempty"`
	Role      string   `yaml:"role,omitempty"`
	Requires  []string `yaml:"requires,omitempty"`
	Platforms []string `yaml:"platforms,omitempty"`
	Tags      []string `yaml:"tags,omitempty"`
	Preserve  []string `yaml:"preserve,omitempty"`
	// PreviousNames lists names this skill was published under before being
	// renamed. Without it a rename reads as "a brand new skill installed next
	// to an unrelated old one", and the user's preserved brain data is stranded
	// in a directory nothing points at. See install.Engine for how it is used.
	PreviousNames []string `yaml:"previous_names,omitempty"`
}

// Upstream declares that a skill is a mirror of a skill published elsewhere.
//
// It is the single source of truth for provenance: ATTRIBUTION.md is rendered
// from it, and `humblskills mirrors` uses it to detect drift.
//
// Preserved is the key field. It points at a byte-for-byte copy of the upstream
// artifact under references/raw/, which doubles as the drift baseline - the
// current upstream is compared against it directly, so no hash or lockfile is
// needed. Overwriting that file without re-distilling destroys the baseline.
type Upstream struct {
	// Name is the upstream skill's own name, which need not match ours.
	Name string `yaml:"name" json:"name"`
	// Source is a human-readable origin, e.g. "jakubkrehel/skills".
	Source string `yaml:"source" json:"source"`
	// URL is the origin's homepage, for attribution rendering.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Fetch is where the CURRENT upstream can be read for drift detection:
	// either an https:// URL to the raw file, or a path (may contain ~) to a
	// local plugin cache. Empty means drift cannot be checked automatically.
	Fetch string `yaml:"fetch,omitempty" json:"fetch,omitempty"`
	// License is the upstream licence identifier.
	License string `yaml:"license,omitempty" json:"license,omitempty"`
	// Preserved is the repo-relative path to the verbatim upstream copy that
	// serves as the drift baseline.
	Preserved string `yaml:"preserved" json:"preserved"`
	// Synced is the ISO date the preserved copy was last refreshed.
	Synced string `yaml:"synced,omitempty" json:"synced,omitempty"`
	// Deltas are intentional differences from upstream. A re-sync must NOT
	// "fix" these; they are the guardrail that stops each sync from
	// re-litigating deliberate choices.
	Deltas []string `yaml:"deltas,omitempty" json:"deltas,omitempty"`
}

// Frontmatter is the typed view of a SKILL.md frontmatter block. Unknown keys
// in the YAML are accepted but not surfaced; humblSKILLS only validates the
// fields it owns.
//
// Read humblSKILLS-owned fields through the accessor methods
// ([Frontmatter.Version], [Frontmatter.Category], [Frontmatter.Requires],
// [Frontmatter.Platforms], [Frontmatter.Tags], [Frontmatter.Preserve])
// rather than reaching into [Frontmatter.Metadata] directly - the accessors
// honour the legacy top-level fallback (except Category, which has none).
type Frontmatter struct {
	Name          string    `yaml:"name"`
	Description   string    `yaml:"description"`
	License       string    `yaml:"license,omitempty"`
	Compatibility string    `yaml:"compatibility,omitempty"`
	AllowedTools  string    `yaml:"allowed-tools,omitempty"`
	Upstream      *Upstream `yaml:"upstream,omitempty"`
	Metadata      Metadata  `yaml:"metadata"`

	// Legacy top-level fields. Populated by [Parse] when present in the
	// YAML. Kept unexported so callers are forced through the accessors,
	// which fall back to these only when [Metadata] is empty for that key.
	legacyVersion   string
	legacyRequires  []string
	legacyPlatforms []string
	legacyTags      []string
	legacyPreserve  []string
}

// Version returns the humblSKILLS version: metadata.version first, then
// the deprecated top-level `version:` key if metadata is empty.
func (f Frontmatter) Version() string {
	if f.Metadata.Version != "" {
		return f.Metadata.Version
	}
	return f.legacyVersion
}

// Category returns the skill's single coarse category (e.g. "development",
// "design"), used for browsing and filtering. Unlike Version/Requires/etc.
// there is no deprecated top-level fallback - category was introduced after
// the legacy-field migration window, so metadata is the only source.
func (f Frontmatter) Category() string {
	return f.Metadata.Category
}

// Role returns the skill's target role (e.g. "fde", "sdr"), used to group
// and filter skills by who they serve. Optional; empty means unscoped. Like
// Category there is no deprecated top-level fallback - role postdates the
// legacy-field migration window, so metadata is the only source.
func (f Frontmatter) Role() string {
	return f.Metadata.Role
}

// Requires returns the humblSKILLS dep list: metadata.requires first,
// then the deprecated top-level `requires:` key if metadata is empty.
func (f Frontmatter) Requires() []string {
	if len(f.Metadata.Requires) > 0 {
		return f.Metadata.Requires
	}
	return f.legacyRequires
}

// Platforms returns the humblSKILLS platforms list: metadata.platforms first,
// then the deprecated top-level `platforms:` key if metadata is empty.
func (f Frontmatter) Platforms() []string {
	if len(f.Metadata.Platforms) > 0 {
		return f.Metadata.Platforms
	}
	return f.legacyPlatforms
}

// Tags returns the humblSKILLS tags list: metadata.tags first, then the
// deprecated top-level `tags:` key if metadata is empty.
func (f Frontmatter) Tags() []string {
	if len(f.Metadata.Tags) > 0 {
		return f.Metadata.Tags
	}
	return f.legacyTags
}

// Preserve returns the humblSKILLS preserve list: metadata.preserve first,
// then the deprecated top-level `preserve:` key if metadata is empty.
func (f Frontmatter) Preserve() []string {
	if len(f.Metadata.Preserve) > 0 {
		return f.Metadata.Preserve
	}
	return f.legacyPreserve
}

// PreviousNames returns the names this skill was published under before a
// rename. There is no deprecated top-level fallback - the key postdates the
// legacy-field migration window.
func (f Frontmatter) PreviousNames() []string {
	return f.Metadata.PreviousNames
}

// DeprecationWarnings returns a list of human-readable strings describing
// humblSKILLS fields that still live at the top level of the frontmatter
// and should be moved under `metadata:`. Empty slice means the skill is
// fully migrated. Tooling (build-registry, lint) should print these as
// warnings but not fail.
func (f Frontmatter) DeprecationWarnings() []string {
	var out []string
	if f.legacyVersion != "" {
		out = append(out, "top-level `version:` is deprecated; move to `metadata.version`")
	}
	if len(f.legacyRequires) > 0 {
		out = append(out, "top-level `requires:` is deprecated; move to `metadata.requires`")
	}
	if len(f.legacyPlatforms) > 0 {
		out = append(out, "top-level `platforms:` is deprecated; move to `metadata.platforms`")
	}
	if len(f.legacyTags) > 0 {
		out = append(out, "top-level `tags:` is deprecated; move to `metadata.tags`")
	}
	if len(f.legacyPreserve) > 0 {
		out = append(out, "top-level `preserve:` is deprecated; move to `metadata.preserve`")
	}
	return out
}

// frontmatterWire is the wire shape unmarshaled directly from YAML. It
// captures BOTH the canonical metadata block AND the deprecated top-level
// fields so [Parse] can reconcile them into [Frontmatter].
type frontmatterWire struct {
	Name          string    `yaml:"name"`
	Description   string    `yaml:"description"`
	License       string    `yaml:"license,omitempty"`
	Compatibility string    `yaml:"compatibility,omitempty"`
	AllowedTools  string    `yaml:"allowed-tools,omitempty"`
	Upstream      *Upstream `yaml:"upstream,omitempty"`
	Metadata      Metadata  `yaml:"metadata"`

	Version   string   `yaml:"version,omitempty"`
	Requires  []string `yaml:"requires,omitempty"`
	Platforms []string `yaml:"platforms,omitempty"`
	Tags      []string `yaml:"tags,omitempty"`
	Preserve  []string `yaml:"preserve,omitempty"`
}

// Parse splits the leading `---` YAML frontmatter block from the body and
// unmarshals it. Returns the parsed frontmatter, the remaining body bytes, and
// any parse error.
func Parse(data []byte) (Frontmatter, []byte, error) {
	var fm Frontmatter

	rest := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	rest = bytes.TrimLeft(rest, " \t\r\n")

	if !bytes.HasPrefix(rest, []byte(delimiter)) {
		return fm, nil, errors.New("missing leading '---' frontmatter delimiter")
	}
	rest = rest[len(delimiter):]
	rest = bytes.TrimLeft(rest, " \t\r")
	if len(rest) == 0 || rest[0] != '\n' {
		return fm, nil, errors.New("opening '---' must be followed by a newline")
	}
	rest = rest[1:]

	end := findClosingDelimiter(rest)
	if end < 0 {
		return fm, nil, errors.New("missing closing '---' frontmatter delimiter")
	}

	yamlBlock := rest[:end]
	body := rest[end+len(delimiter):]
	body = bytes.TrimLeft(body, " \t\r\n")

	var wire frontmatterWire
	if err := yaml.Unmarshal(yamlBlock, &wire); err != nil {
		return fm, nil, fmt.Errorf("yaml parse: %w", err)
	}

	fm = Frontmatter{
		Name:            wire.Name,
		Description:     wire.Description,
		License:         wire.License,
		Compatibility:   wire.Compatibility,
		AllowedTools:    wire.AllowedTools,
		Upstream:        wire.Upstream,
		Metadata:        wire.Metadata,
		legacyVersion:   wire.Version,
		legacyRequires:  wire.Requires,
		legacyPlatforms: wire.Platforms,
		legacyTags:      wire.Tags,
		legacyPreserve:  wire.Preserve,
	}
	return fm, body, nil
}

// findClosingDelimiter returns the byte offset of a line containing exactly
// "---" (possibly with trailing whitespace), or -1 if not found.
func findClosingDelimiter(data []byte) int {
	idx := 0
	for idx < len(data) {
		lineEnd := bytes.IndexByte(data[idx:], '\n')
		var line []byte
		if lineEnd < 0 {
			line = data[idx:]
		} else {
			line = data[idx : idx+lineEnd]
		}
		if bytes.Equal(bytes.TrimRight(line, " \t\r"), []byte(delimiter)) {
			return idx
		}
		if lineEnd < 0 {
			return -1
		}
		idx += lineEnd + 1
	}
	return -1
}
