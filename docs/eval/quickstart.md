# Eval quickstart

```sh
humblskills doctor                          # check runner availability
humblskills eval set-key anthropic          # store key in the OS keyring
humblskills eval runners                    # one-liner per-runner status
humblskills eval                            # dashboard entry → Eval Home TUI
humblskills eval run use-smart-skill        # non-TUI run
humblskills eval showcase                   # canonical demo
humblskills eval ls                         # iterations per skill
humblskills eval prune use-smart-skill --keep-last 5
```

## Secrets

Secrets **do not** land in profile JSON. `eval set-key` resolves credentials in this order:

1. Environment variables
2. OS keyring
3. `$XDG_CONFIG_HOME/humblskills/secrets.json` (mode `0600`)

The TUI can prompt with masked input when appropriate.

## See also

- [Artifacts](artifacts.md) - what is written under `evals/<skill>/iteration-N/`
- [Scenarios](scenarios.md) - `evals/scenarios.json` and assertions
