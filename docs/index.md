# humblSKILLS

humblSKILLS is a **skill registry** plus a single-binary Go CLI, **`humblskills`**, that installs [agentskills.io](https://agentskills.io)-format skills for the agent stack you already use (Claude Code, Cursor, Codex, and similar).

## What you get

1. **Skill registry** - A monorepo of skills in the agentskills.io shape, with humblSKILLS extensions under `metadata:` in `SKILL.md` (`requires`, `platforms`, `tags`, `preserve`, `version`, and similar).
2. **`humblskills` CLI** - Pulls a skill directory from the registry and installs it in the right place for your platform. No hosted account, no telemetry.

## Next steps

- [Installation](getting_started/installation.md) - **Recommended:** agent prompt + raw `SKILL.md`; Homebrew, shell, Go, releases
- [Quickstart](getting_started/quickstart.md) - dashboard (`start`), `doctor`, `search`, `install`, `update`
- [Eval](eval/index.md) - benchmark skills and inspect compound smart-skill runs
- [Preserving user content](using_humblskills/preserving_user_content.md) - `metadata.preserve` in `SKILL.md`

Documentation is hosted on [GitHub Pages](https://jjfantini.github.io/humblSKILLS/); the live site is **published from `main`** so it stays aligned with released, installable `humblskills` builds. Source lives in the [humblSKILLS repo](https://github.com/jjfantini/humblSKILLS).
