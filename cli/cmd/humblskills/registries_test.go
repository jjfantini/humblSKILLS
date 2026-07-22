package main

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

func TestSkillIndex_FindAndRegistryOf(t *testing.T) {
	ix := newSkillIndex()
	ix.add("public", []registry.Skill{{Name: "a"}, {Name: "shared"}})
	ix.add("work", []registry.Skill{{Name: "b"}, {Name: "shared"}})

	// find prefers the named origin registry.
	if _, ok := ix.find("work", "b"); !ok {
		t.Fatal("b not found in work")
	}
	// find falls back to any registry when the origin doesn't have it.
	if _, ok := ix.find("nonexistent", "a"); !ok {
		t.Fatal("a not found via fallback")
	}
	if _, ok := ix.find("", "missing"); ok {
		t.Fatal("missing should not resolve")
	}

	// registryOf attributes a skill to a registry that lists it.
	if got := ix.registryOf("a"); got != "public" {
		t.Fatalf("registryOf(a) = %q, want public", got)
	}
	if got := ix.registryOf("b"); got != "work" {
		t.Fatalf("registryOf(b) = %q, want work", got)
	}
	if got := ix.registryOf("missing"); got != "" {
		t.Fatalf("registryOf(missing) = %q, want empty", got)
	}
}

func TestSanitizeRegistryName(t *testing.T) {
	cases := map[string]string{
		"work":        "work",
		"happy robot": "happy-robot",
		"a/b:c":       "a-b-c",
		"":            "registry",
	}
	for in, want := range cases {
		if got := sanitizeRegistryName(in); got != want {
			t.Errorf("sanitizeRegistryName(%q) = %q, want %q", in, got, want)
		}
	}
}
