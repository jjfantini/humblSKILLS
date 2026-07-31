# Skill catalog

Everything the public humblSKILLS registry currently ships. Install any of them
with:

```sh
humblskills install <skill-name>
```

Or browse them interactively — `humblskills search` with no query opens a
fuzzy-searchable browser, grouped by the categories below.

This page is generated from [`registry.json`](https://github.com/jjfantini/humblSKILLS/blob/main/registry.json)
at docs build time, so it always matches the released registry.

<!-- SKILL_CATALOG -->

## Legend

- **Version** - the skill package version, independent of the CLI version.
  `humblskills update` compares your installed copy against this.
- **↗ mirrored** - the skill is a distillation of an upstream source, with the
  original preserved verbatim inside the skill. See
  [provenance for mirrored skills](using_humblskills/registry_and_format.md#provenance-for-mirrored-skills-upstream).

Every skill here targets **Claude Code, Cursor, and Codex**, and can also be
exported as a [Claude Desktop zip](using_humblskills/platforms.md#claude-desktop-and-claudeai-zip-upload).

## Adding your own

Skills are directories with a `SKILL.md`. To author one, install the
`smart-skill` scaffolder:

```sh
humblskills install smart-skill
```

Then see [Registry & skill format](using_humblskills/registry_and_format.md) for
the frontmatter humblSKILLS understands.
