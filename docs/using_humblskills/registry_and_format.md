# Registry & skill format

Skills in this registry follow the [agentskills.io](https://agentskills.io) format. Each skill is a directory with a `SKILL.md` (and optional supporting files) that agents can load as instructions.

## humblSKILLS frontmatter extensions

humblSKILLS-specific keys live under the optional **`metadata:`** map so the top level stays aligned with [agentskills.io](https://agentskills.io) (`name`, `description`, and other spec fields only).

| Key under `metadata:` | Purpose |
|-------------------------|---------|
| `requires` | Dependencies or constraints (as defined by the skill) |
| `platforms` | Which agent platforms the skill targets |
| `category` | One coarse browsing bucket, from a closed set (see below) |
| `tags` | Freeform keywords for search (many per skill) |
| `version` | Skill package version (semver) |
| `preserve` | Paths to keep on `update` when replacing an installed skill (see [Preserving user content](preserving_user_content.md)) |

Other top-level frontmatter follows the normal agentskills.io expectations.

### Categories

`category` is required and validated at registry-build time against a small,
stable list, unlike `tags`, which is freeform. It exists to give every skill
exactly one home for browsing and filtering (`humblskills search --category=`),
rather than relying on inconsistent tag conventions across skill authors.

| Category | Use for |
|----------|---------|
| `development` | Git/workflow tooling, integrations, interview/system-design skills |
| `design` | Frontend, UI/UX, and creative/media generation skills |
| `writing` | Content and copy editing |
| `meta` | Skill authoring, project onboarding, and other humblSKILLS-about-humblSKILLS skills |

Adding a new category is a taxonomy decision (edit `frontmatter.Categories` in
`cli/internal/frontmatter/validate.go`), not something an individual skill
author should do by picking a new value.

## Where skills live in the repo

Published skills live under `skills/<skill-id>/` in the [humblSKILLS repository](https://github.com/jjfantini/humblSKILLS). The CLI reads the bundled registry and installs the matching directory to your configured location for Cursor, Claude Code, Codex, etc.

## Local install layout

`humblskills install` writes one canonical skill directory, then exposes that
directory to agent platforms with symlinks.

| Mode | Canonical directory |
|------|---------------------|
| `--global` | `~/.humblskills/skills/<skill-id>` |
| User scope | `$XDG_DATA_HOME/humblskills/skills/<skill-id>` |
| Project scope | `<repo>/.humblskills/skills/<skill-id>` |

Platform targets are symlinks:

| Platform | User target |
|----------|-------------|
| Claude Code | `~/.claude/skills/<skill-id>` |
| Cursor | `~/.cursor/skills/<skill-id>` |
| Codex | `$HOME/.agents/skills/<skill-id>` |

Codex officially supports symlinked skill folders in `.agents/skills`, so
humblSKILLS uses direct skill folders for local discovery. Codex plugins remain
out of scope for local installs; use plugins only when distributing reusable
skills with app or MCP integrations.

## Migrating existing Claude Code installs

Run:

```sh
humblskills migrate claude-code --global --yes
```

The migration scans `~/.claude/skills`, reads each `SKILL.md`, and matches the
skill `name` against the humblSKILLS registry. Registry-known skills are copied
into `~/.humblskills/skills`, their preserved local files are retained, and the
Claude Code directory is replaced with a symlink. Unregistered personal skills
are reported and skipped.
