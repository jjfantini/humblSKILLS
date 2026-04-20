# humblSKILLS

humblSKILLS is a **skill registry** plus a single-binary Go CLI, **`humblskills`**, that installs [agentskills.io](https://agentskills.io)-format skills for the agent stack you already use (Claude Code, Cursor, Codex, and similar).

## What you get

1. **Skill registry** - A monorepo of skills in the agentskills.io shape, with small humblSKILLS extensions in `SKILL.md` frontmatter: `requires`, `platforms`, `tags`, `preserve`.
2. **`humblskills` CLI** - Pulls a skill directory from the registry and installs it in the right place for your platform. No hosted account, no telemetry.

## Next steps

- [Installation](getting_started/installation.md) - shell, Go, releases, Homebrew
- [Quickstart](getting_started/quickstart.md) - `doctor`, `search`, `install`, `update`
- [Eval](eval/index.md) - benchmark skills and inspect compound smart-skill runs
- [Preserving user content](using_humblskills/preserving_user_content.md) - `preserve:` in `SKILL.md`

Documentation is hosted on [GitHub Pages](https://jjfantini.github.io/humblSKILLS/); the live site is **published from `main`** so it stays aligned with released, installable `humblskills` builds. Source lives in the [humblSKILLS repo](https://github.com/jjfantini/humblSKILLS).
