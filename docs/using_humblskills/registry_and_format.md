# Registry & skill format

Skills in this registry follow the [agentskills.io](https://agentskills.io) format. Each skill is a directory with a `SKILL.md` (and optional supporting files) that agents can load as instructions.

## humblSKILLS frontmatter extensions

humblSKILLS-specific keys live under the optional **`metadata:`** map so the top level stays aligned with [agentskills.io](https://agentskills.io) (`name`, `description`, and other spec fields only).

| Key under `metadata:` | Purpose |
|-------------------------|---------|
| `version` | Skill package version (semver, `MAJOR.MINOR.PATCH`) |
| `author` | Who maintains the skill |
| `requires` | Dependencies or constraints (as defined by the skill) |
| `platforms` | Which agent platforms the skill targets |
| `category` | One coarse browsing bucket, from a closed set (see below) |
| `role` | Optional target role, from a closed set (see below) |
| `tags` | Freeform keywords for search (many per skill) |
| `previous_names` | Names this skill used to publish under, so `update` can follow a rename (see [Updating skills](updating.md)) |
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

Filter by it, or turn the grouping off if you prefer one flat list:

```sh
humblskills search --category=design
humblskills profile set group_by_category off    # on | off | default (on)
```

The skills browser and `install` / `list` pickers nest skills under collapsible
category headings by default; `group_by_category` is the toggle.

### Roles

`role` is optional and, like `category`, single-valued and validated against a
closed set. It scopes a skill to a job function rather than a subject area,
which is mainly useful for private, company-internal registries.

| Role | Meaning |
|------|---------|
| `fde` | Forward-deployed engineer |
| `ds` | Data scientist |
| `sdr` | Sales development representative |

```sh
humblskills search --role=fde
```

A skill with no `role` is unscoped and shows up regardless. No first-party
humblSKILLS skill sets a role today — the taxonomy exists so registries that
need it can use it. Adding a role, like adding a category, is a taxonomy change
that ships in a CLI release.

### Provenance for mirrored skills (`upstream:`)

Some skills are **distillations of an upstream source** rather than original
work — the `better-*` family, for example. Those declare a top-level `upstream:`
block naming where the source lives, which file in the skill preserves the copy
it was written against, and any deliberate divergences (`deltas:`).

Because the preserved copy *is* the baseline, drift is detectable as a plain
comparison — no hashes or lockfiles. That check is a maintainer tool and runs
against a source checkout of the repo, not installed copies:

```sh
humblskills mirrors check          # which mirrors drifted from upstream, and how badly
humblskills mirrors plan <skill>   # write a re-sync work order for a drifted mirror
```

`check` reports `current`, `drifted` (structure intact — review the diff),
`rewritten` (upstream changed enough to re-distil), or `unknown` (upstream
couldn't be fetched — never silently treated as healthy). `plan` writes the two
files to diff, the concepts affected, the declared deltas that must survive, and
a checklist. Detection is automated; the re-distillation is judgment and is
deliberately never automated.

## Where skills live in the repo

Published skills live under `skills/<skill-id>/` in the [humblSKILLS repository](https://github.com/jjfantini/humblSKILLS). The CLI reads the bundled registry and installs the matching directory to your configured location for Cursor, Claude Code, Codex, etc.

## How `registry.json` is built (contributors)

`make registry` (from the repo root) scans `skills/*/SKILL.md` and writes
`registry.json`. Each entry's download pin is the registry-level
`source.sha` — `humblskills install` fetches the skill tarball at **exactly
that commit**, then verifies the extracted directory matches the advertised
`dir_sha`.

Constraints that trip people up:

- The generator reads skill **content** from the working tree but stamps
  `source.sha` from **git HEAD**. The first local run after you add or edit a
  skill therefore records a SHA that cannot yet serve those new bytes. That is
  expected: commit the skill change, then let the Registry workflow (or a
  second `make registry` at the new HEAD) rewrite the pin.
- "The skill directory exists at that SHA" is not enough. An older revision of
  the same path still installs the wrong bytes and fails the `dir_sha` check.
  The self-heal compares **git tree object ids** per skill, so both missing and
  outdated content trigger a rewrite.
- Run `make registry-check` locally after skill edits. It fails on the same
  conditions a rebuild would fix; CI's Registry job auto-repairs on push rather
  than blocking the PR.

```sh
make registry        # regenerate registry.json
make registry-check  # fail if a rebuild would change anything (local gate)
```

## Local install layout

`humblskills install` writes one canonical skill directory, then exposes that
directory to each agent platform — by symlink for Claude Code, Cursor, and
Codex, and by zip for Claude Desktop. See
[Platforms & where skills land](platforms.md) for the full path tables, scopes,
and the Claude Desktop upload flow.

Codex officially supports symlinked skill folders in `.agents/skills`, so
humblSKILLS uses direct skill folders for local discovery. Codex plugins remain
out of scope for local installs; use plugins only when distributing reusable
skills with app or MCP integrations.

## Multiple registries

A registry is just a `registry.json` URL, and you can configure several at once
(a public one plus your company's private one). See
[Registries](registries.md).

## Related topics

- [Platforms & where skills land](platforms.md)
- [Registries](registries.md)
- [Updating skills](updating.md)
- [Preserving user content](preserving_user_content.md)
