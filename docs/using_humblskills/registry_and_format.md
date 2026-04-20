# Registry & skill format

Skills in this registry follow the [agentskills.io](https://agentskills.io) format. Each skill is a directory with a `SKILL.md` (and optional supporting files) that agents can load as instructions.

## humblSKILLS frontmatter extensions

humblSKILLS-specific keys live under the optional **`metadata:`** map so the top level stays aligned with [agentskills.io](https://agentskills.io) (`name`, `description`, and other spec fields only).

| Key under `metadata:` | Purpose |
|-------------------------|---------|
| `requires` | Dependencies or constraints (as defined by the skill) |
| `platforms` | Which agent platforms the skill targets |
| `tags` | Discovery / grouping |
| `version` | Skill package version (semver) |
| `preserve` | Paths to keep on `update` when replacing an installed skill (see [Preserving user content](preserving_user_content.md)) |

Other top-level frontmatter follows the normal agentskills.io expectations.

## Where skills live in the repo

Published skills live under `skills/<skill-id>/` in the [humblSKILLS repository](https://github.com/jjfantini/humblSKILLS). The CLI reads the bundled registry and installs the matching directory to your configured location for Cursor, Claude Code, Codex, etc.
