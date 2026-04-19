---
name: skill-example-with-dep
description: Seed skill that depends on skill-example-hello - exercises unpinned dependency resolution.
metadata:
  version: 0.1.0
  requires:
    - skill-example-hello
  platforms: [claude-code, cursor]
  tags: [example, seed]
---

# skill-example-with-dep

Placeholder skill used to verify that the registry parses an unpinned
`requires:` entry and that topological install order is respected.
