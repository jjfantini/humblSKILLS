---
name: skill-example-pinned-dep
description: Seed skill with a minimum-version dependency - exercises @>= constraint parsing.
metadata:
  version: 0.1.0
  requires:
    - skill-example-hello@>=0.1.0
  platforms: [claude-code]
  tags: [example, seed]
---

# skill-example-pinned-dep

Placeholder skill used to verify that the registry parses a minimum-version
dependency constraint (`@>=0.1.0`) and that it's satisfied by the registered
version of the dependency.
