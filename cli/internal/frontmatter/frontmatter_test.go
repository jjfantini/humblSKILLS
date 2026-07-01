package frontmatter

import (
	"strings"
	"testing"
)

func TestParse_MetadataCanonicalShape(t *testing.T) {
	src := []byte(`---
name: foo
description: A foo skill.
license: MIT
metadata:
  author: jjfantini
  version: 1.2.3
  category: development
  requires:
    - bar
  platforms: [claude-code]
  tags: [example]
---

# Body

Hello.
`)
	fm, body, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "foo" {
		t.Errorf("name: got %q, want %q", fm.Name, "foo")
	}
	if fm.License != "MIT" {
		t.Errorf("license: got %q", fm.License)
	}
	if fm.Metadata.Author != "jjfantini" {
		t.Errorf("metadata.author: got %q", fm.Metadata.Author)
	}
	if fm.Category() != "development" {
		t.Errorf("category accessor: got %q", fm.Category())
	}
	if fm.Version() != "1.2.3" {
		t.Errorf("version accessor: got %q", fm.Version())
	}
	if fm.Metadata.Version != "1.2.3" {
		t.Errorf("metadata.version: got %q", fm.Metadata.Version)
	}
	if got := fm.Requires(); len(got) != 1 || got[0] != "bar" {
		t.Errorf("requires accessor: got %v", got)
	}
	if got := fm.Platforms(); len(got) != 1 || got[0] != "claude-code" {
		t.Errorf("platforms accessor: got %v", got)
	}
	if got := fm.Tags(); len(got) != 1 || got[0] != "example" {
		t.Errorf("tags accessor: got %v", got)
	}
	if warns := fm.DeprecationWarnings(); len(warns) != 0 {
		t.Errorf("expected no deprecation warnings, got %v", warns)
	}
	if !strings.HasPrefix(string(body), "# Body") {
		t.Errorf("body: got %q", string(body))
	}
}

func TestParse_LegacyTopLevelFallback(t *testing.T) {
	src := []byte(`---
name: foo
description: d
version: 0.9.0
requires:
  - bar
platforms: [claude-code]
tags: [legacy]
preserve:
  - references/log.md
---

# Body
`)
	fm, _, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Metadata.Version != "" {
		t.Errorf("metadata.version should be empty on legacy input, got %q", fm.Metadata.Version)
	}
	if got := fm.Version(); got != "0.9.0" {
		t.Errorf("version accessor fallback: got %q, want 0.9.0", got)
	}
	if got := fm.Requires(); len(got) != 1 || got[0] != "bar" {
		t.Errorf("requires fallback: got %v", got)
	}
	if got := fm.Platforms(); len(got) != 1 || got[0] != "claude-code" {
		t.Errorf("platforms fallback: got %v", got)
	}
	if got := fm.Tags(); len(got) != 1 || got[0] != "legacy" {
		t.Errorf("tags fallback: got %v", got)
	}
	if got := fm.Preserve(); len(got) != 1 || got[0] != "references/log.md" {
		t.Errorf("preserve fallback: got %v", got)
	}
	warns := fm.DeprecationWarnings()
	if len(warns) != 5 {
		t.Fatalf("expected 5 deprecation warnings, got %d: %v", len(warns), warns)
	}
	for _, w := range warns {
		if !strings.Contains(w, "deprecated") {
			t.Errorf("warning missing 'deprecated': %q", w)
		}
	}
}

func TestParse_MetadataTakesPrecedence(t *testing.T) {
	src := []byte(`---
name: foo
description: d
version: 0.9.0
tags: [legacy-tag]
metadata:
  version: 1.0.0
  tags: [canonical-tag]
---
body`)
	fm, _, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := fm.Version(); got != "1.0.0" {
		t.Errorf("metadata should win: got %q, want 1.0.0", got)
	}
	if got := fm.Tags(); len(got) != 1 || got[0] != "canonical-tag" {
		t.Errorf("metadata tags should win: got %v", got)
	}
	// Deprecation is still flagged because legacy fields were PRESENT,
	// even though metadata overrode them. Encourages full cleanup.
	warns := fm.DeprecationWarnings()
	if len(warns) == 0 {
		t.Errorf("expected deprecation warnings even when metadata wins; got none")
	}
}

func TestParse_UnknownKeysArePassThrough(t *testing.T) {
	src := []byte(`---
name: foo
description: d
metadata:
  version: 0.1.0
  author: someone
custom_nested:
  key: value
---
body`)
	fm, _, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "foo" {
		t.Errorf("expected parse to succeed with unknown keys; got name %q", fm.Name)
	}
	if fm.Metadata.Author != "someone" {
		t.Errorf("metadata.author: got %q", fm.Metadata.Author)
	}
}

func TestParse_BOMAndLeadingWhitespace(t *testing.T) {
	src := []byte("\uFEFF\n\n---\nname: foo\ndescription: d\nmetadata:\n  version: 0.1.0\n---\n")
	fm, _, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "foo" {
		t.Errorf("name: got %q", fm.Name)
	}
}

func TestParse_MissingOpeningDelimiter(t *testing.T) {
	src := []byte("name: foo\n")
	if _, _, err := Parse(src); err == nil {
		t.Fatal("expected error for missing opening delimiter")
	}
}

func TestParse_MissingClosingDelimiter(t *testing.T) {
	src := []byte("---\nname: foo\ndescription: d\nversion: 0.1.0\n")
	if _, _, err := Parse(src); err == nil {
		t.Fatal("expected error for missing closing delimiter")
	}
}

func TestParse_BadYAML(t *testing.T) {
	src := []byte("---\nname: [unbalanced\n---\n")
	if _, _, err := Parse(src); err == nil {
		t.Fatal("expected yaml parse error")
	}
}

func TestParse_EmptyFrontmatterBlock(t *testing.T) {
	src := []byte("---\n---\nbody")
	fm, body, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "" {
		t.Errorf("expected empty name, got %q", fm.Name)
	}
	if string(body) != "body" {
		t.Errorf("body: got %q", string(body))
	}
}
