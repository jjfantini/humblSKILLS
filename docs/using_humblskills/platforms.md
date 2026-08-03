# Platforms & where skills land

`humblskills install` writes **one** copy of a skill into a canonical store,
then makes it visible to each agent platform. Understanding that split answers
most "where did my skill go?" questions.

## Supported platforms

| Platform | How the skill is delivered | User target |
|----------|---------------------------|-------------|
| `claude-code` | Symlink | `~/.claude/skills/<skill-id>` |
| `cursor` | Symlink | `~/.cursor/skills/<skill-id>` |
| `codex` | Symlink | `~/.agents/skills/<skill-id>` |
| `claude-desktop` | **Zip you upload** (see below) | `~/.humblskills/desktop/<skill-id>.zip` |

By default the CLI installs to every platform it **detects** on your machine
(it looks for `~/.claude`, `~/.cursor`, `~/.codex` / `~/.agents`, and Claude
Desktop's data directory). Restrict that with `--platform`:

```sh
humblskills install smart-commit --platform claude-code
humblskills install smart-commit --platform claude-code,cursor
```

Set a default once so you don't repeat the flag:

```sh
humblskills profile set platforms claude-code,cursor
```

Check what was detected and whether each target is writable:

```sh
humblskills doctor
```

## The canonical store (and scopes)

The real skill directory lives in the store; platform locations are symlinks
into it. That means one copy on disk, one place to update.

| Scope | Canonical directory |
|-------|---------------------|
| `global` (the default) | `~/.humblskills/skills/<skill-id>` |
| `user` | `$XDG_DATA_HOME/humblskills/skills/<skill-id>` |
| `project` | `<current repo>/.humblskills/skills/<skill-id>` |

```sh
humblskills install smart-commit                    # global (default)
humblskills install smart-commit --scope project    # this repo only
humblskills install smart-commit --global           # alias for --scope global
humblskills profile set scope project               # change your default
```

Use `project` scope for skills a specific repository needs, so they're not
installed machine-wide.

## Claude Desktop and claude.ai (zip upload)

Claude Desktop and claude.ai **cannot read skills from your filesystem** — they
take an account-level zip upload. So the `claude-desktop` platform doesn't
symlink; it writes an upload-ready zip (the skill folder at the zip root).

Select it like any other platform and `install` / `update` keep the zips
current for you:

```sh
humblskills install smart-commit --platform claude-desktop
# → ~/.humblskills/desktop/smart-commit.zip
```

Then upload the zip: **claude.ai (or the Claude desktop app) → Settings →
Capabilities → Skills → upload the zip**.

Need zips without reinstalling anything?

```sh
humblskills export desktop                       # every installed skill
humblskills export desktop smart-commit          # just one
humblskills export desktop -o ./dist             # write somewhere else
```

!!! warning "Uploads don't auto-update"
    A zip you've uploaded is a copy inside your Claude account. `humblskills
    update` refreshes the local zip, but you have to **re-upload** it for Claude
    to see the new version. The filesystem platforms don't have this problem.

## Adopting skills you already installed by hand

If you have skills sitting in `~/.claude/skills` that you installed manually,
`migrate` brings the registry-known ones under humblskills management:

```sh
humblskills migrate claude-code --global --yes
```

It reads each `SKILL.md`, matches the `name` against the registry, copies
matches into the canonical store (keeping their preserved local files), and
replaces the Claude Code directory with a symlink. Your own personal skills that
aren't in any registry are reported and **skipped** — nothing of yours is
deleted.

## Related topics

- [Registry & skill format](registry_and_format.md)
- [Updating skills](updating.md)
- [Preserving user content](preserving_user_content.md)
