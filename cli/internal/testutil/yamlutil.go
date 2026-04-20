package testutil

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v3"
)

// ParseYAMLMapping unmarshals yml into a YAML mapping node. Fails the
// test if the input is not a top-level mapping. Shared helper for
// tests that need to inspect or rewrite frontmatter-style YAML.
func ParseYAMLMapping(t testing.TB, yml string) *yaml.Node {
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

// EncodeYAML serialises a YAML node at indent 2 and returns the result.
func EncodeYAML(t testing.TB, node *yaml.Node) string {
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

// MappingKeys returns the ordered list of top-level scalar keys in a
// YAML mapping node.
func MappingKeys(m *yaml.Node) []string {
	var out []string
	for i := 0; i+1 < len(m.Content); i += 2 {
		out = append(out, m.Content[i].Value)
	}
	return out
}

// MappingSub returns the node stored under key in mapping m, or nil if
// key is absent.
func MappingSub(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}
