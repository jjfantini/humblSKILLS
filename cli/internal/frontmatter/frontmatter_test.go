package frontmatter

import (
	"strings"
	"testing"
)

func TestParse_Valid(t *testing.T) {
	src := []byte(`---
name: foo
description: A foo skill.
version: 1.2.3
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
	if fm.Version != "1.2.3" {
		t.Errorf("version: got %q", fm.Version)
	}
	if len(fm.Requires) != 1 || fm.Requires[0] != "bar" {
		t.Errorf("requires: got %v", fm.Requires)
	}
	if !strings.HasPrefix(string(body), "# Body") {
		t.Errorf("body: got %q", string(body))
	}
}

func TestParse_UnknownKeysArePassThrough(t *testing.T) {
	src := []byte(`---
name: foo
description: d
version: 0.1.0
license: MIT
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
}

func TestParse_BOMAndLeadingWhitespace(t *testing.T) {
	src := []byte("\uFEFF\n\n---\nname: foo\ndescription: d\nversion: 0.1.0\n---\n")
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
