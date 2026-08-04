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
| **Default registry** | The public humblSKILLS one, used when you've configured **no named registries** |

Tokens are stored in your **OS keychain** (macOS Keychain, Windows Credential
Manager, Linux Secret Service), not in a config file. They never appear in
`humblskills profile show`.

!!! warning "Naming any registry replaces the default — it is not a fallback"
    The CLI runs in one of two modes, and there is **no fallback between them**:

    - **No named registries** → single-registry mode, using
      `--registry` → `HUMBLSKILLS_REGISTRY` → `profile set registry` → the hosted public default.
    - **One or more named registries** → the CLI uses **exactly that set** and nothing else.
      The hosted public default is no longer consulted, and `--registry` /
      `HUMBLSKILLS_REGISTRY` are **ignored**.

    So `humblskills registry add work my-company/our-skills` on its own **silently hides every
    public skill** — `search` stops listing them and `install` reports them as not found. If you
    want both, add both:

    ```sh
    humblskills registry add public jjfantini/humblSKILLS
    humblskills registry add work   my-company/our-skills
    ```

    Check what you actually have with `humblskills registry list`.

## Add a second registry

`registry add` takes a name and a location. The location can be a full URL, or
the `owner/repo` shorthand — the CLI expands that into the raw `registry.json`
URL for you.

```sh
humblskills registry add public     jjfantini/humblSKILLS               # add this too — see the warning above
humblskills registry add work       my-company/our-skills
humblskills registry add work       my-company/our-skills@develop   # pick a branch
humblskills registry add                                            # no args → prompts you
```

**Add `public` explicitly** the first time you name a registry. As soon as one named
registry exists, the built-in public default stops being consulted — naming only your
private registry is what makes public skills vanish.

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

If you only ever use one registry — and don't want the public one — you can skip
named registries entirely and just point the CLI at it:

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

!!! note "How a registry is resolved"
    Mode is decided first, by whether **any named registry** exists in your profile:

    | Your profile | What the CLI uses |
    |---|---|
    | No named registries | One registry: `--registry` → `HUMBLSKILLS_REGISTRY` → `profile set registry` → hosted public default |
    | One or more named registries | **Exactly** those, for every command. `--registry` and `HUMBLSKILLS_REGISTRY` are ignored; the hosted default is not added |

    This means `--registry` is **not** a global override — it only takes effect in
    single-registry mode. To temporarily query somewhere else while you have named
    registries configured, use `--from <name>` to pick among them, or
    `humblskills registry remove` / a separate `--profile <path>` for a throwaway config.

Token lookup is independent of that, highest precedence first: `--token` →
`HUMBLSKILLS_TOKEN` → OS keychain → `0600` fallback file.

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
