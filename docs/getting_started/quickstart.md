# Quickstart

```sh
humblskills doctor                    # verify the environment
humblskills search                    # browse the registry
humblskills install use-smart-skill
humblskills list
humblskills update                    # pick which drifted skills to upgrade
humblskills update --all --yes        # non-interactive bulk upgrade
humblskills uninstall use-smart-skill
```

## Machine-friendly output

Every command accepts:

- **`--json`** - machine-readable output
- **`--yes`** - skip confirmation prompts

Use these in scripts and CI.

## Related topics

- [Registry & skill format](../using_humblskills/registry_and_format.md)
- [Preserving user content](../using_humblskills/preserving_user_content.md)
- [Eval quickstart](../eval/quickstart.md)
