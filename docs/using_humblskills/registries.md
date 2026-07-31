# Registries

A **registry** is just a JSON file (`registry.json`) that lists skills and where
to download each one. The CLI reads a registry, then installs the skill
directory it points at.

You do **not** have to set any of this up to get started. humblskills ships with
the public humblSKILLS registry built in, so `humblskills search` and
`humblskills install <skill>` work on a fresh install with zero configuration.

Read this page when you want to:

- pull skills from a **private** repo (your company's own skill library), or
- use **several registries at once** (public + private), or
- share a registry setup with teammates.

## The mental model

| Thing | What it is |
|-------|------------|
| **Registry** | A `registry.json` URL (or local file) listing available skills |
| **Named registry** | A registry you've given a short name, e.g. `public`, `work` |
| **Token** | A GitHub personal access token, needed only for **private** registries |
| **Default registry** | The public humblSKILLS one, used when you've configured nothing |

Tokens are stored in your **OS keychain** (macOS Keychain, Windows Credential
Manager, Linux Secret Service), not in a config file. They never appear in
`humblskills profile show`.

## Add a second registry

`registry add` takes a name and a location. The location can be a full URL, or
the `owner/repo` shorthand — the CLI expands that into the raw `registry.json`
URL for you.

```sh
humblskills registry add public     jjfantini/humblSKILLS
humblskills registry add work       my-company/our-skills
humblskills registry add work       my-company/our-skills@develop   # pick a branch
humblskills registry add                                            # no args → prompts you
```

Run `humblskills registry add` with no arguments and it asks for the name, the
URL, and (optionally) a token — handy if you'd rather not remember the flags.

Once more than one registry is configured, `search` and the skill browser show
them **together, grouped by registry**, with each skill tagged by where it came
from:

```text
── work ──
  internal-deploy-runbook   …
── public ──
  smart-commit              …
```

## Private registries: add a token

If the registry lives in a private GitHub repo, the CLI needs a token to read
both `registry.json` and each skill download. Use a GitHub personal access token
with **read access** to that repo.

```sh
humblskills registry login --name work
```

You'll get a masked prompt. The token is verified against the registry before it
is saved, so a typo fails immediately instead of at install time. Then:

```sh
humblskills search                  # includes the private registry
humblskills install <skill>         # token is applied automatically
humblskills registry logout --name work   # remove the stored token
```

Other ways to supply the token — useful in CI, where there's no prompt:

```sh
echo "$GITHUB_TOKEN" | humblskills registry login --name work   # piped stdin
humblskills registry login --name work --token "$GITHUB_TOKEN"  # flag
```

!!! note "Where the token is looked up"
    Highest precedence first: `--token` → `HUMBLSKILLS_TOKEN` → OS keychain →
    a `0600` fallback file (used only when no keychain is available). Tokens are
    ignored for `file://` registries and for the public default.

## Installing when a name exists in two registries

Skill names are unique **within** a registry, not across them. When two
registries publish the same name, disambiguate with `--from`:

```sh
humblskills install smart-commit --from public
```

When only one registry has the name, no flag is needed. When several do and you
omit `--from`, the CLI asks you to pick on a normal terminal, and fails with the
list of candidates in a script or CI so nothing ambiguous gets installed
silently.

## Managing registries

```sh
humblskills registry list                     # names, URLs, and token state
humblskills registry rename work acme         # rename (moves its stored token too)
humblskills registry add work <new-url>       # re-add an existing name to change its URL
humblskills registry remove work              # drop it (and its stored token)
humblskills registry refresh                  # force-refresh the cached registry
```

Or run bare **`humblskills registry`** (also the **registry** tile on the
dashboard) to open the interactive **registry manager** — add, rename, login,
logout, remove, and refresh from one screen. On a non-interactive terminal it
falls back to printing `registry list`.

## Registries show up everywhere

Once configured, registry provenance flows through the rest of the CLI:

| Command | What it shows |
|---------|---------------|
| `humblskills list` | Tags each installed skill with the registry it came from |
| `humblskills update` | Checks each skill against **its own** registry |
| `humblskills doctor` | Per-registry reachability, token state, and skill count |
| `humblskills profile show` | A **Registries** section listing all of them |

## Single-registry setup (no names)

If you only ever use one registry, you can skip named registries entirely and
just point the CLI at it:

```sh
humblskills profile set registry https://raw.githubusercontent.com/<owner>/<repo>/main/registry.json
humblskills registry login       # token for that one registry
humblskills search
```

Or set it per-command / per-shell:

```sh
export HUMBLSKILLS_REGISTRY=https://raw.githubusercontent.com/<owner>/<repo>/main/registry.json
export HUMBLSKILLS_TOKEN=<github token with read access>
humblskills search
```

Registry resolution order, highest precedence first: `--registry` →
`HUMBLSKILLS_REGISTRY` → named registries / `profile set registry` → the hosted
public default.

## Tab completion

Registry names, skill names, and `--from` / `--name` values all tab-complete
once you install the completion script:

```sh
humblskills completion zsh --help     # also: bash, fish, powershell
```

## Related topics

- [Sharing skillsets](sharing_skillsets.md) - ship a registry list to teammates so `sync` configures it for them
- [Registry & skill format](registry_and_format.md) - what a skill looks like inside a registry
- [Quickstart](../getting_started/quickstart.md)
