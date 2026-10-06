# Contributing

Contributions land as pull requests into `develop`. The maintainer reviews and
merges them with a merge commit; nobody merges their own promotion.

- **A new skill** - use the promotion path below. It is built for FDEs who
  create skills with their agent and want them in a registry.
- **A change to a published skill, or to the CLI** - open a normal PR; see
  [Changing what's already here](#changing-whats-already-here).

## Adding a new skill

Skills are created and promoted with the
[`smart-skill`](skills/smart-skill/SKILL.md) skill, which you probably already
use. Install it once:

```sh
humblskills install smart-skill
```

### 1. Create it

Ask your agent to "create a smart skill for ...". smart-skill scaffolds the
skill (router `SKILL.md`, brain, `scripts/lint.sh`) and fills it in with you.
Use it for real for a while: a skill earns promotion by working.

### 2. Pick a registry

When the skill works, the agent asks where it goes and recommends one:

| | humblSKILLS | happySKILLS |
|---|---|---|
| For | anyone who uses humblskills | HappyRobot only |
| Repo | `jjfantini/humblSKILLS` (public) | happySKILLS (private) |
| Pick it when | an engineer outside HappyRobot gets value from it as-is | it names HappyRobot, customers, internal systems, or internal processes |

When unsure, pick happySKILLS: a public branch is public the moment it is
pushed, while going public later is just another promotion. promote.sh finds
the happySKILLS repo in your `humblskills registry` config.

### 3. Settle its category, role, and tags

These decide where the CLI files the skill (**Registry › Category › Role**
in the browser) and what `humblskills search --category=... --role=...` finds.

| Field | Values |
|---|---|
| `category` | `development`, `design`, `writing`, `meta` |
| `role` | `fde`, `ds`, `sdr` - or `none` for a general-purpose public skill (happySKILLS needs one) |
| `tags` | 3+ kebab-case words people would search for |

The agent proposes them and you confirm. `promote.sh` writes them into your
`SKILL.md`, so you only pass them once.

### 4. Run the gate

```sh
PROMOTE=~/.humblskills/skills/smart-skill/scripts/promote.sh

bash "$PROMOTE" my-skill --to humblskills --check \
  --category development --role none --tags git,code-review,pull-requests
```

`my-skill` is looked up in `~/.humblskills/skills` first, then each agent's
skills directory; a path works too. Fix every `FAIL` and re-run. `WARN`s
don't block; they are copied into the PR so the reviewer sees them.

### 5. Promote

```sh
bash "$PROMOTE" my-skill --to humblskills            # or --to happyskills
bash "$PROMOTE" my-skill --to humblskills --dry-run  # preview the commit first
```

It clones the target (or forks it if you can't push), cuts `feat/add-my-skill`
from `develop`, stages a clean copy of the skill, lints it, validates it with
the real `build-registry` when Go 1.23+ is installed, commits
`feat(skills): add my-skill skill` as you, pushes, and opens the PR. Your own
clones and working trees are never touched. Prerequisites: git, python3, and
the GitHub CLI (`brew install gh && gh auth login`).

### 6. Review, then install

Review happens on the PR. Fix feedback in your local copy and run the same
command again: it pushes a follow-up commit and comments on the PR. After the
merge, the skill ships with the next pre-release and reaches the public
registry when `develop` is promoted to `main`; from then on
`humblskills install my-skill` takes over your local copy and keeps its
preserved brain files, so `humblskills update` manages it.

### The bar

promote.sh refuses to open a PR while any of these fail:

- **Frontmatter**: `name` (kebab-case, at most 64 chars, matches its
  directory), `description` (at most 1024 chars, no `<` or `>`, with a "Use
  when ..." trigger), `license`, `metadata.author`, semver `metadata.version`,
  values the CLI's YAML parser accepts.
- **Taxonomy**: category, role, and tags as above; no scaffold `TODO`s.
- **A real smart skill**: `## Brain Protocol`, the brain files,
  `references/wiki/` and `references/raw/`, its own `scripts/lint.sh` passing,
  and every brain path in `metadata.preserve` so updates never wipe user data.
- **Safe to publish**: no credentials or `.env` files, no symlinks or
  `node_modules/`, no file over 5 MiB (every install downloads the whole repo),
  no files the target's `.gitignore` would silently drop, and no HappyRobot
  mentions in a public promotion.
- **Net-new**: the target doesn't already publish the name or list it as a
  former name.

Each failure says how to fix it. The full table, with the reason behind each
check, is in smart-skill's
[quality-gate concept](skills/smart-skill/references/wiki/smart/promote/quality-gate.md).

### What promotion doesn't do

- **Change an existing skill.** Merging learnings or behavior back into a
  published skill is a separate path that isn't built yet; open a normal PR.
- **Touch `registry.json`.** The Registry workflow regenerates it after the
  merge, so promotion PRs never conflict on it.
- **Merge.** The maintainer reviews and merges.

## Changing what's already here

1. Branch off `develop`; open the PR into `develop`.
2. Write [Conventional Commits](https://www.conventionalcommits.org) for every
   commit (`feat(scope): ...`, `fix(scope): ...`). release-please builds
   releases from them; see [AGENTS.md](AGENTS.md) for the rules and why.
3. Before pushing: `make test` and `make vet`; after editing `skills/`,
   `make registry`; after editing smart-skill's scripts,
   `bash skills/smart-skill/tests/run.sh`.
4. PRs merge with a merge commit (`gh pr merge --merge`), never squash.
