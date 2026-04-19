package install

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
)

// parseMapping is a tiny helper: unmarshal a YAML block string into its
// mapping node so tests can drive setPreserveKey directly.
func parseMapping(t *testing.T, yml string) *yaml.Node {
	t.Helper()
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(yml), &root); err != nil {
		t.Fatalf("parse yaml: %v", err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		t.Fatalf("unexpected root: %#v", root)
	}
	m := root.Content[0]
	if m.Kind != yaml.MappingNode {
		t.Fatalf("root is not a mapping: %#v", m)
	}
	return m
}

func encode(t *testing.T, node *yaml.Node) string {
	t.Helper()
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		t.Fatalf("encode: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return buf.String()
}

// mappingKeys returns the ordered list of top-level scalar keys in a mapping.
func mappingKeys(m *yaml.Node) []string {
	var out []string
	for i := 0; i+1 < len(m.Content); i += 2 {
		out = append(out, m.Content[i].Value)
	}
	return out
}

// mappingSub returns the mapping node at `key` or nil.
func mappingSub(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func TestSetPreserveKey_WritesIntoExistingMetadata(t *testing.T) {
	src := `
name: foo
description: d
metadata:
  author: me
  version: 1.0.0
`
	m := parseMapping(t, src)
	setPreserveKey(m, []string{"references/log.md"})

	meta := mappingSub(m, "metadata")
	if meta == nil || meta.Kind != yaml.MappingNode {
		t.Fatalf("metadata mapping missing or wrong kind: %#v", meta)
	}

	keys := mappingKeys(meta)
	wantKeys := []string{"author", "version", "preserve"}
	if len(keys) != len(wantKeys) {
		t.Fatalf("metadata keys: got %v, want %v", keys, wantKeys)
	}
	for i, k := range wantKeys {
		if keys[i] != k {
			t.Errorf("metadata keys[%d]: got %q, want %q", i, keys[i], k)
		}
	}

	// Existing metadata fields must survive untouched.
	if mappingSub(meta, "author").Value != "me" {
		t.Errorf("author corrupted: %v", mappingSub(meta, "author").Value)
	}
	if mappingSub(meta, "version").Value != "1.0.0" {
		t.Errorf("version corrupted: %v", mappingSub(meta, "version").Value)
	}

	// Preserve should be the only sequence with our entry.
	p := mappingSub(meta, "preserve")
	if p == nil || p.Kind != yaml.SequenceNode || len(p.Content) != 1 || p.Content[0].Value != "references/log.md" {
		t.Errorf("preserve seq wrong: %#v", p)
	}
}

func TestSetPreserveKey_CreatesMetadataWhenAbsent(t *testing.T) {
	src := `
name: foo
description: d
`
	m := parseMapping(t, src)
	setPreserveKey(m, []string{"log.md"})

	meta := mappingSub(m, "metadata")
	if meta == nil {
		t.Fatal("metadata should have been created")
	}
	if meta.Kind != yaml.MappingNode {
		t.Errorf("metadata should be a mapping, got kind %d", meta.Kind)
	}

	p := mappingSub(meta, "preserve")
	if p == nil || p.Kind != yaml.SequenceNode || len(p.Content) != 1 || p.Content[0].Value != "log.md" {
		t.Errorf("preserve entry wrong: %#v", p)
	}

	// metadata must be appended at the END of top-level keys (stable order).
	keys := mappingKeys(m)
	if keys[len(keys)-1] != "metadata" {
		t.Errorf("metadata should be last top-level key, got %v", keys)
	}
}

func TestSetPreserveKey_MigratesLegacyTopLevelPreserve(t *testing.T) {
	src := `
name: foo
description: d
preserve:
  - legacy-entry.md
metadata:
  version: 1.0.0
`
	m := parseMapping(t, src)
	setPreserveKey(m, []string{"new-entry.md"})

	// Legacy top-level preserve must be removed.
	for _, k := range mappingKeys(m) {
		if k == "preserve" {
			t.Errorf("legacy top-level preserve should have been removed; keys: %v", mappingKeys(m))
		}
	}

	// Canonical preserve lives under metadata with the NEW value.
	meta := mappingSub(m, "metadata")
	p := mappingSub(meta, "preserve")
	if p == nil || len(p.Content) != 1 || p.Content[0].Value != "new-entry.md" {
		t.Errorf("metadata.preserve wrong: %#v", p)
	}

	// Round-trip: re-parse the rendered YAML through the frontmatter parser
	// to confirm the result is a valid Anthropic-compliant frontmatter.
	rendered := "---\n" + encode(t, m) + "---\n\n# body\n"
	fm, _, err := frontmatter.Parse([]byte(rendered))
	if err != nil {
		t.Fatalf("re-parse rendered frontmatter: %v", err)
	}
	if got := fm.Preserve(); len(got) != 1 || got[0] != "new-entry.md" {
		t.Errorf("re-parsed preserve: got %v", got)
	}
	if warns := fm.DeprecationWarnings(); len(warns) != 0 {
		t.Errorf("re-parsed frontmatter should have no deprecations; got %v", warns)
	}
}

func TestSetPreserveKey_OverwritesExistingMetadataPreserve(t *testing.T) {
	src := `
name: foo
description: d
metadata:
  version: 1.0.0
  preserve:
    - old-entry.md
`
	m := parseMapping(t, src)
	setPreserveKey(m, []string{"a.md", "b.md"})

	meta := mappingSub(m, "metadata")
	p := mappingSub(meta, "preserve")
	if p == nil || len(p.Content) != 2 {
		t.Fatalf("expected 2 preserve entries, got %#v", p)
	}
	if p.Content[0].Value != "a.md" || p.Content[1].Value != "b.md" {
		t.Errorf("preserve content wrong: %v, %v", p.Content[0].Value, p.Content[1].Value)
	}
}

func TestSetPreserveKey_EmptyListWritesFlowSequence(t *testing.T) {
	src := `
name: foo
description: d
metadata:
  version: 1.0.0
`
	m := parseMapping(t, src)
	setPreserveKey(m, nil)

	meta := mappingSub(m, "metadata")
	p := mappingSub(meta, "preserve")
	if p == nil || p.Kind != yaml.SequenceNode {
		t.Fatalf("preserve should be a sequence node, got %#v", p)
	}
	if p.Style != yaml.FlowStyle {
		t.Errorf("empty preserve should use flow style for `[]` rendering; got style %d", p.Style)
	}
	if len(p.Content) != 0 {
		t.Errorf("empty preserve should have no children; got %d", len(p.Content))
	}
}

func TestSetPreserveKey_RepairsNonMapMetadata(t *testing.T) {
	// Edge case: metadata exists but is null or a scalar instead of a mapping.
	// setPreserveKey must recover rather than panic or corrupt.
	src := `
name: foo
description: d
metadata: null
`
	m := parseMapping(t, src)
	setPreserveKey(m, []string{"log.md"})

	meta := mappingSub(m, "metadata")
	if meta == nil {
		t.Fatal("metadata missing")
	}
	if meta.Kind != yaml.MappingNode {
		t.Errorf("metadata should be coerced to mapping; got kind %d", meta.Kind)
	}
	p := mappingSub(meta, "preserve")
	if p == nil || len(p.Content) != 1 || p.Content[0].Value != "log.md" {
		t.Errorf("preserve after coercion wrong: %#v", p)
	}
}

// TestMergePreserveIntoSkillMD_EndToEnd exercises the full file rewrite:
// legacy top-level preserve is removed, metadata.preserve replaces it,
// body bytes below the closing `---` survive untouched.
func TestMergePreserveIntoSkillMD_EndToEnd(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")

	before := `---
name: foo
description: A foo skill.
version: 0.1.0
preserve:
  - old.md
---

# Body

Content below the frontmatter. This must stay byte-exact.
`
	if err := os.WriteFile(path, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := mergePreserveIntoSkillMD(path, []string{"user-edit.md", "log.md"}); err != nil {
		t.Fatalf("merge: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gotS := string(got)

	// Body must survive byte-for-byte.
	if !strings.Contains(gotS, "# Body\n\nContent below the frontmatter. This must stay byte-exact.") {
		t.Errorf("body corrupted:\n%s", gotS)
	}

	// Legacy top-level `preserve:` must be gone.
	topLevelPreserve := regexp.MustCompile(`(?m)^preserve:`)
	if topLevelPreserve.MatchString(gotS) {
		t.Errorf("legacy top-level `preserve:` should have been stripped:\n%s", gotS)
	}

	// New values must land under metadata.preserve (indented).
	if !strings.Contains(gotS, "- user-edit.md") || !strings.Contains(gotS, "- log.md") {
		t.Errorf("new preserve entries missing:\n%s", gotS)
	}

	// Round-trip through the frontmatter parser: accessor returns the new
	// list, and there are no deprecation warnings left for preserve.
	fm, _, err := frontmatter.Parse(got)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if p := fm.Preserve(); len(p) != 2 || p[0] != "user-edit.md" || p[1] != "log.md" {
		t.Errorf("preserve accessor wrong: %v", p)
	}
	for _, w := range fm.DeprecationWarnings() {
		if strings.Contains(w, "preserve") {
			t.Errorf("preserve deprecation should be gone after merge; got %q", w)
		}
	}
}
