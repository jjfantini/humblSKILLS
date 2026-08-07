# humblSKILLS

humblSKILLS is a **skill registry** plus a single-binary Go CLI, **`humblskills`**, that installs [agentskills.io](https://agentskills.io)-format skills for the agent stack you already use — Claude Code, Cursor, Codex, and Claude Desktop.

A **skill** is a folder with a `SKILL.md` that teaches your coding agent how to do one thing well. humblskills fetches those folders and puts them where each agent looks for them, so you don't have to know where that is.

## Three commands to try

```sh
brew install jjfantini/humbl/humblskills   # or see Installation for other options
humblskills search                         # browse what's available
humblskills install smart-commit           # install one
```

No account, no server, no telemetry. [Full installation options →](getting_started/installation.md)

## What you get

1. **Skill registry** - A monorepo of skills in the agentskills.io shape, with humblSKILLS extensions under `metadata:` in `SKILL.md` (`version`, `category`, `role`, `requires`, `platforms`, `tags`, `previous_names`, `preserve`).
2. **`humblskills` CLI** - Pulls a skill directory from a registry and installs it in the right place for your platform, keeps it updated, and preserves the content you accumulate inside it.

## Where to go next

| If you want to… | Read |
|-----------------|------|
| Install the CLI | [Installation](getting_started/installation.md) - agent prompt, Homebrew, shell, Go, releases |
| Learn the everyday commands | [Quickstart](getting_started/quickstart.md) - `doctor`, `search`, `install`, `list`, `update` |
| See what's installable | [Skill catalog](skills.md) - every skill in the public registry |
| Understand self-learning skills | [Smart skills & the brain](smart_skills.md) - the CCCCC ontology and the brain protocol |
| Know where skills land on disk | [Platforms](using_humblskills/platforms.md) - paths, scopes, Claude Desktop zips |
| Use a private or second registry | [Registries](using_humblskills/registries.md) - tokens, multiple registries, `--from` |
| Keep skills current | [Updating skills](using_humblskills/updating.md) - `update` vs `upgrade`, renames |
| Keep your own notes inside a skill | [Preserving user content](using_humblskills/preserving_user_content.md) |
| Share a skill set with a team | [Sharing skillsets](using_humblskills/sharing_skillsets.md) - `init`, `export`, `sync` |
| Author or benchmark a skill | [Registry & skill format](using_humblskills/registry_and_format.md) · [Eval](eval/index.md) |

Documentation is hosted on [GitHub Pages](https://jjfantini.github.io/humblSKILLS/); the live site is **published from `main`** so it stays aligned with released, installable `humblskills` builds. Source lives in the [humblSKILLS repo](https://github.com/jjfantini/humblSKILLS).
