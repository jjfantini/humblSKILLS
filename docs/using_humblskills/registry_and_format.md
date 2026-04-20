# Registry & skill format

Skills in this registry follow the [agentskills.io](https://agentskills.io) format. Each skill is a directory with a `SKILL.md` (and optional supporting files) that agents can load as instructions.

## humblSKILLS frontmatter extensions

Authors may include extra keys in YAML frontmatter:

| Key | Purpose |
|-----|---------|
| `requires` | Dependencies or constraints (as defined by the skill) |
| `platforms` | Which agent platforms the skill targets |
| `tags` | Discovery / grouping |
| `preserve` | Paths to keep on `update` when replacing an installed skill (see [Preserving user content](preserving_user_content.md)) |

Other frontmatter fields (for example `name`, `description`, `version`) follow the normal agentskills.io expectations.

## Where skills live in the repo

Published skills live under `skills/<skill-id>/` in the [humblSKILLS repository](https://github.com/jjfantini/humblSKILLS). The CLI reads the bundled registry and installs the matching directory to your configured location for Cursor, Claude Code, Codex, etc.
