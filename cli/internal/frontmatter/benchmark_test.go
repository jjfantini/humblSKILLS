package frontmatter_test

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
)

var benchFrontmatter = []byte(`---
name: bench-skill
description: A skill used in benchmarks; description is long enough to satisfy
version: 1.2.3
metadata:
  tags: [a, b, c, d, e]
  requires:
    - foo@1.0.0
    - bar@>=2.0.0
    - baz
  platforms: [claude-code, cursor]
  preserve:
    - data/log.md
    - notes/
---

# Bench skill body

Some body content that simulates a realistic SKILL.md. Lorem ipsum dolor sit
amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore
et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation.
`)

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, _, err := frontmatter.Parse(benchFrontmatter); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDep(b *testing.B) {
	cases := []string{"foo", "foo@1.0.0", "foo@>=1.0.0", "some-skill@2.1.0"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range cases {
			_, _ = frontmatter.ParseDep(c)
		}
	}
}
